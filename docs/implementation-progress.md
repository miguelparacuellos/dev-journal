# Implementation progress

Last updated: 2026-09-27. This is the durable handoff for resuming after a session or credit interruption. Check `git status` and `git log` before acting. Tickets 01–04 are implemented and reviewed; the next implementation ticket is 05.

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
| [05 — O2O meetings and actions](../.scratch/dev-journal/issues/05-record-o2o-meetings-and-follow-up-actions.md) | Not started | Unblocked (02 and 04 done). Next: fresh subagent, current HEAD as review baseline. |
| [06 — Markdown export](../.scratch/dev-journal/issues/06-export-the-journal-as-markdown.md) | Not started | Blocked by 03 and 05. |
| [07 — JSON backup and restore](../.scratch/dev-journal/issues/07-back-up-and-restore-the-complete-journal.md) | Not started | Blocked by 03 and 05. |

## Exact current handoff: start ticket 05

A fresh implementation agent completed ticket 04. `3aa2880` adds `journal.Topic`
with `Journal.AddTopic(day, text)` and `Journal.OpenTopics()`, stored under
`topics` in the version 1 JSON. It also adds the `topic TEXT` and `topics` CLI
commands, the O2O TUI view in `cmd/devjournal/o2o.go` (`o` opens it, with
capture, an open-topic list with collected dates, full-text reading, a preview
at 110+ columns, and an empty state), behavior tests, README, and
`scripts/o2o_qa.py`. `76597b4` fixes the review findings.

Validation: the full Go suite, vet, gofmt and build passed. The terminal, Tasks,
Daily and O2O PTY checks passed. The startup median was 77.68 ms and p95
82.40 ms. The render probe measured 23 rows at 80x24 for O2O and help. One Daily
PTY run flaked on a token split by a cell-diff redraw; it passed on three reruns
and predates ticket 04.

Review against `387ba08`:

- Standards: the only hard finding was the stale tracker, now updated. The
  wide-layout breakpoint now has a name (`wideLayoutWidth`), and the O2O key
  blocklist became an allowlist. Pre-existing duplication (validation helpers,
  PTY helper copies, string section/mode chains) was left as it is.
- Spec: no P1 findings. The P2 was fixed with a PTY regression: Esc on an
  emptied correction kept `editingID`, so later O2O or Tasks capture showed the
  correction focus. For the P3 hint widths, full hints now show whenever they
  fit (`fitHints`), and the 80-column lines rely on `?` help. The reviewer's
  recheck found nothing further.
- All five ticket 04 acceptance items are checked. Native emulator, font and
  contrast review is still unverified, as for earlier tickets.

Notes for ticket 05:

- `Topic` has no addressed state yet, and `OpenTopics()` returns every topic.
  Add a field (for example an addressed marker or meeting ID) with `omitempty`,
  make `OpenTopics()` filter on it, and remember that `change()` snapshots
  every slice for rollback, so new slices must be added there too.
- In `o2o.go`, `updateTopics` consumes every key except `n`, `x`, `?`, `q`,
  `ctrl+c` and `esc`. Add ticket 05's meeting keys there. `g`, `b`, `o`, Tab
  and `t` are handled earlier in `Update`.
- `save()` routes by `captureKind`, then section, then `editingID`. New capture
  kinds (meeting notes, agreements, follow-up tasks) need explicit routing and
  focus labels. Keep hint lines within 72 columns at 80x24, or pass a compact
  variant to `fitHints`.

Next action: spawn a fresh implementation subagent for ticket 05, using the
current HEAD as its fixed code-review baseline. Then continue sequentially
through 06 and 07.

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
```

Use an explicit temporary `--data` path when trying the application to avoid mixing test fixtures with the user's journal. See `README.md` for current commands and shortcuts.

## Design and validation context

- [Specification](../.scratch/dev-journal/spec.md)
- [TUI research](research/tui-design.md): primary-source framework and interaction investigation.
- [Terminal validation](research/tui-validation.md): ticket 01 evidence and limitations. Later ticket evidence is also recorded in ticket files and validation documents.
- [Glossary](../CONTEXT.md)

Real OS PTY checks cover capture, multiline paste, save/retry, correction, draft protection, resize, tasks, plans, and daily preparation. They do not establish native font rendering or visual contrast in every emulator. Computer-use access to macOS Terminal was explicitly rejected for safety reasons; do not bypass that rejection or claim native visual QA completed. macOS Terminal, the preferred emulator, VS Code terminal, tmux, and screen-reader visual checks remain unverified. This does not technically block subsequent behavior implementation.

Do not add Linear integration, scheduled skills, resources/training, synchronization, or multiple spaces: they are outside the current scope.
