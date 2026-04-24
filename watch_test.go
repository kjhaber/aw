package main

import (
	"strings"
	"testing"
)

func TestFzfArgs(t *testing.T) {
	args := buildFzfArgs(9999)

	joined := strings.Join(args, " ")

	// Must pass --listen with our port
	if !strings.Contains(joined, "--listen") || !strings.Contains(joined, "9999") {
		t.Errorf("fzf args missing --listen 9999: %v", args)
	}
	// Must pass --ansi for colour support
	if !strings.Contains(joined, "--ansi") {
		t.Errorf("fzf args missing --ansi: %v", args)
	}
	// Preview must use {-2} (pane ID) so it targets the agent pane directly,
	// not the currently active pane in the window.
	if !strings.Contains(joined, "preview {-2}") {
		t.Errorf("fzf args preview should use {-2} (pane ID), got: %v", args)
	}
	// Respond/switch must still use {-1} (window target).
	if !strings.Contains(joined, "respond {-1}") {
		t.Errorf("fzf args respond binding should use {-1} (window target): %v", args)
	}
	// Last field delimiter for target
	if !strings.Contains(joined, "--delimiter") {
		t.Errorf("fzf args missing --delimiter: %v", args)
	}
}

func TestParseSelectedTarget(t *testing.T) {
	// fzf outputs the selected row; target is the last tab-delimited field.
	// Row format: icon+state \t session \t window \t agent \t paneID \t target
	row := "🔴 \033[31mWAIT\033[0m\tmyapp\tauth-feature\tclaude\t%41\tmyapp:1"
	target := parseSelectedTarget(row)
	if target != "myapp:1" {
		t.Errorf("parseSelectedTarget: got %q want %q", target, "myapp:1")
	}
}

func TestParseSelectedTargetEmpty(t *testing.T) {
	target := parseSelectedTarget("")
	if target != "" {
		t.Errorf("expected empty target for empty row, got %q", target)
	}
}

func TestFzfArgsBackBinding(t *testing.T) {
	args := buildFzfArgs(9999)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "ctrl-b") {
		t.Errorf("fzf args missing ctrl-b back binding: %v", args)
	}
	if !strings.Contains(joined, "pick -") {
		t.Errorf("fzf args ctrl-b binding should invoke 'pick -': %v", args)
	}
}
