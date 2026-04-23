package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AgentState represents the known state of an agent in a pane.
type AgentState string

const (
	AgentStopped AgentState = "stopped" // agent stopped, needs user input
	AgentActive  AgentState = "active"  // agent is running
)

// PaneState is the persisted record for a single tmux pane.
type PaneState struct {
	PaneID      string     `json:"pane"`
	Session     string     `json:"session"`
	Window      string     `json:"window"`
	WindowIndex string     `json:"window_index"`
	AgentState  AgentState `json:"state"`
	UpdatedAt   time.Time  `json:"time"`
}

// stateDir returns the default state directory.
func stateDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "aw", "state")
}

// stateFilePath returns the path for a given pane ID (e.g. "%42" → ".../state/%42.json").
// The '%' is replaced with '_' to avoid shell quoting issues in filenames.
func stateFilePath(dir, paneID string) string {
	safe := strings.ReplaceAll(paneID, "%", "_")
	return filepath.Join(dir, safe+".json")
}

// writeState persists s to dir.
func writeState(dir string, s PaneState) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir state dir: %w", err)
	}
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	path := stateFilePath(dir, s.PaneID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}
	return nil
}

// readState reads the state for a single pane ID from dir.
func readState(dir, paneID string) (PaneState, error) {
	data, err := os.ReadFile(stateFilePath(dir, paneID))
	if err != nil {
		return PaneState{}, fmt.Errorf("read state file: %w", err)
	}
	var s PaneState
	if err := json.Unmarshal(data, &s); err != nil {
		return PaneState{}, fmt.Errorf("parse state file: %w", err)
	}
	return s, nil
}

// readAllStates reads all state files from dir. Non-.json files are ignored.
func readAllStates(dir string) ([]PaneState, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read state dir: %w", err)
	}

	var states []PaneState
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue // best-effort
		}
		var s PaneState
		if err := json.Unmarshal(data, &s); err != nil {
			continue // best-effort
		}
		states = append(states, s)
	}
	return states, nil
}
