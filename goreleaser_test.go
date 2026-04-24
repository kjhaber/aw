package main

import (
	"os"
	"strings"
	"testing"
)

// TestGoreleaserBrewInstall verifies that the goreleaser config generates a
// correct Homebrew install block for shell completions. The correct pattern
// uses (bash_completion/"aw").write shell_output(...) rather than
// bash_completion.install Utils.safe_popen_read(...), which would pass
// completion script *content* as a file path (causing ENOENT on brew install).
func TestGoreleaserBrewInstall(t *testing.T) {
	data, err := os.ReadFile(".goreleaser.yaml")
	if err != nil {
		t.Fatalf("could not read .goreleaser.yaml: %v", err)
	}
	content := string(data)

	if strings.Contains(content, "safe_popen_read") {
		t.Error(".goreleaser.yaml uses Utils.safe_popen_read for completions; " +
			"this passes script content as a file path and breaks 'brew install'. " +
			"Use (bash_completion/\"aw\").write shell_output(...) instead.")
	}

	for _, check := range []string{
		`(bash_completion/"aw").write`,
		`(zsh_completion/"_aw").write`,
		`(fish_completion/"aw.fish").write`,
	} {
		if !strings.Contains(content, check) {
			t.Errorf(".goreleaser.yaml missing expected install pattern: %s", check)
		}
	}
}
