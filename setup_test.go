package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatchSettingsAddsHooks(t *testing.T) {
	// Start with settings that have no hooks at all
	initial := `{"permissions": {"allow": []}}`

	stopScript := "/home/user/.local/share/aw/hooks/aw-hook-stop"
	resumeScript := "/home/user/.local/share/aw/hooks/aw-hook-resume"

	patched, err := patchSettings([]byte(initial), stopScript, resumeScript)
	if err != nil {
		t.Fatalf("patchSettings: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(patched, &doc); err != nil {
		t.Fatalf("unmarshal patched: %v", err)
	}

	hooks, ok := doc["hooks"].(map[string]any)
	if !ok {
		t.Fatal("patched settings missing 'hooks' key")
	}

	stopHooks, ok := hooks["Stop"].([]any)
	if !ok || len(stopHooks) == 0 {
		t.Fatal("patched settings missing Stop hooks")
	}

	resumeHooks, ok := hooks["UserPromptSubmit"].([]any)
	if !ok || len(resumeHooks) == 0 {
		t.Fatal("patched settings missing UserPromptSubmit hooks")
	}
}

func TestPatchSettingsIdempotent(t *testing.T) {
	stopScript := "/share/aw/hooks/aw-hook-stop"
	resumeScript := "/share/aw/hooks/aw-hook-resume"

	initial := `{"permissions": {}}`

	// Apply twice — should not duplicate entries
	once, err := patchSettings([]byte(initial), stopScript, resumeScript)
	if err != nil {
		t.Fatalf("first patchSettings: %v", err)
	}
	twice, err := patchSettings(once, stopScript, resumeScript)
	if err != nil {
		t.Fatalf("second patchSettings: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(twice, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Count Stop hook entries — should be exactly 1 group with our command
	raw := string(twice)
	count := strings.Count(raw, stopScript)
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of stop script, got %d\n%s", count, raw)
	}
}

func TestPatchSettingsPreservesExistingHooks(t *testing.T) {
	stopScript := "/share/aw/hooks/aw-hook-stop"
	resumeScript := "/share/aw/hooks/aw-hook-resume"

	// Settings already have a Stop hook from something else
	initial := `{
  "hooks": {
    "Stop": [{"hooks": [{"command": "some-other-script", "type": "command"}]}]
  }
}`

	patched, err := patchSettings([]byte(initial), stopScript, resumeScript)
	if err != nil {
		t.Fatalf("patchSettings: %v", err)
	}

	raw := string(patched)
	if !strings.Contains(raw, "some-other-script") {
		t.Error("patchSettings removed existing Stop hook")
	}
	if !strings.Contains(raw, stopScript) {
		t.Error("patchSettings did not add our Stop hook")
	}
}

func TestWriteHookScript(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "aw-hook-stop")
	stateDir := filepath.Join(dir, "state")

	if err := writeHookScript(path, AgentStopped, stateDir); err != nil {
		t.Fatalf("writeHookScript: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read script: %v", err)
	}
	content := string(data)

	// Must be executable
	info, _ := os.Stat(path)
	if info.Mode()&0111 == 0 {
		t.Error("hook script is not executable")
	}

	// Must reference the state dir and the correct state value
	if !strings.Contains(content, stateDir) {
		t.Errorf("hook script missing state dir %q: %s", stateDir, content)
	}
	if !strings.Contains(content, string(AgentStopped)) {
		t.Errorf("hook script missing state value %q: %s", AgentStopped, content)
	}
	// Must be a shell script
	if !strings.HasPrefix(content, "#!/") {
		t.Errorf("hook script missing shebang: %s", content)
	}
}
