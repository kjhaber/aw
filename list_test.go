package main

import (
	"strings"
	"testing"
	"time"
)

// windowEntry is the model for one row in the fzf list.
// Tests here verify formatting and sorting — not the live tmux calls.

func TestFormatWindowEntry(t *testing.T) {
	e := windowEntry{
		session:    "myapp",
		windowIdx:  "1",
		windowName: "auth-feature",
		state:      stateWait,
		agentName:  "claude",
		target:     "myapp:1",
	}
	row := e.fzfRow()

	// Must contain the window name
	if !strings.Contains(row, "auth-feature") {
		t.Errorf("row missing window name: %q", row)
	}
	// Must contain the session
	if !strings.Contains(row, "myapp") {
		t.Errorf("row missing session: %q", row)
	}
	// Target must be the last tab-delimited field
	fields := strings.Split(row, "\t")
	if fields[len(fields)-1] != "myapp:1" {
		t.Errorf("last field should be target, got %q", fields[len(fields)-1])
	}
	// State indicator for WAIT should be present
	if !strings.Contains(row, "WAIT") {
		t.Errorf("row missing WAIT state indicator: %q", row)
	}
}

func TestFormatWindowEntryWork(t *testing.T) {
	e := windowEntry{
		session:    "tzc",
		windowIdx:  "0",
		windowName: "main",
		state:      stateWork,
		agentName:  "claude",
		target:     "tzc:0",
	}
	row := e.fzfRow()
	if !strings.Contains(row, "WORK") {
		t.Errorf("row missing WORK state: %q", row)
	}
}

func TestFormatWindowEntryUnknown(t *testing.T) {
	e := windowEntry{
		session:    "tzc",
		windowIdx:  "0",
		windowName: "main",
		state:      stateUnknown,
		target:     "tzc:0",
	}
	row := e.fzfRow()
	// Should not crash and target should still be last field
	fields := strings.Split(row, "\t")
	if fields[len(fields)-1] != "tzc:0" {
		t.Errorf("last field should be target, got %q", fields[len(fields)-1])
	}
}

func TestSortWindowEntries(t *testing.T) {
	entries := []windowEntry{
		{windowName: "c-active", state: stateWork, target: "s:2"},
		{windowName: "a-wait", state: stateWait, target: "s:0"},
		{windowName: "b-wait", state: stateWait, target: "s:1"},
		{windowName: "d-unknown", state: stateUnknown, target: "s:3"},
	}
	sortWindowEntries(entries)

	// WAIT entries must come first
	if entries[0].state != stateWait || entries[1].state != stateWait {
		t.Errorf("first two should be stateWait, got %v %v", entries[0].state, entries[1].state)
	}
	// WORK before unknown
	if entries[2].state != stateWork {
		t.Errorf("third should be stateWork, got %v", entries[2].state)
	}
	if entries[3].state != stateUnknown {
		t.Errorf("fourth should be stateUnknown, got %v", entries[3].state)
	}
}

func TestBuildWindowEntry_StoppedFromState(t *testing.T) {
	stateMap := map[string]PaneState{
		"%41": {
			PaneID:     "%41",
			Session:    "myapp",
			Window:     "auth",
			AgentState: AgentStopped,
			UpdatedAt:  time.Now(),
		},
	}
	procs := []procInfo{} // no live processes needed; state file wins

	win := TmuxWindow{
		Session:     "myapp",
		WindowIndex: "0",
		WindowName:  "auth",
		PaneID:      "%41",
		PanePID:     999,
	}

	e := buildWindowEntry(win, stateMap, procs)
	if e == nil {
		t.Fatal("expected non-nil entry for window with stopped agent")
	}
	if e.state != stateWait {
		t.Errorf("state: got %v want stateWait", e.state)
	}
	if e.target != "myapp:0" {
		t.Errorf("target: got %q want %q", e.target, "myapp:0")
	}
}

func TestBuildWindowEntry_ActiveFromState(t *testing.T) {
	stateMap := map[string]PaneState{
		"%42": {
			PaneID:     "%42",
			AgentState: AgentActive,
		},
	}
	procs := []procInfo{}

	win := TmuxWindow{Session: "myapp", WindowIndex: "1", WindowName: "fix", PaneID: "%42", PanePID: 999}
	e := buildWindowEntry(win, stateMap, procs)
	if e == nil {
		t.Fatal("expected non-nil entry")
	}
	if e.state != stateWork {
		t.Errorf("state: got %v want stateWork", e.state)
	}
}

func TestBuildWindowEntry_NoStateNoProcess(t *testing.T) {
	stateMap := map[string]PaneState{} // no state file
	procs := []procInfo{
		{pid: 999, ppid: 1, comm: "zsh"}, // pane shell, no agent
		{pid: 1000, ppid: 999, comm: "vim"},
	}

	win := TmuxWindow{Session: "myapp", WindowIndex: "2", WindowName: "notes", PaneID: "%43", PanePID: 999}
	e := buildWindowEntry(win, stateMap, procs)
	// No agent detected → nil (skip)
	if e != nil {
		t.Errorf("expected nil for window with no agent, got %+v", e)
	}
}

func TestBuildWindowEntry_NoStateWithProcess(t *testing.T) {
	stateMap := map[string]PaneState{} // no state file
	procs := []procInfo{
		{pid: 999, ppid: 1, comm: "zsh"},
		{pid: 1000, ppid: 999, comm: "claude"},
	}

	win := TmuxWindow{Session: "myapp", WindowIndex: "3", WindowName: "feat", PaneID: "%44", PanePID: 999}
	e := buildWindowEntry(win, stateMap, procs)
	if e == nil {
		t.Fatal("expected non-nil entry when claude process detected")
	}
	if e.state != stateWork {
		t.Errorf("state: got %v want stateWork (process detected, no state file)", e.state)
	}
	if e.agentName != "claude" {
		t.Errorf("agentName: got %q want %q", e.agentName, "claude")
	}
}
