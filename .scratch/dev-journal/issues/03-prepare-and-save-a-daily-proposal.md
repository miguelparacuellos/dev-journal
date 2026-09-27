# 03: Prepare and save a daily proposal

**What to build:** Record blockers and prepare a concise Daily update using recent work, today’s plan, and blockers. Select and edit what to share, then save the daily proposal independently from the original journal content.

**Blocked by:** 01: Capture and browse the daily log; 02: Manage tasks and today's plan.

**Status:** ready-for-agent

- [x] Users can record blockers and access them from Today and Daily; saved blockers survive reopening.
- [x] Daily displays recent work from the latest workday with entries before today and visibly identifies its source date.
- [x] Mondays and absences use the latest recorded workday rather than assuming yesterday; missing previous entries produce a clear empty state.
- [x] Recent work, today’s plan, and blockers appear as clearly separated sections.
- [x] Users can select progress to share and edit the daily proposal in their own words using keyboard controls.
- [x] The saved preparation survives reopening and does not modify the original entries, tasks, or blockers.
- [x] No Linear integration or automated generation is required; the preparation remains conceptually separate from any future generated draft.
- [ ] Behavior tests verify source-day selection, empty states, selecting and editing content, persistence, and preservation of source data; review the Daily view in a real terminal.


## Comments

Implementation completed with durable dated blockers and independent daily
preparation. Public Journal boundary red/green tests cover source-day selection,
weekends, absences, missing work, copying selected progress, editing, persistence,
and preservation of entries, tasks, plans and blockers. Full Go suite, vet and
build passed. Actual OS PTY checks passed at 80x24 monochrome/ASCII and 120x40
light/Unicode, including safe drafts, accumulated sections, long text resizing,
and the blocker-from-plan-focus regression. A temporary render probe measured
23 rows for populated Today, Daily, Tasks and complete help at 80x24.

The final validation item remains partial: behavior tests and real PTY review
passed, but native emulator/font/contrast review is unverified because the prior
computer-use attempt was rejected for safety. No bypass was attempted.

Two-axis review against 06eeb99: Standards identified a glossary gap and repeated
detail-mode checks; both fixed. Spec identified a blocker save from plan focus
that could leave an invalid task index; fixed with a PTY regression. The native
visual review limitation remains recorded. No future integration or automation
was added.
