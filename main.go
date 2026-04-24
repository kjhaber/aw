package main

import (
	"fmt"
	"os"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "aw: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cmd := "watch"
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	switch cmd {
	case "list":
		return cmdList(args)
	case "pick":
		return cmdPick(args)
	case "preview":
		return cmdPreview(args)
	case "setup":
		return cmdSetup(args)
	case "watch", "run":
		return cmdWatch(args)
	case "version", "--version", "-v":
		fmt.Println("aw version", version)
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown subcommand %q (try: aw help)", cmd)
	}
}

func printUsage() {
	fmt.Print(`aw — Agent Watcher: monitor coding agents across tmux sessions

Usage:
  aw [watch]         open fzf session picker (default)
  aw pick -          switch back to the previously visited window
  aw pick TARGET     switch directly to TARGET (session:window)
  aw list            print agent window states to stdout
  aw preview TARGET  print last 20 lines of pane at TARGET (session:window)
  aw setup           install Claude Code hooks and hook scripts

Keybindings in picker:
  Enter    switch to selected window
  Ctrl-b   go back to previous window
  Ctrl-r   force refresh
  Esc/q    quit
`)
}
