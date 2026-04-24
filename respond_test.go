package main

import (
	"strings"
	"testing"
)

func TestCmdRespondMissingTarget(t *testing.T) {
	err := cmdRespond([]string{})
	if err == nil {
		t.Error("expected error when no arguments given")
	}
}

func TestCmdRespondEmptyTarget(t *testing.T) {
	err := cmdRespond([]string{""})
	if err == nil {
		t.Error("expected error for empty target")
	}
}

func TestRespondDefaultKeys(t *testing.T) {
	if defaultRespondKeys == "" {
		t.Error("defaultRespondKeys must not be empty")
	}
}

func TestFzfArgsRespondBinding(t *testing.T) {
	args := buildFzfArgs(9999)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "ctrl-y") {
		t.Errorf("fzf args missing ctrl-y respond binding: %v", args)
	}
	if !strings.Contains(joined, "respond") {
		t.Errorf("fzf args ctrl-y binding should invoke 'respond': %v", args)
	}
}

func TestFzfArgsRespondInHeader(t *testing.T) {
	args := buildFzfArgs(9999)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "Ctrl-y") {
		t.Errorf("fzf header should mention Ctrl-y for respond: %v", args)
	}
}
