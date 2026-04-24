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

// TestDeduplicateByWindow verifies that when multiple panes belong to the same
// window, we keep one entry per window using the most urgent state.
func TestDeduplicateByWindow(t *testing.T) {
	entries := []windowEntry{
		// window myapp:0 — Claude pane (WAIT) + shell pane (unknown/skipped)
		{session: "myapp", windowIdx: "0", windowName: "auth", state: stateWait, agentName: "claude", target: "myapp:0"},
		// window myapp:1 — Claude pane (WORK) + shell pane (also present)
		{session: "myapp", windowIdx: "1", windowName: "fix", state: stateWork, agentName: "claude", target: "myapp:1"},
		{session: "myapp", windowIdx: "1", windowName: "fix", state: stateUnknown, agentName: "", target: "myapp:1"},
	}

	got := deduplicateByWindow(entries)

	if len(got) != 2 {
		t.Fatalf("expected 2 window entries, got %d", len(got))
	}

	byTarget := make(map[string]windowEntry)
	for _, e := range got {
		byTarget[e.target] = e
	}

	if byTarget["myapp:0"].state != stateWait {
		t.Errorf("myapp:0: want stateWait, got %v", byTarget["myapp:0"].state)
	}
	// myapp:1 had one WORK pane and one unknown pane — WORK wins
	if byTarget["myapp:1"].state != stateWork {
		t.Errorf("myapp:1: want stateWork, got %v", byTarget["myapp:1"].state)
	}
}

func TestDeduplicateByWindow_WaitBeatsWork(t *testing.T) {
	// Same window, one pane WAIT, one pane WORK — WAIT must win
	entries := []windowEntry{
		{session: "s", windowIdx: "0", windowName: "w", state: stateWork, target: "s:0"},
		{session: "s", windowIdx: "0", windowName: "w", state: stateWait, target: "s:0"},
	}
	got := deduplicateByWindow(entries)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].state != stateWait {
		t.Errorf("want stateWait, got %v", got[0].state)
	}
}

// TestBuildWindowEntryPopulatesPaneID verifies that buildWindowEntry stores the
// pane ID from the TmuxWindow so the preview command can target it directly.
func TestBuildWindowEntryPopulatesPaneID(t *testing.T) {
	stateMap := map[string]PaneState{
		"%55": {PaneID: "%55", AgentState: AgentActive},
	}
	win := TmuxWindow{Session: "s", WindowIndex: "0", WindowName: "w", PaneID: "%55", PanePID: 100}
	e := buildWindowEntry(win, stateMap, nil)
	if e == nil {
		t.Fatal("expected non-nil entry")
	}
	if e.paneID != "%55" {
		t.Errorf("paneID: got %q want %q", e.paneID, "%55")
	}
}

// TestFzfRowPaneIDField verifies that fzfRow embeds the pane ID as the
// second-to-last tab-delimited field, one slot before the window target.
func TestFzfRowPaneIDField(t *testing.T) {
	e := windowEntry{
		session:    "myapp",
		windowIdx:  "1",
		windowName: "auth",
		state:      stateWait,
		agentName:  "claude",
		paneID:     "%41",
		target:     "myapp:1",
	}
	row := e.fzfRow()
	fields := strings.Split(row, "\t")
	// Last field is window target, second-to-last is pane ID.
	if fields[len(fields)-1] != "myapp:1" {
		t.Errorf("last field should be target, got %q", fields[len(fields)-1])
	}
	if fields[len(fields)-2] != "%41" {
		t.Errorf("second-to-last field should be pane ID %%41, got %q", fields[len(fields)-2])
	}
}

// TestDeduplicateByWindowPreservesPaneID verifies that the winning entry's
// pane ID is kept after deduplication.
func TestDeduplicateByWindowPreservesPaneID(t *testing.T) {
	entries := []windowEntry{
		{session: "s", windowIdx: "0", state: stateWait, paneID: "%10", target: "s:0"},
		{session: "s", windowIdx: "0", state: stateWork, paneID: "%11", target: "s:0"},
	}
	got := deduplicateByWindow(entries)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	// stateWait wins; its pane ID (%10) should be preserved.
	if got[0].paneID != "%10" {
		t.Errorf("paneID: got %q want %%10", got[0].paneID)
	}
}
