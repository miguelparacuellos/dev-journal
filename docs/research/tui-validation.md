# Today design and terminal validation

Date: 2026-09-27. Implemented slice: ticket 01, daily log only.

## Framework decision

The intensive primary-source comparison is recorded in `tui-design.md`. Selected
Go 1.27 with Bubble Tea 2.0.10, Bubbles 2.2.1, and Lip Gloss 2.0.6 after building
and exercising the complete capture flow in an actual OS PTY. The first terminal
check exposed an initial-focus bug that was fixed before extending the design.

Charm offers measured fast startup, controlled cell rendering, Unicode-aware
multiline editing, viewport scrolling, and a single compiled binary. The public
Journal behavior boundary keeps workflow tests independent of its model/update
implementation. Textual would provide faster CSS-oriented composition but adds a
Python runtime and packaging decisions; Ratatui requires more editor assembly;
Ink offers React familiarity but requires additional editing composition. None
of the alternative frameworks was benchmarked; these tradeoffs are sourced
capabilities plus project-specific judgement, not comparative performance claims.

## Visual system and layout choice

Two treatments were implemented and exercised with the same content: a compact
writing layout at 80×24 and a spacious log-plus-preview layout at 120×40. The
compact version keeps the composer, selected row, durable-save status, and action
hints visible. The wide version uses spare horizontal space for a full-text
preview rather than empty future-feature panels. Enter always opens a scrollable
full entry. Longer lists are windowed around selection; long rows show ellipsis
and an explicit complete-entry action.

Hierarchy uses bold titles, one separator, four-cell margins, plain body text,
and restrained cool accents. Focus is labelled `[FOCUS]`, selection uses `>`,
success says `Saved`, and failures say `Not saved`. Color is supplementary.
Dark, light, and monochrome palettes are explicit overrides. ASCII mode replaces
decorative separators and middle-dot hints. No custom fonts, icons, animations,
or automatic network operations are needed.

The composer starts focused. Enter is consistently a newline, Ctrl+S saves,
Esc retains a recoverable session draft, and browsing q refuses to lose a draft.
This deliberately chooses one multiline editor over separate one-line and
expanded modes: it removes a hidden mode and makes pasted paragraphs predictable.
The save hint remains visible. No timing metadata is required.

## Recorded evidence

Environment: macOS arm64 development host; temporary Go 1.27.1 toolchain;
compiled binary; OS PTY with `TERM=xterm-256color`; no terminal library mocks.
Manual interactive PTY review and repeatable `scripts/terminal_qa.py` checks:

- Empty and populated Today at 80×24; capture/save/exit and reopen.
- Multiline bracketed paste including Spanish, CJK, and command-like characters.
- Correction from the 120×40 light-theme split layout.
- Draft retained across 60×18 and restored 80×24 resizing.
- Accidental q refused until explicit draft discard.
- Failed persistence keeps text; fixing the destination and retrying saves it.
- A 10,000-entry historical fixture still opens and captures responsively.
- ANSI screen state, cursor, and paste mode restored on normal exit.

Initial 30-sample measurement: median opening 68.60 ms, p95 71.02 ms;
multiline save-to-feedback 16.72 ms; opening with 10,000 entries 65.30 ms;
save with that fixture 33.33 ms. Opening is measured to the first frame containing
the capture hint, not process spawn alone. Save latency is key-to-feedback.
These are local measurements, not universal emulator guarantees. The repeated
PTY check reports current results when run; the one-second startup target passed.

## Supported behavior and limitations

macOS and Linux are the packaging targets. PTY output confirms compact/wide
layouts, fallback text, and keyboard workflows; it does not establish native
font rendering or visual contrast in every terminal. Native macOS Terminal,
the user's preferred emulator, VS Code's terminal, tmux, and screen readers
have not been individually reviewed. Root attempted native Terminal review through
the computer-use tool, which rejected access to `com.apple.Terminal` for safety
reasons. No bypass was attempted; native visual QA remains unverified. Light/dark palette checks here concern
rendered roles and text, not photometric contrast measurements. Color
adaptation is delegated to Charm, and monochrome works independently of it.

There is no screen-reader accessibility claim. Plain `log` and quick `add` are
available without full-screen rendering. Force termination can lose unsaved
session drafts. Saved data is retained. Startup is offline; installing build
dependencies needs network access once. Remaining emulator checks should be
performed on the user's actual terminal before describing universal visual
compatibility.

## Required two-axis code review

### Standards

No actionable findings. English prose, glossary vocabulary, and public-boundary
behavior tests conform to repository standards. UI mode strings and the rendering
method can evolve as real additional views arrive; no speculative abstraction was
requested for this single-view slice.

### Spec

One P2: the library's default 99-row limit silently blocked inserted newlines
when correcting an existing long entry. Removed maximum editor dimensions and
added a 100-line correction PTY regression. The regression passed. No scope
creep found. Native visual compatibility remains partial as recorded above.

Final post-fix PTY run: 30 samples; opening median 61.07 ms, p95 86.15 ms;
multiline save 16.63 ms; 10,000-entry opening 52.91 ms and save 33.92 ms.
Monochrome checks also reject emitted RGB color sequences. Directory metadata is
synced after replacement; failed writes restore the in-memory saved view.


## Ticket 02: Tasks and today's plan

Extended the same visual roles to undated Tasks, an explicit open/completed
filter, and a separate dated plan above the daily log. Selection remains `>`,
completion remains text `[done]`, and intended actions say `[open]`. Tab changes
Today/Tasks from browsing; retained drafts require saving or discarding before a
view change so text cannot be routed into the wrong kind of record.

`python3 scripts/tasks_qa.py bin/devjournal` passed with the actual built binary
at 80x24 monochrome/ASCII and 120x40 light/Unicode: task capture, planning,
deselection, completion, next-day empty plan, reopening, safe drafts, accumulated
Tasks/plans, and long task detail shrinking 120 to80 columns. A temporary render
probe measured 23 lines for an accumulated 12-task list at 80x24. The baseline
terminal check also passed: 30 launches, median 69.52ms, p95 71.36ms; save 16.66ms;
10,000-entry opening 52.27ms and save 33.34ms. Native terminal rendering remains
unverified under the limitation recorded above.

### Standards

No hard standard breaches. Two duplication smells (ID generation and CLI task
formatting) were removed. Typed navigation is a judgement call deferred until
additional views establish their needs. The dated Task association is represented
in storage; the existing Today's plan glossary concept describes its behavior.

### Spec

Two P2 findings fixed: task-detail text needed rewrapping on resize, and
accumulated Tasks needed a smaller compact row budget to retain visible actions.
Both have actual PTY regression evidence. Help was shortened to fit compact
screens. No scope creep or normal keyboard-path panic found.


## Ticket 03: Daily preparation

Daily extends the existing visual roles with three labelled sections: recent
work with its actual source date, today's intended actions, and today's blockers.
Today shows the blocker count and latest blocker above the log. `g` opens Daily;
`b` captures a blocker. Left/Right changes section and Up/Down changes selection;
Enter reads full text. `s` selects progress; `e` edits saved preparation (or
prepares the first selection); `r` explicitly starts fresh from selected sources.
The saved personal preparation is a dated independent record, never a source
mutation or a generated draft. `v` reads it in the scrollable full-text view.

Public Journal boundary tests were implemented incrementally with observed
red/green results for source dates, weekends, absences, future/today exclusion,
empty recent work, blocker persistence, selected progress copying, personal
editing, reopening, dated proposals and source preservation. A temporary render
probe with 12 entries, planned tasks and blockers measured 23 rows for Today,
Daily, Tasks and complete help at 80x24; the probe was removed after inspection.
The visible row, section focus, source date and footer remain understandable in
monochrome ASCII. Real OS PTY checks exercise compact/wide capture, editing,
selection, safe drafts, reopening, long-content resizing and accumulated sections.
Native emulator/font/contrast review remains unverified under the existing
computer-use safety limitation; this is PTY and cell-layout evidence only.
