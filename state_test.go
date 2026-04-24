package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStateRoundTrip(t *testing.T) {
	dir := t.TempDir()

	s := PaneState{
		PaneID:      "%42",
		Session:     "myapp",
		Window:      "auth-feature",
		WindowIndex: "1",
		AgentState:  AgentStopped,
		UpdatedAt:   time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC),
	}

	if err := writeState(dir, s); err != nil {
		t.Fatalf("writeState: %v", err)
	}

	got, err := readState(dir, "%42")
	if err != nil {
		t.Fatalf("readState: %v", err)
	}

	if got.PaneID != s.PaneID {
		t.Errorf("PaneID: got %q want %q", got.PaneID, s.PaneID)
	}
	if got.Session != s.Session {
		t.Errorf("Session: got %q want %q", got.Session, s.Session)
	}
	if got.Window != s.Window {
		t.Errorf("Window: got %q want %q", got.Window, s.Window)
	}
	if got.AgentState != s.AgentState {
		t.Errorf("AgentState: got %q want %q", got.AgentState, s.AgentState)
	}
	if !got.UpdatedAt.Equal(s.UpdatedAt) {
		t.Errorf("UpdatedAt: got %v want %v", got.UpdatedAt, s.UpdatedAt)
	}
}

func TestReadStateMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := readState(dir, "%99")
	if err == nil {
		t.Fatal("expected error for missing state file, got nil")
	}
}

func TestReadAllStates(t *testing.T) {
	dir := t.TempDir()

	states := []PaneState{
		{PaneID: "%1", Session: "app", Window: "feat-a", WindowIndex: "0", AgentState: AgentStopped},
		{PaneID: "%2", Session: "app", Window: "feat-b", WindowIndex: "1", AgentState: AgentActive},
	}
	for _, s := range states {
		if err := writeState(dir, s); err != nil {
			t.Fatalf("writeState: %v", err)
		}
	}

	all, err := readAllStates(dir)
	if err != nil {
		t.Fatalf("readAllStates: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("readAllStates: got %d states, want 2", len(all))
	}

	byPane := make(map[string]PaneState)
	for _, s := range all {
		byPane[s.PaneID] = s
	}
	if byPane["%1"].AgentState != AgentStopped {
		t.Errorf("%%1: want AgentStopped, got %q", byPane["%1"].AgentState)
	}
	if byPane["%2"].AgentState != AgentActive {
		t.Errorf("%%2: want AgentActive, got %q", byPane["%2"].AgentState)
	}
}

func TestReadAllStatesEmptyDir(t *testing.T) {
	dir := t.TempDir()
	all, err := readAllStates(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 0 {
		t.Fatalf("expected empty, got %d", len(all))
	}
}

func TestReadAllStatesIgnoresNonJSON(t *testing.T) {
	dir := t.TempDir()
	// Write a non-.json file — should be ignored
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	all, err := readAllStates(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 0 {
		t.Fatalf("expected empty, got %d", len(all))
	}
}

func TestLastWindowRoundTrip(t *testing.T) {
	dir := t.TempDir()
	target := "myapp:3"
	if err := writeLastWindow(dir, target); err != nil {
		t.Fatalf("writeLastWindow: %v", err)
	}
	got, err := readLastWindow(dir)
	if err != nil {
		t.Fatalf("readLastWindow: %v", err)
	}
	if got != target {
		t.Errorf("got %q want %q", got, target)
	}
}

func TestReadLastWindowMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := readLastWindow(dir)
	if err == nil {
		t.Fatal("expected error for missing last-window file, got nil")
	}
}

func TestLastWindowOverwrite(t *testing.T) {
	dir := t.TempDir()
	if err := writeLastWindow(dir, "session1:0"); err != nil {
		t.Fatalf("writeLastWindow: %v", err)
	}
	if err := writeLastWindow(dir, "session2:1"); err != nil {
		t.Fatalf("writeLastWindow second: %v", err)
	}
	got, err := readLastWindow(dir)
	if err != nil {
		t.Fatalf("readLastWindow: %v", err)
	}
	if got != "session2:1" {
		t.Errorf("got %q want %q", got, "session2:1")
	}
}
