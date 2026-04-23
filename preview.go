package main

import (
	"fmt"
	"strings"
)

const previewLines = 20

// trimPreview strips leading/trailing blank lines from pane capture output.
func trimPreview(raw string) string {
	lines := strings.Split(raw, "\n")

	// Drop leading blank lines
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	// Drop trailing blank lines
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return strings.Join(lines[start:end], "\n")
}

// cmdPreview captures and prints the last previewLines lines of the named target.
func cmdPreview(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: aw preview SESSION:WINDOW")
	}
	target := args[0]

	content, err := capturePane(target, previewLines)
	if err != nil {
		return fmt.Errorf("capture pane %q: %w", target, err)
	}

	fmt.Print(trimPreview(content))
	return nil
}
