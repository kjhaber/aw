# aw — Agent Watcher

A CLI tool for monitoring coding agents (Claude Code, Cursor, etc.) running across
multiple tmux sessions. Pop up an fzf picker that shows which agents need your
attention, with a live preview of their output — then jump straight to the right
window.

## The problem

Running 6–12 concurrent Claude Code sessions across tmux windows is powerful but
cognitively expensive. You have to manually flip through windows, check whether each
agent has stopped (finished, waiting for a reply, or stuck at a permission prompt),
and then context-switch back. `aw` centralizes that triage loop.

## How it works

- `aw setup` installs small shell hooks into `~/.claude/settings.json`. The hooks
  write a JSON state file whenever an agent stops or resumes.
- `aw` (or `aw watch`) launches an fzf picker that reads those state files and
  scans process trees to classify every tmux window:
  - 🔴 **WAIT** — agent stopped, needs your input (shown first)
  - 🟢 **WORK** — agent actively running
- The list live-refreshes every 2 seconds while the picker is open.
- The fzf preview pane shows the last 20 lines of that window's pane content.
- Press **Enter** to switch to the selected window. Press **Esc** to dismiss.

## Install

```sh
brew install kjhaber/tap/aw
```

Then install the Claude Code hooks:

```sh
aw setup
```

**Suggested tmux binding** (add to `~/.tmux.conf`):

```
bind-key A display-popup -E -w '80%' -h '80%' 'aw'
```

### Build from source

```sh
git clone https://github.com/kjhaber/aw && cd aw
make build            # produces build/aw
make install          # copies build/aw to ~/.local/bin/aw
aw setup
```

## Usage

```
aw [watch]         open fzf session picker (default)
aw list            print agent window states to stdout
aw preview TARGET  print last 20 lines of pane at TARGET (session:window)
aw setup           install Claude Code hooks and hook scripts

Keybindings in picker:
  Enter    switch to selected window
  Ctrl-r   force refresh
  Esc      quit
```

## State files

State is written to `~/.local/share/aw/state/` as JSON files keyed by tmux pane ID.
These are safe to delete — `aw` treats a missing state file as "no info, check
processes instead."

## Supported agents

Process detection recognises: `claude`, `cursor`, `pi`, `aider`, `goose`.

## Development

```sh
make all    # fmt + lint + test + build
make test   # tests only
make clean  # remove build/
```

Requires Go 1.26+ and fzf 0.36+.
