# Implementation progress

Last updated: 2026-09-27. This is the durable handoff for resuming after a session or credit interruption. Check `git status` and `git log` before acting. Tickets 01–03 are implemented and reviewed; the next implementation ticket is 04.

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
| [04 — O2O topics](../.scratch/dev-journal/issues/04-collect-and-review-o2o-topics.md) | Not started | Start after ticket 03 is finalized, with a fresh subagent and current HEAD as review baseline. |
| [05 — O2O meetings and actions](../.scratch/dev-journal/issues/05-record-o2o-meetings-and-follow-up-actions.md) | Not started | Blocked by 02 and 04. |
| [06 — Markdown export](../.scratch/dev-journal/issues/06-export-the-journal-as-markdown.md) | Not started | Blocked by 03 and 05. |
| [07 — JSON backup and restore](../.scratch/dev-journal/issues/07-back-up-and-restore-the-complete-journal.md) | Not started | Blocked by 03 and 05. |

## Exact current handoff: start ticket 04

The fresh agent `implement_ticket_03` implemented blockers, latest recorded workday selection, selectable Daily material, and durable daily preparation independent of original entries, plans, and blockers. TUI, CLI, README, behavior tests, and `scripts/daily_qa.py` are included in `c179593`.

Before review, the full Go suite, vet/build, Daily and Tasks PTY checks, and baseline terminal QA passed. Accumulated compact layouts were checked at 23 rendered rows. Startup median was 61.28 ms and p95 71.16 ms on this machine.

Two-axis review found:

- Standards: a missing blocker glossary term and duplicated detail-mode predicates; both fixed.
- Spec: P1 — starting blocker capture while Today plan had focus could reuse a log-derived selection index and panic. The agent cleared plan focus when starting blocker capture and added a PTY regression. The reviewer rechecked the fix and reported no remaining behavioral findings.
- Native emulator/font/contrast review is not completed; the corresponding acceptance checkbox must stay honest.

Ticket 03 is finalized in `8dce835`: full Go suite, vet/build, and Daily/Tasks/baseline PTY checks passed. Seven acceptance items are checked; native visual validation remains explicitly partial. The working tree was clean after the ticket completion and tracker updates.

Next action: spawn a fresh implementation subagent for ticket 04, using the current HEAD as its fixed code-review baseline. Read the ticket and existing TUI before adding O2O topic capture and review. Do not restart tickets 01–03. Continue sequentially through 05, 06, and 07, updating this tracker and committing each ticket after review. No ticket 04 implementation agent has been started yet.

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
