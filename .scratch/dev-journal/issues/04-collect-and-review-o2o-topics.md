# 04: Collect and review O2O topics

**What to build:** Capture O2O topics throughout the month and review the persistent collection of open topics before the next meeting.

**Blocked by:** 01: Capture and browse the daily log.

**Status:** ready-for-agent

- [x] Users can add free-text O2O topics without starting a meeting.
- [x] The O2O view clearly presents open topics for meeting preparation and has a helpful empty state.
- [x] Open topics remain available across days and application restarts.
- [x] Topic capture and review are usable by keyboard with visible actions and English UI text.
- [x] Behavior tests through the public application interface verify topic capture, review, and persistence independently of meeting recording.


## Comments

Implementation completed in `3aa2880`, with review fixes in `76597b4`
(review baseline `387ba08`). The `journal.Journal` boundary gains persistent
O2O topics (`AddTopic`, `OpenTopics`) that are collected on any day without a
meeting. Quick capture adds `topic TEXT` and `topics`. The TUI adds an O2O view
(**o** from browsing) with topic capture, an open-topic list showing the day
each topic was collected, full-text reading, a selected-topic preview at 110
columns or more, and an empty state that explains topics stay open until they
are addressed.

Red/green behavior tests through the public boundary with real temporary
storage cover capture across days, persistence after reopening, validation,
concurrent sessions, and separation from entries, tasks, and blockers. None of
them involve meeting recording. The full Go suite, vet, gofmt and build passed.
`scripts/o2o_qa.py` passed in a real OS PTY at 80x24 monochrome/ASCII and
120x40 light/Unicode: empty state, keyboard and multiline capture, validation,
draft safety, blocker routing, review, long full text across a resize,
reopening on a later day, and correction routing. The terminal, Tasks and Daily
PTY checks also passed. A temporary render probe measured 23 of 24 rows for
O2O (empty, 3 topics, 30 topics) and the complete help screen at 80x24. The
probe was removed afterwards.

Native emulator, font and contrast review is still unverified. The earlier
computer-use request was rejected and nothing was done to bypass it. This
ticket's acceptance items do not depend on that review.

Two-axis review:
- Standards: the only hard finding was the stale tracker, which is now updated.
  Two judgement calls were fixed: the wide-layout breakpoint now has a name,
  and the O2O key blocklist became an allowlist of the keys shared with other
  views. Pre-existing duplication (date/text validation, PTY helper copies,
  string modes) was noted and left as it is.
- Spec: no P1 findings. One P2 was fixed: emptying a correction draft and
  pressing Esc kept the correction target, so O2O or Tasks capture then showed
  "Correct entry". A PTY regression covers it and fails without the fix. For
  the P3 hint trade-off, full hint lines are now shown whenever they fit, and
  the 80-column lines keep view navigation and rely on **?** help. The
  `topic`/`topics` CLI commands are kept as a quick-capture complement.
