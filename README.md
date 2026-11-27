# dev-journal (`dj`)

Keyboard-first TUI for a developer's work log: daily standup, monthly 1:1s, a backlog of refactors/ideas, and a wins log.
Data is plain Markdown in `~/dev-journal` (override with `DJ_HOME`), so Claude Code, your editor or Obsidian can read and edit it too.

```bash
make install   # builds to ~/.local/bin/dj
dj
```

## Privacy

This repo holds only code. Your journal (dailies, 1:1 notes, backlog, wins, Linear cache) is written to `DJ_HOME`
(`~/dev-journal` by default), outside the repo, and never committed. `.gitignore` also ignores those files in case
`DJ_HOME` points inside the repo. Keep `DJ_HOME` local; if you `git init` it for history, don't add a remote.

## Keys

Global: `1`–`4` switch section · `e` open the current file in `$EDITOR` · `c` open Claude Code in the journal · `q` quit.

| Screen | Keys |
|---|---|
| Daily | `g` ✦ draft from Linear · `a` add · `enter` edit · `x` toggle · `d` delete · `J/K` move · `tab` next section · `←/→` day · `t` today · `y` copy for Slack |
| 1:1s | `n` new 1:1 · `P` ✦ draft agenda · same editing keys · `←/→` meeting |
| Backlog | `a` add (`refactor: Title #repo !M`) · `s` cycle status · `T` cycle type · `t` filter type · `.` show done · `enter` rename · `d` delete |
| Wins | `a` add (dated automatically) · `enter` edit · `d` delete |

A new day starts from the latest daily: done "Today" items become "Yesterday", open ones carry over.
Open 1:1 actions carry to the next 1:1, and wins since the previous 1:1 show beside it.

## Claude integration

- `g` runs `claude -p` headless with only the Linear connector allowed (read-only prompt), then lets you pick which
  suggestions go into Yesterday / Today / Blockers. It also refreshes the Linear side panel.
- `P` sends your dailies, wins and open actions since the last 1:1 to `claude -p` (no tools) and suggests agenda topics.
- `c` opens interactive Claude Code in the journal dir. `CLAUDE.md` there explains the file formats, and
  `/daily-draft` and `/one-on-one-prep` are available as slash commands.
- `dj draft` prints the raw Linear draft JSON (debugging).

Env: `DJ_HOME`, `DJ_LINEAR_TEAM` (default `DEV`), `DJ_LINEAR_TOOLS` (default `mcp__claude_ai_Linear`), `DJ_CLAUDE_BIN`.

## Global hotkey in Warp

Settings › Features › Keys › Global hotkey › **Dedicated hotkey window**, pick a shortcut (e.g. `⌥ Space`), then type `dj`.
For a one-keystroke launch, add a launch configuration `~/.warp/launch_configurations/dj.yaml`:

```yaml
name: dj
windows:
  - tabs:
      - layout:
          cwd: ~/dev-journal
          commands:
            - exec: dj
```

and bind a Raycast / macOS Shortcuts hotkey to open `warp://launch/dj`.
