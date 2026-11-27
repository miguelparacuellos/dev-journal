# dj — your work journal, in the terminal

**A keyboard-first TUI for the things you track at work but never put in the ticket tracker:**
your standup, your 1:1s, the refactors you'll do "someday", and your wins.
Plain Markdown on disk, with [Claude Code](https://docs.anthropic.com/en/docs/claude-code) and Linear built in.

![dj demo: drafting a standup from Linear, then the 1:1s, Backlog and Wins screens](demo/dj.gif)

<sub>Recorded with [VHS](https://github.com/charmbracelet/vhs) on a fictional journal and a stubbed Claude response (`make demo`).</sub>

- **Standup in seconds.** Press `g` and Claude Code reads your current Linear cycle, then drafts *Yesterday / Today / Blockers*. You tick which suggestions to keep.
- **1:1s with context.** Open actions from earlier 1:1s and your wins since the last one show beside the agenda. Press `P` for a drafted agenda.
- **A home for "I should refactor that".** Capture refactors, ideas and improvements in one line, then filter them by type and repo.
- **Your data stays yours.** Everything is Markdown in a folder outside this repo. The only thing that leaves your machine is a Claude run you start yourself.

<details>
<summary><b>Static screens:</b> Daily, 1:1s and Backlog</summary>

```text
 dev-journal  Fri 27 Nov · Linear ABC
╭─ sections ───────────╮╭─ daily · fri 27 nov ───────────────────────────────────╮╭─ linear · my cycle issues · 09:02 ─────╮
│ 1 ▸ Daily            ││ YESTERDAY ✦ drafted from Linear                        ││ ◑ ABC-420 In Progress                  │
│ 2   1:1s             ││ • ABC-412 Refund flow for cancelled bookings — op…     ││   Webhook retries with backoff         │
│ 3   Backlog          ││                                                        ││ ◐ ABC-412 In Review                    │
│ 4   Wins             ││ TODAY                                                  ││   Refund flow for cancelled bookings   │
╰──────────────────────╯│ [ ] ABC-420 Webhook retries with backoff               ││ ○ ABC-431 Todo                         │
╭─ days ───────────────╮│                                                        ││   Rate-limit docs                      │
│ ▸ Fri 27 Nov today   ││ BLOCKERS                                               ││ ● ABC-98 Done                          │
│   Thu 26 Nov         ││ • Waiting on infra for queue permissions (ABC-420)     │╰────────────────────────────────────────╯
│   Wed 25 Nov         ││                                                        │╭─ next 1:1 ─────────────────────────────╮
│   Tue 24 Nov         ││                                                        ││ Thu 03 Dec · in 6 days                 │
│                      ││                                                        ││ 3 topics prepped                       │
│                      ││                                                        ││ 1 open action from earlier 1:1s        │
│                      ││                                                        ││                                        │
│                      ││                                                        ││                                        │
│                      ││                                                        ││                                        │
│                      ││                                                        ││                                        │
│                      ││                                                        ││                                        │
╰──────────────────────╯╰────────────────────────────────────────────────────────╯╰────────────────────────────────────────╯
  g  ✦ draft from Linear   a  add   x  toggle   d  delete   y  copy for Slack   ←→  day   e  editor   c  claude   q  quit
```

```text
 dev-journal  1:1 · Thu 03 Dec (in 6 days)
╭─ sections ───────────╮╭─ 1:1 · thu 03 dec ─────────────────────────────────────╮╭─ carried actions · open ───────────────╮
│ 1   Daily            ││ TOPICS                                                 ││ [ ] Write ADR for event retries        │
│ 2 ▸ 1:1s             ││ • Ownership of the payments module going forward       ││     from 29 Oct                        │
│ 3   Backlog          ││ • Feedback on my last PR reviews — too nitpicky?       ││                                        │
│ 4   Wins             ││ • Conference budget for next year                      ││                                        │
╰──────────────────────╯│                                                        ││                                        │
╭─ meetings ───────────╮│ NOTES                                                  ││                                        │
│ ▸ 03 Dec 2026 next   ││   (empty — a to add)                                   ││                                        │
│   29 Oct 2026        ││                                                        │╰────────────────────────────────────────╯
│                      ││ FEEDBACK                                               │╭─ wins since 29 Oct ────────────────────╮
│                      ││   (empty — a to add)                                   ││ + Shipped refund flow ABC-412          │
│                      ││                                                        ││ + Fixed flaky checkout e2e             │
│                      ││ ACTIONS                                                ││ + Onboarded a new teammate             │
│                      ││   (empty — a to add)                                   ││                                        │
│                      ││                                                        ││                                        │
│                      ││                                                        ││                                        │
│                      ││                                                        ││                                        │
│                      ││                                                        ││                                        │
╰──────────────────────╯╰────────────────────────────────────────────────────────╯╰────────────────────────────────────────╯
  n  new 1:1   P  ✦ draft agenda   a  add   tab  next section   x  toggle   ←→  meeting   e  editor   c  claude   q  quit
```

```text
 dev-journal  Backlog · 3 open · 1 started
╭─ sections ───────────╮╭─ items ────────────────────────────────────────────────╮╭─ detail ───────────────────────────────╮
│ 1   Daily            ││ TYPE     TITLE                   REPO     EFF STATUS   ││ Contract tests between services        │
│ 2   1:1s             ││ refactor Extract pricing rules … pricing… M   started  ││ idea                                   │
│ 3 ▸ Backlog          ││ idea     Contract tests between… —        L   open     ││                                        │
│ 4   Wins             ││ improve  Faster local seed scri… tooling  S   open     ││ No notes. Press e to add them          │
╰──────────────────────╯│ refactor Kill legacy retry help… payments S   open     ││ in backlog.md.                         │
╭─ filter ─────────────╮│                                                        ││                                        │
│ ▸ all 4              ││                                                        ││                                        │
│   refactor 2         ││                                                        ││                                        │
│   idea 1             ││                                                        ││                                        │
│   improve 1          ││                                                        ││                                        │
│                      ││                                                        ││                                        │
│ done: hidden (.)     ││                                                        ││                                        │
│                      ││                                                        ││                                        │
│                      ││                                                        ││                                        │
│                      ││                                                        ││                                        │
│                      ││                                                        ││                                        │
│                      ││                                                        ││                                        │
╰──────────────────────╯╰────────────────────────────────────────────────────────╯╰────────────────────────────────────────╯
  a  add   s  status   T  type   t  filter   enter  rename   d  delete   e  editor   c  claude   q  quit
```

</details>

## Contents

- [Install](#install)
- [Quick start](#quick-start)
- [Keys](#keys)
- [Configuration](#configuration)
- [Claude Code & Linear](#claude-code--linear)
- [Your data](#your-data)
- [Open it with a hotkey](#open-it-with-a-hotkey)
- [Development](#development)

## Install

**Requirements:**
- Go 1.24+
- A terminal with Unicode and 256/true colour
- *(Optional, for the ✦ features)* the [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) with a Linear MCP connector

```bash
go install github.com/miguelparacuellos/dev-journal/cmd/dj@latest
```

Or from a clone (installs to `~/.local/bin/dj`):

```bash
git clone https://github.com/miguelparacuellos/dev-journal && cd dev-journal && make install
```

> Developed on macOS. On Linux, set `DJ_CLIPBOARD_CMD` (e.g. `wl-copy`) so `y` can copy.

## Quick start

1. Run `dj`. On first run it creates `~/dev-journal/` with a commented `config.env`.
2. Open `~/dev-journal/config.env` and set your Linear team, e.g. `DJ_LINEAR_TEAM=ENG`. You can leave it empty to use all your teams.
3. Press `g` on the Daily screen to draft your standup from Linear. Use `a` to add items by hand and `y` to copy the result for Slack.

Each new day starts from your last entry: what you ticked under *Today* moves to *Yesterday*, and open items and blockers carry over.

## Keys

**Everywhere:**

| Key | Action |
|---|---|
| `1` `2` `3` `4` | Daily · 1:1s · Backlog · Wins |
| `e` | Open the current file in your editor |
| `c` | Open Claude Code in your journal folder |
| `q` / `ctrl+c` | Quit |

**Editing lists** (Daily, 1:1s, Wins):

| Key | Action |
|---|---|
| `j` `k` / `↑` `↓` | Move |
| `tab` / `shift+tab` | Jump to the next / previous section |
| `a` | Add an item below the cursor |
| `enter` | Edit the item |
| `x` / `space` | Toggle a checkbox |
| `d` | Delete |
| `J` `K` | Move the item down / up |

**Per screen:**

| Screen | Key | Action |
|---|---|---|
| Daily | `g` | ✦ Draft from Linear |
| | `y` | Copy as a Slack message |
| | `←` `→` / `t` | Previous / next day, back to today |
| 1:1s | `n` | New 1:1 (asks for the date) |
| | `P` | ✦ Draft agenda topics |
| | `←` `→` | Previous / next meeting |
| Backlog | `a` | Quick add: `refactor: Extract pricing rules #pricing-api !M` |
| | `s` / `T` | Cycle status (open → started → done) / type |
| | `t` / `.` | Filter by type / show done items |
| | `enter` | Rename |
| Wins | `a` | Log a win (today's date is added for you) |

In the ✦ suggestion picker: `space` toggles an item, `enter` inserts the ticked ones, `r` regenerates, `esc` cancels.

## Configuration

Settings come from three places, in this order of priority:
1. Environment variables
2. `$DJ_HOME/config.env`, which `dj` creates on first run
3. The defaults below

Nothing personal is hard-coded, and your `config.env` lives with your data, not in this repo.

| Variable | Default | What it does |
|---|---|---|
| `DJ_HOME` | `~/dev-journal` | Where your journal lives (env var only) |
| `DJ_LINEAR_TEAM` | *empty*, meaning all your teams | Team key whose active cycle feeds the daily draft, e.g. `ENG` |
| `DJ_LINEAR_ASSIGNEE` | `me` | Whose issues to read. `me` is whoever is signed in to the Linear connector |
| `DJ_LINEAR_TOOLS` | `mcp__claude_ai_Linear` | Comma-separated Claude Code tools or MCP servers allowed during the draft ([see below](#which-linear-tool-name)) |
| `DJ_LANGUAGE` | `English` | Language Claude writes drafts in |
| `DJ_CLAUDE_BIN` | `claude` | Claude Code executable |
| `DJ_EDITOR` | `$VISUAL` → `$EDITOR` → `vi` | Editor that `e` opens |
| `DJ_CLIPBOARD_CMD` | `pbcopy` | Command that reads stdin into the clipboard, e.g. `wl-copy` or `xclip -selection clipboard` |

Example `config.env`:

```bash
DJ_LINEAR_TEAM=ENG
DJ_LANGUAGE=Spanish
DJ_EDITOR=code --wait
```

## Claude Code & Linear

`dj` doesn't talk to Linear directly. It uses your **Claude Code** install and whatever **Linear MCP connector** you already have connected.

| | What runs | What Claude can access |
|---|---|---|
| `g` Daily draft | `claude -p` (headless) with a read-only prompt | Only the tools in `DJ_LINEAR_TOOLS` |
| `P` 1:1 agenda | `claude -p` (headless) | No tools. `dj` passes in your dailies, wins and open actions since the last 1:1 |
| `c` Claude Code | Interactive `claude` in your journal folder | Whatever you normally allow |

Both drafts come back as structured suggestions, and nothing is written until you confirm in the picker. A draft takes about 30–60 seconds and counts as one Claude Code run.

The journal folder also gets a `CLAUDE.md` describing the file formats, and two slash commands for interactive sessions: `/daily-draft` and `/one-on-one-prep`.

### Which Linear tool name?

Run `claude mcp list` and find your Linear server:

| Listed as | Set `DJ_LINEAR_TOOLS` to |
|---|---|
| `claude.ai Linear` (claude.ai connector) | `mcp__claude_ai_Linear` *(default)* |
| `linear` (added with `claude mcp add linear …`) | `mcp__linear` |

To check the setup without the TUI, run `dj draft`. It prints the raw draft JSON.

## Your data

```text
~/dev-journal/                  # DJ_HOME
├── config.env                  # your settings
├── daily/2026-11-27.md         # ## Yesterday · ## Today (checkboxes) · ## Blockers
├── one-on-ones/2026-12-03.md   # ## Topics · ## Notes · ## Feedback · ## Actions
├── backlog.md                  # one "## Title" + meta line per item
├── wins.md                     # - 2026-11-20 · Shipped the refund flow
├── CLAUDE.md                   # file formats, for Claude Code
└── .claude/commands/           # /daily-draft, /one-on-one-prep
```

The files are ordinary Markdown: edit them by hand, grep them, or open the folder in Obsidian.
A backlog item looks like this:

```markdown
## Extract pricing rules from controller
type: refactor · repo: pricing-api · effort: M · status: started · added: 2026-11-12

Rules live inline in the controller; hard to test and duplicated in the quote endpoint.
```

**Privacy:**
- This repository contains only code. Your journal is never inside it.
- `.gitignore` also blocks journal files in case you point `DJ_HOME` into a clone.
- If you want history, `git init` your journal folder, but keep it local: don't add a remote.

## Open it with a hotkey

`dj` works best when it's one keystroke away:

- **Warp:** Settings › Features › Keys › Global hotkey › *Dedicated hotkey window*, then type `dj`. For one-keystroke launch, save this as `~/.warp/launch_configurations/dj.yaml` and bind a Raycast or macOS Shortcuts hotkey to `warp://launch/dj`:
  ```yaml
  name: dj
  windows:
    - tabs:
        - layout:
            cwd: ~/dev-journal
            commands:
              - exec: dj
  ```
- **Other terminals:** use the terminal's own drop-down or hotkey window (e.g. iTerm2, Ghostty quick terminal), or a launcher hotkey that opens a terminal running `dj`.

## Development

```bash
make test      # go test ./...
make build     # ./dj
make install   # ~/.local/bin/dj (override with BIN=…)
make demo      # re-record demo/dj.gif (needs vhs)
```

```text
cmd/dj/            entry point (+ `dj draft` debug command)
demo/              VHS tape, fictional journal and Claude stub for the README GIF
internal/config/   env + config.env resolution
internal/store/    Markdown parsing and the journal layout
internal/claude/   headless `claude -p` calls and prompts
internal/ui/       Bubble Tea screens, list editor, suggestion picker
```

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## License

[MIT](LICENSE)
