# Implementation progress

Last updated: 2026-09-27. This is the durable handoff for resuming after a session or credit interruption. Check `git status` and `git log` before acting. Tickets 01–05 are implemented and reviewed; the next implementation ticket is 06.

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
| [06 — Markdown export](../.scratch/dev-journal/issues/06-export-the-journal-as-markdown.md) | Not started | Unblocked (03 and 05 done). Next: fresh subagent, current HEAD as review baseline. |
| [07 — JSON backup and restore](../.scratch/dev-journal/issues/07-back-up-and-restore-the-complete-journal.md) | Not started | Unblocked (03 and 05 done); run after 06. |

## Exact current handoff: start ticket 06

A fresh implementation agent completed ticket 05. `66bf651` adds O2O meetings
to `journal.Journal`: `StartMeeting`, `CurrentMeeting`, `SaveMeetingNotes`,
`AddressTopic`, `RecordAgreement`, `CreateFollowUpTask`, `CloseMeeting`,
`Agenda` and `Meetings`, which return `MeetingRecord` values (meeting, addressed
topics, agreements, follow-up tasks). The O2O view gains the meeting workflow
(`s` start, `a` addressed, `w` notes, `r` agreement, `f` follow-up, `v`
record, `c` then `c` close, `m` meetings pane). Follow-up tasks are shown with
an `O2O:` marker in Tasks and Today's plan. There is a `meetings` CLI command.
README, glossary (addressed topic, follow-up task) and `scripts/o2o_qa.py` were
updated. `1a48be4` and `8f749f7` fix the review findings.

Validation: the full Go suite, vet, gofmt and build passed. The terminal, Tasks,
Daily and O2O PTY checks passed; the O2O check passed three runs in a row. The
startup median was 78.69 ms and p95 82.00 ms. A temporary render probe measured
23 of 24 rows and at most 80 columns at 80x24, and 39 rows and 120 columns at
120x40, for every new O2O state and help. One Daily run flaked again on a
cell-diff split token ("Accumulated blocker 11") and passed on three reruns.
This predates ticket 05.

Review against `c3613f3`: Standards found no hard violations. The judgement
calls it raised were fixed: one table for the meeting capture kinds, shared
list windowing and preview helpers, named layout rows, and no more middle man.
Spec raised two P2s, both fixed: `v` was not discoverable at 80x24, and reading
notes with `w` then pressing Esc left a draft that blocked closing. Three P3s
were fixed. The `meetings` CLI was kept as a listing that complements `topics`.
The Spec recheck found no P1 or P2; its hint and footer P3s were fixed in `8f749f7`.
All eight acceptance items are checked. Native emulator, font and contrast
review is still unverified.

Notes for tickets 06 and 07:

- Stored data (version 1) now also holds `meetings` (`id`, `day`, `notes`,
  `closed_on`), `agreements` (`id`, `meeting_id`, `text`),
  `topics[].addressed_in` (a meeting ID) and `tasks[].meeting_id`. All of them
  use `omitempty`. Relationships are by ID only: a follow-up task and its
  meeting, an agreement and its meeting, an addressed topic and the meeting
  that addressed it, and a plan selection and its task.
- For Markdown export (06), `Journal.Meetings()` returns every meeting newest
  first as a `MeetingRecord`. `meetingRecordText` in `cmd/devjournal/o2o.go`
  already renders a readable plain-text record, and the CLI `meetings` uses it.
  Reuse or adapt it rather than writing a second renderer. `OpenTopics()`
  excludes addressed topics, so export the addressed ones through the meeting
  records. Tasks keep their current completion state, and follow-ups carry
  `MeetingID`.
- For JSON backup and restore (07), a restore must keep these ID references
  consistent and must preserve at most one meeting without `closed_on` (only
  one can be in progress). Validate references on restore rather than trusting
  them. `change()` snapshots every slice for rollback. Add any new slice there
  and extend `TestFailedMeetingSavesLeaveTheSessionUnchangedAndCanBeRetried`,
  which fails when a snapshot line is missing.
- PTY scripts: Bubble Tea redraws only changed cells, so wait for tokens that
  are emitted whole (unique status suffixes). Consecutive `until` calls cannot
  match the same frame. `saved_until` in `o2o_qa.py` polls the saved JSON for
  state checks.
- Hint lines must fit within 72 columns at 80x24; use `fitHints` with a compact
  variant. The help screen uses 23 of 24 rows at 80x24, so there is no room
  for a new line. Change an existing line instead.

Next action: spawn a fresh implementation subagent for ticket 06, using the
current HEAD as its fixed code-review baseline. Then continue with 07.

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
```

Use an explicit temporary `--data` path when trying the application to avoid mixing test fixtures with the user's journal. See `README.md` for current commands and shortcuts.

## Design and validation context

- [Specification](../.scratch/dev-journal/spec.md)
- [TUI research](research/tui-design.md): primary-source framework and interaction investigation.
- [Terminal validation](research/tui-validation.md): ticket 01 evidence and limitations. Later ticket evidence is also recorded in ticket files and validation documents.
- [Glossary](../CONTEXT.md)

Real OS PTY checks cover capture, multiline paste, save/retry, correction, draft protection, resize, tasks, plans, and daily preparation. They do not establish native font rendering or visual contrast in every emulator. Computer-use access to macOS Terminal was explicitly rejected for safety reasons; do not bypass that rejection or claim native visual QA completed. macOS Terminal, the preferred emulator, VS Code terminal, tmux, and screen-reader visual checks remain unverified. This does not technically block subsequent behavior implementation.

Do not add Linear integration, scheduled skills, resources/training, synchronization, or multiple spaces: they are outside the current scope.
