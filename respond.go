package main

import "fmt"

// defaultRespondKeys is the text sent when no explicit response is given.
// "1" selects the primary (yes/allow) option in Claude Code's numbered prompts.
const defaultRespondKeys = "1"

// cmdRespond handles 'aw respond TARGET [text]' — sends text followed by Enter
// to the active pane at TARGET, allowing the user to answer agent prompts
// (permission requests, questions) directly without switching to the window.
func cmdRespond(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: aw respond SESSION:WINDOW [text]")
	}
	target := args[0]
	if target == "" {
		return fmt.Errorf("respond: target must not be empty")
	}
	text := defaultRespondKeys
	if len(args) >= 2 {
		text = args[1]
	}
	return sendKeysToTarget(target, text)
}
