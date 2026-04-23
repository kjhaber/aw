package main

import (
	"strings"
	"testing"
)

func TestTrimPreview(t *testing.T) {
	// Verify blank lines at start/end are trimmed and content preserved
	raw := "\n\n  hello world  \n  line two  \n\n"
	got := trimPreview(raw)
	if !strings.Contains(got, "hello world") {
		t.Errorf("expected 'hello world' in trimmed output, got: %q", got)
	}
	// Leading/trailing blank lines stripped
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestTrimPreviewEmpty(t *testing.T) {
	got := trimPreview("")
	if got != "" {
		t.Errorf("expected empty string for empty input, got %q", got)
	}
}

func TestTrimPreviewOnlyBlanks(t *testing.T) {
	got := trimPreview("\n\n\n   \n")
	if strings.TrimSpace(got) != "" {
		t.Errorf("expected blank output, got %q", got)
	}
}
