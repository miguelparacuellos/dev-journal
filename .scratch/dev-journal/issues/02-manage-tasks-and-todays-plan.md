# 02: Manage tasks and today's plan

**What to build:** Keep a persistent Tasks list and deliberately select actions for today. Clearly separate intended actions from completed work, without automatically carrying unfinished actions into the next day.

**Blocked by:** 01: Capture and browse the daily log.

**Status:** ready-for-agent

**Completion:** implemented and reviewed on 2026-09-27

- [x] Users can create tasks without assigning a workday and view open tasks in Tasks.
- [x] Users can select an open task for today and see it in today’s plan, separately from daily log entries.
- [x] Selecting a task for today does not mark it completed.
- [x] Users can mark a task completed and distinguish completed tasks from open tasks.
- [x] An unfinished task remains open after the day changes but is not automatically added to the new day’s plan; the user can select it again.
- [x] Tasks, completion states, and dated plan selections survive reopening; the workflows are usable by keyboard with visible actions.
- [x] Behavior tests through the public application interface verify creation, planning, completion, persistence, and absence of automatic carryover across a date change.


## Comments

Implemented persistent undated Tasks and deliberate dated plan selections through
the shared public Journal interface. Tasks can be completed, selected again on a
later date while open, and removed from a plan without losing the task. Planning
never records an entry or completes a task. Historical selections persist and
show the task's current completion state.

Incremental red-green tests cover creation, dated planning, idempotent selection,
completion, reopening, invalid actions, deselection, and no automatic carryover.
Full Go tests and vet pass. Real OS PTY checks exercise Tasks capture, planning,
deselection, completion, date rollover through reopening, retained draft routing,
accumulated lists, and long full text resized from 120x40 to 80x24.

Standards review found no hard violations. Two small duplication smells were
removed; richer typed navigation remains a future judgement call. The review
flagged that the internal dated Task association is not separately named in the
glossary; Today's plan already defines its user-facing concept.
Spec review found two P2 issues: task detail wrapping on resize and accumulated
Tasks overflowing the compact layout. Both are fixed with PTY regressions;
a temporary render-height probe measured 23 lines at 80x24 with 12 tasks.

The existing native-emulator visual QA limitation remains: computer-use access
to native Terminal was rejected for safety reasons; no bypass was attempted.
The running session's date refreshes on reopening, as documented in README.
The configured triage vocabulary has no completed state; completion is recorded
separately above.
