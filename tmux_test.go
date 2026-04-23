package main

import (
	"testing"
)

func TestParseTmuxWindows(t *testing.T) {
	// Simulated output of: tmux list-windows -a -F '#{session_name}\t#{window_index}\t#{window_name}\t#{window_active}\t#{pane_id}\t#{pane_pid}'
	raw := "myapp\t0\tauth-feature\t0\t%41\t101\nmyapp\t1\tfix-perf\t1\t%42\t200\ntzc\t0\tmain\t0\t%43\t301\n"

	wins := parseTmuxWindows(raw)
	if len(wins) != 3 {
		t.Fatalf("expected 3 windows, got %d", len(wins))
	}

	w := wins[0]
	if w.Session != "myapp" {
		t.Errorf("Session: got %q want %q", w.Session, "myapp")
	}
	if w.WindowIndex != "0" {
		t.Errorf("WindowIndex: got %q want %q", w.WindowIndex, "0")
	}
	if w.WindowName != "auth-feature" {
		t.Errorf("WindowName: got %q want %q", w.WindowName, "auth-feature")
	}
	if w.Active {
		t.Error("window 0 should not be active")
	}
	if w.PaneID != "%41" {
		t.Errorf("PaneID: got %q want %q", w.PaneID, "%41")
	}
	if w.PanePID != 101 {
		t.Errorf("PanePID: got %d want 101", w.PanePID)
	}

	if !wins[1].Active {
		t.Error("window 1 should be active")
	}
}

func TestParseTmuxWindowsEmpty(t *testing.T) {
	wins := parseTmuxWindows("")
	if len(wins) != 0 {
		t.Errorf("expected 0 windows for empty input, got %d", len(wins))
	}
}

func TestParseTmuxWindowsMalformed(t *testing.T) {
	// A line with wrong field count should be skipped
	raw := "myapp\t0\tauth-feature\n" // missing fields
	wins := parseTmuxWindows(raw)
	if len(wins) != 0 {
		t.Errorf("expected 0 windows for malformed line, got %d", len(wins))
	}
}

func TestTmuxTarget(t *testing.T) {
	w := TmuxWindow{Session: "myapp", WindowIndex: "2"}
	if got := w.Target(); got != "myapp:2" {
		t.Errorf("Target(): got %q want %q", got, "myapp:2")
	}
}
