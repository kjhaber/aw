package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// hookDir returns the directory where aw installs its hook scripts.
func hookDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "aw", "hooks")
}

// claudeSettingsPath returns the path to ~/.claude/settings.json.
func claudeSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

// hookScriptTemplate is the shell script written for each Claude Code hook.
// %s placeholders: stateDir, paneID format, session, window, windowIndex, stateValue, stateDir, paneID.
const hookScriptTemplate = `#!/usr/bin/env zsh
# aw hook script — managed by 'aw setup', do not edit manually
[[ -z "$TMUX" ]] && exit 0

STATE_DIR="%s"
mkdir -p "$STATE_DIR"

PANE_ID=$(tmux display-message -p "$TMUX_PANE" '#{pane_id}' 2>/dev/null)
SESSION=$(tmux display-message -p "$TMUX_PANE" '#{session_name}' 2>/dev/null)
WINDOW=$(tmux display-message -p "$TMUX_PANE" '#{window_name}' 2>/dev/null)
WINDOW_INDEX=$(tmux display-message -p "$TMUX_PANE" '#{window_index}' 2>/dev/null)

SAFE_ID=$(printf '%%s' "$PANE_ID" | tr '%%' '_')
printf '{"pane":"%%s","session":"%%s","window":"%%s","window_index":"%%s","state":"%s","time":"%%s"}\n' \
  "$PANE_ID" "$SESSION" "$WINDOW" "$WINDOW_INDEX" "$(date -u +%%Y-%%m-%%dT%%H:%%M:%%SZ)" \
  > "$STATE_DIR/$SAFE_ID.json"
`

// writeHookScript writes a hook shell script to path for the given AgentState.
func writeHookScript(path string, state AgentState, stateDir string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("mkdir hook dir: %w", err)
	}
	content := fmt.Sprintf(hookScriptTemplate, stateDir, string(state))
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return fmt.Errorf("write hook script: %w", err)
	}
	return nil
}

// hookEntry is the JSON structure Claude Code expects in hooks arrays.
type hookEntry struct {
	Hooks []hookCommand `json:"hooks"`
}

type hookCommand struct {
	Command string `json:"command"`
	Type    string `json:"type"`
}

// patchSettings merges aw's hooks into the provided settings.json bytes.
// It is idempotent: re-running with the same scripts won't duplicate entries.
func patchSettings(data []byte, stopScript, resumeScript string) ([]byte, error) {
	// Parse as generic map to preserve unknown fields.
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}

	// Ensure "hooks" key exists.
	if doc["hooks"] == nil {
		doc["hooks"] = map[string]any{}
	}
	hooksRaw, ok := doc["hooks"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("settings 'hooks' field is not an object")
	}

	hooksRaw["Stop"] = mergeHook(hooksRaw["Stop"], stopScript)
	hooksRaw["UserPromptSubmit"] = mergeHook(hooksRaw["UserPromptSubmit"], resumeScript)

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal settings: %w", err)
	}
	return append(out, '\n'), nil
}

// mergeHook adds our command to the existing hook array if not already present.
// existing is the raw JSON value for the hook key (may be nil or []any).
func mergeHook(existing any, script string) []any {
	var groups []any
	if existing != nil {
		if arr, ok := existing.([]any); ok {
			groups = arr
		}
	}

	// Check if our script is already there.
	for _, g := range groups {
		gmap, ok := g.(map[string]any)
		if !ok {
			continue
		}
		hooks, ok := gmap["hooks"].([]any)
		if !ok {
			continue
		}
		for _, h := range hooks {
			hmap, ok := h.(map[string]any)
			if !ok {
				continue
			}
			if hmap["command"] == script {
				return groups // already present
			}
		}
	}

	// Not present — append a new group.
	newGroup := map[string]any{
		"hooks": []any{
			map[string]any{
				"command": script,
				"type":    "command",
			},
		},
	}
	return append(groups, newGroup)
}

// cmdSetup installs hook scripts and patches ~/.claude/settings.json.
func cmdSetup(_ []string) error {
	hDir := hookDir()
	sDir := stateDir()

	stopScript := filepath.Join(hDir, "aw-hook-stop")
	resumeScript := filepath.Join(hDir, "aw-hook-resume")

	fmt.Println("Installing hook scripts...")
	if err := writeHookScript(stopScript, AgentStopped, sDir); err != nil {
		return err
	}
	fmt.Printf("  wrote %s\n", stopScript)

	if err := writeHookScript(resumeScript, AgentActive, sDir); err != nil {
		return err
	}
	fmt.Printf("  wrote %s\n", resumeScript)

	settingsPath := claudeSettingsPath()
	fmt.Printf("Patching %s...\n", settingsPath)

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("read settings: %w", err)
	}

	patched, err := patchSettings(data, stopScript, resumeScript)
	if err != nil {
		return err
	}

	if err := os.WriteFile(settingsPath, patched, 0644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	fmt.Println("Done. aw hooks are installed.")
	fmt.Println()
	fmt.Printf("State files will be written to: %s\n", sDir)
	fmt.Println("Run 'aw' in any tmux session to open the agent picker.")
	return nil
}
