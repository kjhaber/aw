package main

import "fmt"

// cmdPick handles 'aw pick -' (go back) and 'aw pick <target>' (switch directly).
// Recording the current window before switching enables toggle-style navigation.
func cmdPick(args []string) error {
	sDir := stateDir()

	goBack := len(args) == 0 || (len(args) == 1 && args[0] == "-")

	if goBack {
		prev, err := readLastWindow(sDir)
		if err != nil {
			return err
		}
		// Record current window so repeated 'aw pick -' toggles back and forth.
		if cur, err := activeWindow(); err == nil && cur != "" {
			_ = writeLastWindow(sDir, cur)
		}
		return switchToWindow(prev)
	}

	target := args[0]
	if target == "" {
		return fmt.Errorf("pick: target must not be empty")
	}
	if cur, err := activeWindow(); err == nil && cur != "" {
		_ = writeLastWindow(sDir, cur)
	}
	return switchToWindow(target)
}
