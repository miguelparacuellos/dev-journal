# Dev Journal

**A calm, local-first terminal journal for developers.** Capture what you did in
a few keystrokes, walk into your daily stand-up knowing exactly what to say, and
never arrive at a one-to-one having forgotten the thing you wanted to raise.

- **Offline and private.** Your journal is a single JSON file on your machine.
  The app never opens a network connection.
- **Built for brief visits.** Open it, type, press **Ctrl+S**, close it. The
  composer is focused the moment the app starts, and startup takes about 70 ms.
- **Safe by default.** Saves are atomic and synced to disk, drafts are never
  silently thrown away, and corrupted or unknown files fail loudly instead of
  being reset.
- **Scriptable.** Everything important also works from the command line, so you
  can log work from a shell alias, a Git hook, or a script.

```
devjournal                     # open the TUI on Today
devjournal add "Shipped the login fix"   # or capture without opening it
```

---

## Contents

- [What it helps you with](#what-it-helps-you-with)
- [Installation](#installation)
- [Quick start](#quick-start)
- [A typical day](#a-typical-day)
- [Command-line reference](#command-line-reference)
- [Keyboard reference](#keyboard-reference)
- [Your data](#your-data)
- [Terminal support](#terminal-support)
- [Development](#development)
- [Project status and limitations](#project-status-and-limitations)
- [License](#license)

## What it helps you with

| View | What it is for |
| --- | --- |
| **Today** | A log of what you got done today (outcomes, not timesheets) plus the plan of tasks you intend to work on. |
| **Tasks** | Things you want to do at some point. Pick some for today's plan; unfinished ones stay open, and each day's plan starts empty. |
| **Daily** | Stand-up preparation. Shows your latest previous workday (even after a weekend or time off), today's plan, and blockers, so you can pick what to share and save a proposal in your own words. |
| **O2O** | One-to-one preparation. Collect topics whenever they come up, then run the meeting: mark topics addressed, take notes, record agreements, and create follow-up tasks. |

When you need the data elsewhere, export it to **Markdown** for reading or make
a complete **JSON backup** that can be restored later.

The vocabulary used throughout the app (entry, blocker, daily proposal,
addressed topic, and so on) is defined in [`CONTEXT.md`](CONTEXT.md).

## Installation

Dev Journal requires **Go 1.27 or newer** and runs on **macOS and Linux**.

Install the latest version onto your `PATH`:

```sh
go install github.com/miguelparacuellos/dev-journal/cmd/devjournal@latest
```

Make sure your Go bin directory (usually `~/go/bin`) is on your `PATH`, then run:

```sh
devjournal
```

Or build from source:

```sh
git clone https://github.com/miguelparacuellos/dev-journal.git
cd dev-journal
go build -o bin/devjournal ./cmd/devjournal
./bin/devjournal
```

## Quick start

1. Run `devjournal`. The TUI opens in **Today** with the composer focused.
2. Type what you did. **Enter** inserts a line break; **Ctrl+S** saves.
3. Press **Esc** to stop typing and browse. Press **?** at any time for help.
4. Press **q** to quit.

That is the whole core loop. Everything else is optional.

> **Tip:** Esc keeps your unsaved draft. If you try to leave with a draft, the
> app asks you to save it or explicitly discard it with **x**. Drafts live only
> in the running session, so force-killing the process can lose unsaved text.

## A typical day

### Log work as it happens

From the TUI, press **n** to capture an entry, or skip the TUI entirely:

```sh
devjournal add "Fixed login and added regression coverage"
printf 'Fixed login\nAdded coverage' | devjournal add -      # multiline from stdin
devjournal --date 2026-09-25 add "Reviewed deployment"       # backfill another day
```

Made a mistake? Select the entry and press **e** to correct it. Corrections keep
the entry on its original workday.

### Plan the day

Capture tasks in **Tasks** (**Tab** from Today) with **n**, then press **p** to
add the selected task to today's plan. In Today, **p** focuses the plan, **d**
marks a task done, and **u** removes it from the plan.

Planning never completes a task or creates a log entry on its own. Tasks you
don't finish stay open, and tomorrow's plan starts empty so you choose again
deliberately.

### Prepare the daily stand-up

Press **b** to record a blocker, then **g** to open **Daily**. It shows three
sections: recent work from your last workday, today's plan, and blockers.

1. Use **Left/Right** to pick a section and **Up/Down** to pick a row.
2. In recent work, press **s** to mark the items you want to share.
3. Press **e** to prepare a proposal from your selection, edit it freely, and
   save with **Ctrl+S**.
4. Press **v** to read it back during the meeting.

Your saved proposal is stored per day and never changes when the source entries
do. Press **r** to start a fresh proposal from the current selection.

### Run a one-to-one

Press **o** to open **O2O** and **n** to add a topic whenever one comes to mind.
Topics stay open across days and restarts until they are addressed.

When the meeting starts, press **s**. Then, for the selected topic or meeting:

| Key | Action |
| --- | --- |
| **a** | Mark the topic addressed (press again to undo) |
| **w** | Write or edit the meeting notes |
| **r** | Record an agreement |
| **f** | Create a follow-up task (it shows up in Tasks marked `O2O:`) |
| **v** | Read the full meeting record |
| **c**, **c** | Close the meeting (the second **c** confirms) |

Closing keeps everything you recorded. Unaddressed topics stay open for next
time. Press **m** to browse past meetings, newest first.

## Command-line reference

Global options go **before** the command:

```
devjournal [options] [command] [arguments]
```

| Option | Description |
| --- | --- |
| `--data PATH` | Journal file to use (see [Your data](#your-data) for the default). |
| `--date YYYY-MM-DD` | Workday to act on. Defaults to today. Also sets the date printed in exports and backups. |
| `--theme dark\|light\|mono` | Color theme for the TUI. Defaults to `dark`. |
| `--ascii` | Replace decorative Unicode with ASCII. |

Running `devjournal` with no command opens the TUI.

| Command | Description |
| --- | --- |
| `add TEXT` / `add -` | Add a log entry to the workday (`-` reads from stdin). |
| `log` | Print the workday's entries. |
| `task TEXT` | Capture a task. Prints its ID. |
| `tasks` | List tasks with their IDs and state. |
| `plan TASK_ID` | Add a task to the workday's plan. |
| `unplan TASK_ID` | Remove a task from the workday's plan. |
| `planned` | Print the workday's plan. |
| `complete TASK_ID` | Mark a task done. |
| `topic TEXT` | Collect an O2O topic. |
| `topics` | List open O2O topics. |
| `meetings` | Print every O2O meeting record, newest first. |
| `export FILE.md` / `export -` | Export the whole journal as Markdown (to a file or stdout). |
| `backup FILE.json` / `backup -` | Write a complete JSON backup (to a file or stdout). |
| `restore FILE.json` / `restore -` | Restore a backup into an **empty** journal. |

Example session:

```console
$ devjournal task "Review deployment"
Task saved: 030fe407c3ee0b5208a67bb7724a3d05
$ devjournal plan 030fe407c3ee0b5208a67bb7724a3d05
Saved: plan
$ devjournal planned
Plan for 2026-09-27
030fe407c3ee0b5208a67bb7724a3d05  [open] Review deployment
```

## Keyboard reference

Press **?** inside the app for contextual help.

**Writing** (capture, correction, blockers, proposals, notes)

| Key | Action |
| --- | --- |
| Ctrl+S | Save |
| Enter | Insert a line break |
| Esc | Stop writing and browse (the draft is kept) |

**Browsing**

| Key | Action |
| --- | --- |
| n | Capture, or resume a kept draft |
| ↑/↓ or j/k | Select an entry |
| Enter | Read the full text |
| e | Correct the selected entry |
| ←/→ or h/l | Browse recorded workdays |
| t | Back to Today |
| Tab | Switch between Today and Tasks |
| b | Record a blocker |
| g | Open Daily |
| o | Open O2O |
| x | Discard a kept draft |
| ? | Help |
| q | Quit (only when no draft is pending) |

**Tasks**: **n** capture · **p** plan for today · **d** complete · **c** switch
open/completed · **Enter** read.

**Today's plan** (press **p** in Today): arrows select · **d** complete ·
**u** remove from plan · **p** back to the log.

**Daily**: **←/→** section · **↑/↓** row · **Enter** read · **s** toggle
share · **e** prepare/edit · **v** read proposal · **r** fresh proposal ·
**t** Today · **Tab** Tasks.

**O2O**: **n** add topic · **s** start meeting · **m** past meetings; during a
meeting **a** · **w** · **r** · **f** · **v** · **c** as described
[above](#run-a-one-to-one).

**Reading full text**: **↑/↓** and **PgUp/PgDn** scroll · **Esc** returns.

A kept draft must be saved or discarded before switching views.

## Your data

### Where it lives

By default the journal is stored in your OS user configuration directory:

| OS | Default path |
| --- | --- |
| macOS | `~/Library/Application Support/devjournal/journal.json` |
| Linux | `$XDG_CONFIG_HOME/devjournal/journal.json` (usually `~/.config/devjournal/journal.json`) |

Use `--data PATH` to keep separate journals, for example one for work and one
for personal projects.

### How it is protected

- **Atomic writes.** Each save writes a new file, syncs it, and replaces the old
  one, so a crash never leaves a half-written journal. The app reports success
  only after the write completes; a failed save keeps your text for a retry.
- **Concurrent use.** An advisory lock serializes writes from several open
  instances or scripts.
- **No silent resets.** Malformed files and unsupported format versions produce
  an error instead of being overwritten.

### Export to Markdown

```sh
devjournal export journal.md
devjournal export - | less
```

Produces a readable document: workdays newest first (entries, plan, blockers,
saved proposal), then open and completed tasks, open O2O topics, and every
meeting with its notes, agreements, and follow-ups. Your text is escaped where
needed so it cannot break the document structure. Exporting never changes the
journal and refuses to overwrite the journal file itself.

### Back up and restore

```sh
devjournal backup journal-backup.json
devjournal --data ~/restored/journal.json restore journal-backup.json
devjournal backup - | devjournal --data /tmp/copy.json restore -   # copy a journal
```

A backup contains every record, state, and relationship, so a restored journal
behaves exactly like the original, including a meeting still in progress.

`restore` only writes into an **empty** journal (a new file or one with no
records). It never merges or overwrites existing data. The whole backup is
validated first; malformed JSON, foreign files, unsupported versions, invalid
dates, duplicate IDs, or broken references are reported as clear errors and
leave the journal untouched.

## Terminal support

- Works in any modern terminal with an ordinary monospace font. No Nerd Fonts
  required.
- Minimum size is **80×24**. At **110 columns or more**, views gain a preview
  pane for the selected item. Below the minimum, a resize message appears and
  your draft is kept.
- Choose colors with `--theme dark` (default), `--theme light`, or `--theme mono`.
  Setting `NO_COLOR` to any non-empty value selects monochrome.
- `--ascii` removes decorative Unicode for limited terminals.
- The theme is chosen explicitly rather than by querying the terminal, so
  startup never waits on color detection.

Measured startup times and compatibility notes are in
[`docs/research/tui-validation.md`](docs/research/tui-validation.md).

## Development

The app is written in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea),
[Bubbles](https://github.com/charmbracelet/bubbles), and
[Lip Gloss](https://github.com/charmbracelet/lipgloss).

```
cmd/devjournal/   CLI entry point and TUI (Today, Tasks, Daily, O2O views)
journal/          Domain logic, storage, Markdown export, JSON backup/restore
scripts/          End-to-end terminal QA checks that drive the real binary in a PTY
docs/             Design notes and TUI research
CONTEXT.md        Domain glossary
```

Both the CLI and the TUI go through the same public `journal.Journal`
interface, and behavior tests exercise it against real temporary storage.

```sh
go vet ./...
go test ./...
go build -o bin/devjournal ./cmd/devjournal
```

The Python QA scripts (Python 3, no extra packages) run the built binary in a
pseudo-terminal and check multiline paste, correction, save and retry, resizing,
draft-safe exit, and startup time:

```sh
for check in terminal tasks daily o2o export backup; do
  python3 scripts/${check}_qa.py bin/devjournal
done
```

Further reading:

- [`docs/design.md`](docs/design.md): product agreements and design decisions
- [`docs/research/tui-design.md`](docs/research/tui-design.md): TUI design research
- [`docs/research/tui-validation.md`](docs/research/tui-validation.md): real-terminal validation results

Issues and pull requests are welcome. Please keep all repository content in
English and run the checks above before opening a pull request.

## Project status and limitations

Dev Journal is a personal tool in active use. The first version covers Today,
Tasks, Daily, O2O, Markdown export, and JSON backup. Current limitations:

- **macOS and Linux only.** File locking does not yet support Windows.
- **No sync.** The journal is a local file; use backups or your own file sync if
  you need it on several machines.
- **Date changes need a restart.** If you keep the app open past midnight,
  reopen it to switch to the new day.
- **No integrations yet.** A generated daily proposal from issue-tracker
  activity (for example, Linear) is a future idea; it would stay separate from
  the proposal you write yourself. Resources and training tracking are also
  planned for a later iteration.

## License

[MIT](LICENSE)
