package main

import (
	"fmt"
	"strings"
)

var completionSubcommands = []string{
	"list", "pick", "preview", "respond", "setup", "watch", "version", "help", "completion",
}

var subcommandDescriptions = map[string]string{
	"list":       "print agent window states to stdout",
	"pick":       "switch to TARGET (session:window)",
	"preview":    "print last 20 lines of pane at TARGET",
	"respond":    "send text to TARGET's pane and press Enter",
	"setup":      "install Claude Code hooks and hook scripts",
	"watch":      "open fzf session picker (default)",
	"version":    "print version",
	"help":       "print usage",
	"completion": "print shell completion script",
}

func cmdCompletion(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: aw completion <shell>\nSupported shells: bash, zsh, fish")
	}
	script, err := completionScript(args[0])
	if err != nil {
		return err
	}
	fmt.Print(script)
	return nil
}

func completionScript(shell string) (string, error) {
	switch shell {
	case "bash":
		return bashCompletionScript(), nil
	case "zsh":
		return zshCompletionScript(), nil
	case "fish":
		return fishCompletionScript(), nil
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish)", shell)
	}
}

func bashCompletionScript() string {
	subcommands := strings.Join(completionSubcommands, " ")
	return fmt.Sprintf(`# bash completion for aw
# Add to your ~/.bashrc or ~/.bash_profile:
#   eval "$(aw completion bash)"

_aw_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local subcommands="%s"
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=($(compgen -W "${subcommands}" -- "${cur}"))
    fi
}

complete -F _aw_completions aw
`, subcommands)
}

func zshCompletionScript() string {
	var b strings.Builder
	for _, sub := range completionSubcommands {
		desc := subcommandDescriptions[sub]
		fmt.Fprintf(&b, "        '%s:%s'\n", sub, desc)
	}
	descriptions := b.String()
	return fmt.Sprintf(`#compdef aw
# zsh completion for aw
# Add to your ~/.zshrc:
#   eval "$(aw completion zsh)"
# Or place this file in a directory on your $fpath.

_aw() {
    local -a subcommands
    subcommands=(
%s    )

    if [[ ${CURRENT} -eq 2 ]]; then
        _describe 'subcommand' subcommands
    fi
}

_aw
`, descriptions)
}

func fishCompletionScript() string {
	var b strings.Builder
	b.WriteString("# fish completion for aw\n")
	b.WriteString("# Add to your fish config:\n")
	b.WriteString("#   aw completion fish | source\n\n")
	for _, sub := range completionSubcommands {
		desc := subcommandDescriptions[sub]
		fmt.Fprintf(&b, "complete -c aw -f -n '__fish_use_subcommand' -a %s -d '%s'\n", sub, desc)
	}
	return b.String()
}
