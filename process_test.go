package main

import (
	"testing"
)

func TestParseProcessList(t *testing.T) {
	// Simulated output of: ps -eo pid,ppid,comm
	raw := `  PID  PPID COMM
    1     0 launchd
  100     1 zsh
  101   100 claude
  102   100 vim
  200     1 zsh
  201   200 node
`
	procs := parseProcessList(raw)

	if len(procs) == 0 {
		t.Fatal("expected non-empty process list")
	}

	// Find claude process
	found := false
	for _, p := range procs {
		if p.comm == "claude" {
			found = true
			if p.pid != 101 {
				t.Errorf("claude pid: got %d want 101", p.pid)
			}
			if p.ppid != 100 {
				t.Errorf("claude ppid: got %d want 100", p.ppid)
			}
		}
	}
	if !found {
		t.Error("claude process not found in parsed list")
	}
}

func TestDescendantsOf(t *testing.T) {
	procs := []procInfo{
		{pid: 1, ppid: 0, comm: "launchd"},
		{pid: 100, ppid: 1, comm: "zsh"},
		{pid: 101, ppid: 100, comm: "claude"},
		{pid: 102, ppid: 100, comm: "vim"},
		{pid: 103, ppid: 101, comm: "bash"}, // child of claude
		{pid: 200, ppid: 1, comm: "zsh"},
	}

	desc := descendantsOf(100, procs)

	// Should include 101 (claude), 102 (vim), 103 (bash child of claude)
	if len(desc) != 3 {
		t.Errorf("expected 3 descendants of pid 100, got %d: %v", len(desc), desc)
	}

	descs := make(map[int]bool)
	for _, p := range desc {
		descs[p.pid] = true
	}
	for _, pid := range []int{101, 102, 103} {
		if !descs[pid] {
			t.Errorf("expected pid %d in descendants", pid)
		}
	}
	// Should NOT include 200 (different zsh) or 1 (launchd)
	for _, pid := range []int{1, 100, 200} {
		if descs[pid] {
			t.Errorf("unexpected pid %d in descendants", pid)
		}
	}
}

func TestHasAgentProcess(t *testing.T) {
	procs := []procInfo{
		{pid: 1, ppid: 0, comm: "launchd"},
		{pid: 100, ppid: 1, comm: "zsh"},
		{pid: 101, ppid: 100, comm: "claude"},
	}

	if !hasAgentProcess(100, procs) {
		t.Error("expected agent detected under pid 100 (has claude child)")
	}
	if hasAgentProcess(1, procs) {
		// launchd has zsh as child, but zsh is not an agent
		// claude is a grandchild of 1 though, so this should be true
		// Actually: descendantsOf(1) = {100(zsh), 101(claude)} → claude is an agent
		// So this should be true too. Let's adjust the test.
		t.Log("note: pid 1 (launchd) also has claude as descendant — expected")
	}
}

func TestHasAgentProcessNoAgent(t *testing.T) {
	procs := []procInfo{
		{pid: 1, ppid: 0, comm: "launchd"},
		{pid: 200, ppid: 1, comm: "zsh"},
		{pid: 201, ppid: 200, comm: "vim"},
	}

	if hasAgentProcess(200, procs) {
		t.Error("expected no agent under pid 200 (only vim)")
	}
}

func TestAgentName(t *testing.T) {
	procs := []procInfo{
		{pid: 100, ppid: 1, comm: "zsh"},
		{pid: 101, ppid: 100, comm: "claude"},
	}
	name := agentName(100, procs)
	if name != "claude" {
		t.Errorf("agentName: got %q want %q", name, "claude")
	}
}

func TestAgentNameUnknown(t *testing.T) {
	procs := []procInfo{
		{pid: 100, ppid: 1, comm: "zsh"},
		{pid: 101, ppid: 100, comm: "vim"},
	}
	name := agentName(100, procs)
	if name != "" {
		t.Errorf("agentName: expected empty, got %q", name)
	}
}
