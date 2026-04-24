package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// TmuxWindow represents a tmux window with its top pane.
type TmuxWindow struct {
	Session     string
	WindowIndex string
	WindowName  string
	Active      bool
	PaneID      string
	PanePID     int
}

// Target returns the tmux target string for this window (e.g. "myapp:2").
func (w TmuxWindow) Target() string {
	return w.Session + ":" + w.WindowIndex
}

// listTmuxWindows runs tmux to list all panes across all sessions.
// Using list-panes -a (rather than list-windows -a) ensures we see every pane
// in split windows, not just the currently active one. Callers deduplicate by
// window after determining which pane holds an agent.
func listTmuxWindows() ([]TmuxWindow, error) {
	out, err := exec.Command("tmux", "list-panes", "-a", "-F",
		"#{session_name}\t#{window_index}\t#{window_name}\t#{window_active}\t#{pane_id}\t#{pane_pid}",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("tmux list-panes: %w", err)
	}
	return parseTmuxWindows(string(out)), nil
}

// parseTmuxWindows parses the tab-delimited output of tmux list-windows.
func parseTmuxWindows(raw string) []TmuxWindow {
	var wins []TmuxWindow
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 6 {
			continue
		}
		panePID, err := strconv.Atoi(fields[5])
		if err != nil {
			continue
		}
		wins = append(wins, TmuxWindow{
			Session:     fields[0],
			WindowIndex: fields[1],
			WindowName:  fields[2],
			Active:      fields[3] == "1",
			PaneID:      fields[4],
			PanePID:     panePID,
		})
	}
	return wins
}

// activeWindow returns the current active tmux window as "session:window_index".
func activeWindow() (string, error) {
	out, err := exec.Command("tmux", "display-message", "-p", "#{session_name}:#{window_index}").Output()
	if err != nil {
		return "", fmt.Errorf("tmux display-message: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// switchToWindow runs tmux switch-client to focus the given target.
func switchToWindow(target string) error {
	return exec.Command("tmux", "switch-client", "-t", target).Run()
}

// capturePane returns the last n lines of pane content for the given target.
func capturePane(target string, lines int) (string, error) {
	out, err := exec.Command("tmux", "capture-pane", "-p", "-t", target,
		"-S", fmt.Sprintf("-%d", lines),
	).Output()
	if err != nil {
		return "", fmt.Errorf("tmux capture-pane: %w", err)
	}
	return string(out), nil
}
