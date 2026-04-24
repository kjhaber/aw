package main

import (
	"strings"
	"testing"
)

func TestCmdCompletion_NoArgs(t *testing.T) {
	err := cmdCompletion([]string{})
	if err == nil {
		t.Fatal("expected error when no shell specified, got nil")
	}
	if !strings.Contains(err.Error(), "shell") {
		t.Errorf("expected error to mention 'shell', got: %v", err)
	}
}

func TestCmdCompletion_UnknownShell(t *testing.T) {
	err := cmdCompletion([]string{"powershell"})
	if err == nil {
		t.Fatal("expected error for unknown shell, got nil")
	}
	if !strings.Contains(err.Error(), "powershell") {
		t.Errorf("expected error to mention 'powershell', got: %v", err)
	}
}

func TestCompletionBash_ContainsSubcommands(t *testing.T) {
	script, err := completionScript("bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	subcommands := []string{"list", "pick", "preview", "respond", "setup", "watch", "version", "help", "completion"}
	for _, sub := range subcommands {
		if !strings.Contains(script, sub) {
			t.Errorf("bash completion script missing subcommand %q", sub)
		}
	}
	// Should wire up to the 'aw' command
	if !strings.Contains(script, "aw") {
		t.Errorf("bash completion script should reference 'aw'")
	}
	// Should use bash complete builtin
	if !strings.Contains(script, "complete") {
		t.Errorf("bash completion script should use 'complete'")
	}
}

func TestCompletionZsh_ContainsSubcommands(t *testing.T) {
	script, err := completionScript("zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	subcommands := []string{"list", "pick", "preview", "respond", "setup", "watch", "version", "help", "completion"}
	for _, sub := range subcommands {
		if !strings.Contains(script, sub) {
			t.Errorf("zsh completion script missing subcommand %q", sub)
		}
	}
	// Should use #compdef directive
	if !strings.Contains(script, "#compdef aw") {
		t.Errorf("zsh completion script should start with '#compdef aw'")
	}
}

func TestCompletionFish_ContainsSubcommands(t *testing.T) {
	script, err := completionScript("fish")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	subcommands := []string{"list", "pick", "preview", "respond", "setup", "watch", "version", "help", "completion"}
	for _, sub := range subcommands {
		if !strings.Contains(script, sub) {
			t.Errorf("fish completion script missing subcommand %q", sub)
		}
	}
	// Should use fish complete command
	if !strings.Contains(script, "complete -c aw") {
		t.Errorf("fish completion script should use 'complete -c aw'")
	}
}

func TestCompletionScript_UnknownShell(t *testing.T) {
	_, err := completionScript("tcsh")
	if err == nil {
		t.Fatal("expected error for unknown shell 'tcsh', got nil")
	}
	if !strings.Contains(err.Error(), "tcsh") {
		t.Errorf("expected error to mention 'tcsh', got: %v", err)
	}
}
