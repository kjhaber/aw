package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// windowState is the display state for a window entry in the fzf list.
type windowState int

const (
	stateWait    windowState = iota // agent stopped, needs user input (highest priority)
	stateWork                       // agent actively running
	stateUnknown                    // no agent detected
)

// windowEntry is one row in the fzf picker list.
type windowEntry struct {
	session    string
	windowIdx  string
	windowName string
	state      windowState
	agentName  string
	target     string // hidden last field: "session:windowIdx"
}

// stateLabel returns the coloured label for this window's state.
func (e windowEntry) stateLabel() string {
	switch e.state {
	case stateWait:
		return "\033[31mWAIT\033[0m" // red
	case stateWork:
		return "\033[32mWORK\033[0m" // green
	default:
		return "\033[90m----\033[0m" // dim
	}
}

// stateIcon returns a unicode indicator for this window's state.
func (e windowEntry) stateIcon() string {
	switch e.state {
	case stateWait:
		return "🔴"
	case stateWork:
		return "🟢"
	default:
		return "⚫"
	}
}

// fzfRow returns the tab-delimited string for this entry.
// Format: icon+state \t session \t window \t agent \t target(hidden)
// The target is always the last field so fzf can reference it with {-1}.
func (e windowEntry) fzfRow() string {
	agent := e.agentName
	if agent == "" {
		agent = "-"
	}
	label := fmt.Sprintf("%s %s", e.stateIcon(), e.stateLabel())
	return strings.Join([]string{
		label,
		e.session,
		e.windowName,
		agent,
		e.target,
	}, "\t")
}

// sortWindowEntries sorts entries by state priority (WAIT first, then WORK, then unknown).
// Within same state, sort by session+window name for stability.
func sortWindowEntries(entries []windowEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].state != entries[j].state {
			return entries[i].state < entries[j].state
		}
		ki := entries[i].session + ":" + entries[i].windowName
		kj := entries[j].session + ":" + entries[j].windowName
		return ki < kj
	})
}

// buildWindowEntry constructs a windowEntry for win, consulting the state map and
// live process list. Returns nil if no agent is associated with this window.
func buildWindowEntry(win TmuxWindow, stateMap map[string]PaneState, procs []procInfo) *windowEntry {
	e := &windowEntry{
		session:    win.Session,
		windowIdx:  win.WindowIndex,
		windowName: win.WindowName,
		target:     win.Target(),
	}

	if ps, ok := stateMap[win.PaneID]; ok {
		// We have a recorded state for this pane.
		switch ps.AgentState {
		case AgentStopped:
			e.state = stateWait
		case AgentActive:
			e.state = stateWork
		}
		// Try to derive agent name from process tree as well.
		e.agentName = agentName(win.PanePID, procs)
		return e
	}

	// No state file — fall back to process detection.
	name := agentName(win.PanePID, procs)
	if name == "" {
		return nil // no agent here, skip
	}
	e.state = stateWork
	e.agentName = name
	return e
}

// collectEntries gathers all window entries across tmux sessions.
func collectEntries() ([]windowEntry, error) {
	wins, err := listTmuxWindows()
	if err != nil {
		return nil, fmt.Errorf("list tmux windows: %w", err)
	}

	rawStates, err := readAllStates(stateDir())
	if err != nil {
		return nil, fmt.Errorf("read states: %w", err)
	}
	stateMap := make(map[string]PaneState, len(rawStates))
	for _, s := range rawStates {
		stateMap[s.PaneID] = s
	}

	procRaw, err := liveProcessList()
	if err != nil {
		// Process detection failing is non-fatal; carry on with empty procs.
		procRaw = ""
	}
	procs := parseProcessList(procRaw)

	var entries []windowEntry
	for _, win := range wins {
		e := buildWindowEntry(win, stateMap, procs)
		if e != nil {
			entries = append(entries, *e)
		}
	}
	sortWindowEntries(entries)
	return entries, nil
}

// cmdList prints fzf-formatted rows to stdout.
func cmdList(_ []string) error {
	entries, err := collectEntries()
	if err != nil {
		return err
	}
	for _, e := range entries {
		fmt.Println(e.fzfRow())
	}
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "aw: no agent windows detected")
	}
	return nil
}
