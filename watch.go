package main

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// refreshInterval is how often the fzf list is reloaded while the picker is open.
const refreshInterval = 2 * time.Second

// buildFzfArgs returns the fzf command-line arguments for the agent picker.
// port is the fzf --listen port used to send reload commands.
func buildFzfArgs(port int) []string {
	// The binary path: prefer the running binary so aw preview works even if
	// aw is not on PATH yet.
	self, err := os.Executable()
	if err != nil {
		self = "aw"
	}

	previewCmd := fmt.Sprintf("%s preview {-1}", self)
	listenAddr := fmt.Sprintf("localhost:%d", port)

	backCmd := fmt.Sprintf("execute(%s pick -)+abort", self)
	respondCmd := fmt.Sprintf("execute-silent(%s respond {-1})+reload(%s list)", self, self)

	return []string{
		"--ansi",
		"--listen", listenAddr,
		"--delimiter", "\t",
		"--with-nth", "1..4", // show all fields except last (target)
		"--preview", previewCmd,
		"--preview-window", "right:55%:wrap:follow",
		"--bind", fmt.Sprintf("ctrl-r:reload(%s list)", self),
		"--bind", "esc:abort",
		"--bind", fmt.Sprintf("ctrl-b:%s", backCmd),
		"--bind", fmt.Sprintf("ctrl-y:%s", respondCmd),
		"--header", "Agent Watcher │ Enter: switch  Ctrl-r: refresh  Ctrl-b: back  Ctrl-y: answer  Esc: quit",
		"--no-sort", // preserve our priority ordering
		"--layout", "reverse",
	}
}

// parseSelectedTarget extracts the tmux target from a selected fzf row.
// The target is always the last tab-delimited field.
func parseSelectedTarget(row string) string {
	row = strings.TrimSpace(row)
	if row == "" {
		return ""
	}
	fields := strings.Split(row, "\t")
	return strings.TrimSpace(fields[len(fields)-1])
}

// sendReload posts a reload command to fzf's --listen HTTP endpoint.
func sendReload(port int, listCmd string) {
	body := fmt.Sprintf("reload(%s)", listCmd)
	addr := fmt.Sprintf("http://localhost:%d", port)
	_, _ = http.Post(addr, "text/plain", bytes.NewBufferString(body)) //nolint:noctx
}

// cmdWatch is the default command: launches the fzf agent picker with live refresh.
func cmdWatch(_ []string) error {
	self, err := os.Executable()
	if err != nil {
		self = "aw"
	}

	// Pick a random high port to avoid collisions with other aw instances.
	port := 40000 + rand.IntN(10000) //nolint:gosec

	// Build initial list.
	entries, err := collectEntries()
	if err != nil {
		return fmt.Errorf("collect entries: %w", err)
	}
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "aw: no agent windows detected")
	}

	// Convert entries to newline-delimited input for fzf.
	var input strings.Builder
	for _, e := range entries {
		input.WriteString(e.fzfRow())
		input.WriteByte('\n')
	}

	fzfArgs := buildFzfArgs(port)
	cmd := exec.Command("fzf", fzfArgs...)
	cmd.Stdin = strings.NewReader(input.String())
	cmd.Stderr = os.Stderr

	// Capture fzf output (the selected row).
	var out bytes.Buffer
	cmd.Stdout = &out

	// Start fzf.
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start fzf: %w", err)
	}

	// Background goroutine: periodically reload fzf list.
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Give fzf a moment to start listening.
		time.Sleep(300 * time.Millisecond)
		listCmd := fmt.Sprintf("%s list", self)
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sendReload(port, listCmd)
			case <-done:
				return
			}
		}
	}()

	// Wait for fzf to exit.
	waitErr := cmd.Wait()
	done <- struct{}{} // stop the refresh goroutine

	// fzf exits with code 1 when user presses Esc/q (no selection) — that's fine.
	selected := parseSelectedTarget(out.String())
	if selected == "" {
		return nil
	}

	// Record current window as "previous" so 'aw pick -' can return here.
	if cur, err := activeWindow(); err == nil && cur != "" {
		_ = writeLastWindow(stateDir(), cur)
	}

	// Switch to the selected tmux window.
	if err := switchToWindow(selected); err != nil {
		return fmt.Errorf("switch to %q: %w", selected, err)
	}

	_ = waitErr // fzf exit code is not an error we surface
	return nil
}
