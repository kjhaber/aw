package main

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// knownAgents is the set of process names recognized as coding agents.
var knownAgents = map[string]bool{
	"claude": true,
	"cursor": true,
	"pi":     true,
	"aider":  true,
	"goose":  true,
}

type procInfo struct {
	pid  int
	ppid int
	comm string
}

// liveProcessList runs ps and returns the output for parsing.
func liveProcessList() (string, error) {
	out, err := exec.Command("ps", "-eo", "pid,ppid,comm").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// parseProcessList parses the output of `ps -eo pid,ppid,comm`.
func parseProcessList(raw string) []procInfo {
	var procs []procInfo
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		// Use basename: macOS ps -eo comm can return the full path.
		comm := filepath.Base(fields[2])
		procs = append(procs, procInfo{pid: pid, ppid: ppid, comm: comm})
	}
	return procs
}

// descendantsOf returns all processes whose ancestor (at any depth) is parentPID.
func descendantsOf(parentPID int, procs []procInfo) []procInfo {
	// Build children map
	children := make(map[int][]procInfo)
	for _, p := range procs {
		children[p.ppid] = append(children[p.ppid], p)
	}

	var result []procInfo
	queue := []int{parentPID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range children[cur] {
			result = append(result, child)
			queue = append(queue, child.pid)
		}
	}
	return result
}

// hasAgentProcess reports whether any descendant of rootPID is a known agent.
func hasAgentProcess(rootPID int, procs []procInfo) bool {
	for _, p := range descendantsOf(rootPID, procs) {
		if knownAgents[p.comm] {
			return true
		}
	}
	return false
}

// agentName returns the name of the first known agent found under rootPID, or "".
func agentName(rootPID int, procs []procInfo) string {
	for _, p := range descendantsOf(rootPID, procs) {
		if knownAgents[p.comm] {
			return p.comm
		}
	}
	return ""
}
