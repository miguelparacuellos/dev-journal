# Dev Journal

A calm, local terminal journal for brief visits throughout the workday.

## Run

Install Go 1.27 or newer, then:

```sh
go build -o bin/devjournal ./cmd/devjournal
./bin/devjournal
```

The TUI opens directly in Today with the composer focused. Type immediately,
use Enter for line breaks, then **Ctrl+S** to save. **Esc** pauses capture and
retains the draft; **q** quits from browsing when no draft remains.

```sh
./bin/devjournal add "Fixed login and added regression coverage"
printf 'Fixed login\nAdded coverage' | ./bin/devjournal add -
./bin/devjournal --date 2026-09-25 add "Reviewed deployment"
./bin/devjournal --date 2026-09-25 log
```

Options precede commands. `--data /path/to/journal.json` selects a local file.
The default is the operating system's user configuration directory followed by
`devjournal/journal.json` (`~/Library/Application Support/devjournal/` on macOS).
No network connection is used by the application.

## Keyboard

| Context | Action |
| --- | --- |
| Capture or correction | Ctrl+S saves; Enter inserts a newline; Esc retains the draft and browses |
| Browsing | n captures/resumes; arrows or j/k select; Enter reads full text; e corrects |
| Browsing | Left/Right or h/l browse recorded workdays; t returns to Today |
| Browsing | ? opens help; x explicitly discards a retained draft; q quits |
| Full entry | Up/Down and PgUp/PgDown scroll; Esc returns |

Corrections preserve the original workday. Saving reports success only after the
write completes. Failed saves retain the composer for retry. Exiting with a draft
requires saving it or explicitly discarding it; drafts are session-local, so
force-killing the process can lose unsaved text.

The compact layout supports **80×24**. At **110 columns or more**, the log gains
a selected-entry preview. Below 80×24 a resize message preserves the draft.
Use `--theme light`, `--theme dark` (default), or `--theme mono`; a nonempty
`NO_COLOR` selects monochrome. `--ascii` removes decorative Unicode. An ordinary
monospace font works without Nerd Fonts. Theme selection is explicit so terminal
color queries never delay writing.

Only the daily log is implemented in this slice. Daily, Tasks, O2O, and exports
are separate approved tickets.

## Development

```sh
go test ./journal -run TestCaptureSurvivesReopening
go vet ./...
go test ./...
go build -o bin/devjournal ./cmd/devjournal
python3 scripts/terminal_qa.py bin/devjournal
```

Behavior tests use the public `journal.Journal` interface shared by both UI
adapters, with real temporary storage. The PTY check exercises the built binary,
bracketed multiline paste, correction, save/retry, resizing, draft-safe exit, and
30 startup samples. See the design validation document for measured results and
terminal compatibility limits.

Storage uses versioned JSON with atomic replacement, file syncing, and an
advisory lock for concurrent writes. Unsupported versions and malformed files
fail visibly instead of resetting data. Locking currently targets macOS and
Linux; Windows packaging is not supported in this slice.
