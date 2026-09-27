# Implementation progress

Last updated: 2026-09-27. This is the durable handoff for resuming after a session or credit interruption. Check `git status` and `git log` before acting. All seven tickets are implemented and reviewed; native visual QA remains open.

## Working agreement

- Implement all seven approved tickets sequentially, using a fresh implementation subagent for each ticket.
- Each subagent must read and apply the `implement` skill: incremental TDD at the approved public seam, regular focused tests and compilation/static checks, full suite at the end, two-axis code review, fixes, and commits on the current branch.
- The approved seam is `journal.Journal`, shared by CLI and TUI, with real temporary local storage. Actual terminal keyboard interaction checks are also approved. No need to ask again about these seams.
- Code review compares the ticket's starting commit against its committed implementation. Run separate Standards and Spec reviewers. There are four concurrent agent slots including the root; completed reviewers may need to be released or reused for review if a spawn hits the limit. Implementation agents must still be fresh.
- Everything in the repository and application UI is English. Visual design and developer experience are core requirements.
- Read the specification and ticket before implementation. Do not reimplement completed tickets merely because their triage `Status` still says `ready-for-agent`; use completion metadata and this tracker for implementation lifecycle.

## Ticket tracker

| Ticket | Implementation state | Commits / next action |
| --- | --- | --- |
| [01 — Daily log](../.scratch/dev-journal/issues/01-capture-and-browse-the-daily-log.md) | Implemented and reviewed; native visual QA pending | `8484f23`, `947cac7`. Working capture, history, correction, CLI, local persistence, compact/wide layouts. |
| [02 — Tasks and today's plan](../.scratch/dev-journal/issues/02-manage-tasks-and-todays-plan.md) | Implemented and reviewed | `fedc8f6`, `25efd5d`, `06eeb99`. Task capture, dated selection/removal, completion, no automatic carryover. |
| [03 — Daily proposal](../.scratch/dev-journal/issues/03-prepare-and-save-a-daily-proposal.md) | Implemented and reviewed; native visual QA pending | `c179593`, `8dce835`; review baseline `06eeb99`. |
| [04 — O2O topics](../.scratch/dev-journal/issues/04-collect-and-review-o2o-topics.md) | Implemented and reviewed | `3aa2880`, `76597b4`; review baseline `387ba08`. Persistent open topics, `topic`/`topics` CLI, O2O view (`o`), `scripts/o2o_qa.py`. |
| [05 — O2O meetings and actions](../.scratch/dev-journal/issues/05-record-o2o-meetings-and-follow-up-actions.md) | Implemented and reviewed | `66bf651`, `1a48be4`, `8f749f7`; review baseline `c3613f3`. Meetings, addressed topics, agreements, follow-up tasks, history pane (`m`), `meetings` CLI, meeting cycle in `scripts/o2o_qa.py`. |
| [06 — Markdown export](../.scratch/dev-journal/issues/06-export-the-journal-as-markdown.md) | Implemented and reviewed | `d59d8ee`, `43bc67f`, `391c9c8`, `7661821`; review baseline `ea74da1`. `Journal.Markdown`/`ExportMarkdown`, `export FILE.md` / `export -` CLI, help pointer, `scripts/export_qa.py`. |
| [07 — JSON backup and restore](../.scratch/dev-journal/issues/07-back-up-and-restore-the-complete-journal.md) | Implemented and reviewed | `3879956`, `2ea37cb`; review baseline `f10a228`. `Journal.Backup`/`ExportBackup`/`Restore`, `backup`/`restore` CLI (files or pipes), help pointer, `scripts/backup_qa.py`. |

## Exact current handoff: all seven tickets implemented

A fresh implementation agent completed ticket 07, the last ticket. `3879956`
adds `Journal.Backup(backedUpOn)`, `Journal.ExportBackup(path, backedUpOn)` and
`Journal.Restore(backup)`. The backup is the version 1 stored journal plus a
`devjournal-backup` format marker and date; it keeps absent versus `""`
proposal sources. Restore decodes strictly, validates dates, text, unique IDs,
every ID reference and at most one meeting in progress before touching
anything, and only fills an empty journal (a journal with records is refused,
never merged). The CLI has `backup FILE.json`, `backup -`, `restore FILE.json`
and `restore -`; the TUI help's last line names them. `2ea37cb` fixes review
findings (shared helpers in `journal.go`, `isDate` everywhere, backup date
validation, glossary term **Backup**).

Validation: the full Go suite, vet, gofmt and build passed. Terminal, Tasks,
Daily, O2O, Export and Backup PTY checks passed (startup median 59.17 ms, p95
69.05 ms). Review against `f10a228`: Standards had no hard violations
(judgement calls fixed); Spec had no P1, and its P2 (tracker update) and P3s
were addressed. All seven acceptance items are checked.

Remaining open items (no implementation ticket is pending):

- Native visual QA: macOS Terminal, the preferred emulator, VS Code terminal,
  tmux, fonts, contrast on light and dark backgrounds, and screen-reader checks
  remain unverified. Computer-use of macOS Terminal was rejected; do not claim
  it.
- Markdown export rendering in a Markdown viewer was not visually checked.
- `scripts/daily_qa.py` has a known intermittent cell-diff flake (it passed
  this time); rerun it when it fails.
- Triage `Status:` lines in ticket files still read `ready-for-agent`; use the
  completion comments and this tracker for implementation state.
- Windows is not supported (advisory locking targets macOS and Linux).

## Runtime and commands

The project uses Go with Charm v2 (Bubble Tea, Bubbles, Lip Gloss) and local versioned JSON with atomic replacement and advisory write locking. The current implementation targets macOS and Linux. The branch is `main`; no remote has been configured.

A temporary toolchain is available at `/tmp/go/bin/go`, with GOROOT `/tmp/go` and GOPATH `/Users/miguelparacuellos/go`. Temporary paths may disappear between sessions: if missing, install a supported official Go toolchain. Go cache and Git writes required sandbox escalation in this session; commits are explicitly authorized by the user.

```sh
/tmp/go/bin/go test ./journal -run TestDaily
/tmp/go/bin/go vet ./...
/tmp/go/bin/go test ./...
/tmp/go/bin/go build -o /tmp/devjournal ./cmd/devjournal
python3 scripts/terminal_qa.py /tmp/devjournal
python3 scripts/tasks_qa.py /tmp/devjournal
python3 scripts/daily_qa.py /tmp/devjournal
python3 scripts/o2o_qa.py /tmp/devjournal
python3 scripts/export_qa.py /tmp/devjournal
python3 scripts/backup_qa.py /tmp/devjournal
```

Use an explicit temporary `--data` path when trying the application to avoid mixing test fixtures with the user's journal. See `README.md` for current commands and shortcuts.

## Design and validation context

- [Specification](../.scratch/dev-journal/spec.md)
- [TUI research](research/tui-design.md): primary-source framework and interaction investigation.
- [Terminal validation](research/tui-validation.md): ticket 01 evidence and limitations. Later ticket evidence is also recorded in ticket files and validation documents.
- [Glossary](../CONTEXT.md)

Real OS PTY checks cover capture, multiline paste, save/retry, correction, draft protection, resize, tasks, plans, and daily preparation. They do not establish native font rendering or visual contrast in every emulator. Computer-use access to macOS Terminal was explicitly rejected for safety reasons; do not bypass that rejection or claim native visual QA completed. macOS Terminal, the preferred emulator, VS Code terminal, tmux, and screen-reader visual checks remain unverified. This does not technically block subsequent behavior implementation.

Do not add Linear integration, scheduled skills, resources/training, synchronization, or multiple spaces: they are outside the current scope.
