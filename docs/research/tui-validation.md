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


Baseline PTY rerun: 30 launches, median 61.28 ms, p95 71.16 ms; save 17.27 ms;
10,000-entry opening 50.89 ms and save 34.33 ms. Tasks PTY regressions also passed.

### Standards

The blocker glossary definition was missing and detail-mode checks repeated the
same three-way condition. Added the glossary term and a shared detail predicate.
No remaining documented standard breach was reported.

### Spec

One P1 fixed: recording a blocker from a focused Today plan could retain a log
index for subsequent task navigation and panic. Blocker capture now returns
focus to the log; the actual PTY checks exercise two log entries with a one-task
plan and reading after the save. An additional draft-state audit fixed an empty
blocker editor retaining its capture kind after Escape, with a Tasks routing
regression. Native visual validation remains partial as documented above.


## Ticket 04 — O2O topics

The O2O view (**o**) keeps the capture-first composer and lists open topics with
the day each was collected. At 110 columns or more, a word-wrapped
selected-topic preview is added. A temporary render probe measured 23 rows at
80x24 for the empty state, 3 topics, 30 topics (six visible rows plus a range
footer), and the complete help screen. The probe was removed. At 120x40 the list
and preview widths add up exactly to the content width.

`scripts/o2o_qa.py` exercises the built binary in a real OS PTY at 80x24
monochrome/ASCII and 120x40 light/Unicode. It covers the empty state, capture
without a meeting, rejection of an empty topic, bracketed multiline paste, draft
protection when switching views, blocker routing from O2O, full-text reading
after a resize, CLI capture on another day, review after reopening later in the
month, and a regression for an abandoned correction. Bubble Tea redraws only
changed cells, so the script waits on tokens that are emitted whole.

Baseline PTY rerun: 30 launches, median 77.68 ms, p95 82.40 ms; save 16.85 ms.
The Tasks and Daily PTY checks also passed. One Daily check run failed on a
token split by a cell-diff redraw and passed on three reruns. This fragility
predates ticket 04. Native emulator, font and contrast review is still
unverified.


## Ticket 05 — O2O meetings

During a meeting the O2O view keeps the composer. It lists every open topic and
every topic addressed in this meeting with `[open]` or `[addressed]` markers, in
capture order, so marking a topic never moves its row. Below the list, the
compact layout shows three summary lines for notes, agreements, and follow-ups.
At 110 columns or more the meeting record appears beside the list instead. The
meetings pane (`m`) lists meetings newest first, and at 110 columns or more it
previews the selected record. A temporary render probe measured 23 of 24 rows
and at most 80 columns at 80x24, and 39 rows and 120 columns at 120x40, for
30 topics, a meeting in progress with 40 lines of notes, help, the full record,
the meetings pane and the state after closing. The probe was removed. Adding the
meetings summary line reduced the no-meeting topic list at 80x24 from six
visible rows to five.

`scripts/o2o_qa.py` adds the complete meeting cycle in a real OS PTY at 80x24
mono/ASCII, followed by history browsing, record reading across a resize and a
second meeting at 120x40 light/Unicode. It checks the saved JSON relationships
and the `meetings` CLI output. The script polls the saved journal for state
changes whose redraws only touch a few cells.

Baseline PTY rerun: 30 launches, median 78.69 ms, p95 82.00 ms; save 16.39 ms.
The Tasks and Daily PTY checks also passed. One Daily run failed again on a
token split by a cell-diff redraw and passed on three reruns; this predates
ticket 05. Native emulator, font and contrast review is still unverified.


## Ticket 06 — Markdown export

Export is a CLI command (`export FILE.md`, `export -`); the TUI only gains a
pointer on the last help line, replacing "Saved locally after a durable write.
Esc closes help." (the Esc Back hint still shows how to leave help). A
temporary render probe measured the help within 24 rows and at most 80 columns
at 80x24, and within 40 rows and 120 columns at 120x40; it was removed.
`scripts/export_qa.py` checks the built binary and the help pointer in a real
OS PTY at 80x24 mono/ASCII. Baseline PTY rerun: 30 launches, median 66.07 ms,
p95 69.6 ms. The Tasks, Daily and O2O PTY checks also passed. The exported
Markdown was read raw; its rendering in a Markdown viewer was not visually
checked. Native emulator, font and contrast review is still unverified.


## Ticket 07 — JSON backup and restore

Backup and restore are CLI commands (`backup FILE.json`, `backup -`,
`restore FILE.json`, `restore -`); the TUI only changes its last help line to
"Saved locally · devjournal export FILE.md · backup/restore FILE.json". The
line keeps its 68-column width and no row was added, so the help layout
measured for ticket 06 is unchanged. `scripts/backup_qa.py` checks the built
binary, opens the restored journal and reads the help pointer in a real OS PTY
at 80x24 mono/ASCII. Baseline PTY rerun: 30 launches, median 59.17 ms, p95
69.05 ms; save 17.74 ms. The Tasks, Daily, O2O and Export PTY checks also
passed. Native emulator, font and contrast review is still unverified.
