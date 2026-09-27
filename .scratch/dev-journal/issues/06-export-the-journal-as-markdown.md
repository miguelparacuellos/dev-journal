# 06: Export the journal as Markdown

**What to build:** Export the journal as readable English Markdown for reading and use outside the application, including daily records and O2O history.

**Blocked by:** 03: Prepare and save a daily proposal; 05: Record O2O meetings and follow-up actions.

**Status:** ready-for-agent

- [x] Users can invoke a documented export action and save a Markdown document locally.
- [x] The export includes entries, dated plans, blockers, saved daily proposals, tasks and their states, O2O topics, meeting notes, agreements, and follow-up actions.
- [x] Clear headings, dates, and task or topic states make the content understandable without opening the application.
- [x] Meeting follow-up actions and daily proposal source dates remain understandable in the exported document.
- [x] Export works offline, handles an empty journal clearly, and leaves the stored journal unchanged.
- [x] Behavior tests verify readable representation of representative journal content and successful export through the public application interface without asserting internal structures.


## Comments

Implementation completed in `d59d8ee`, with review fixes in `43bc67f`, `391c9c8`
and `7661821` (review baseline `ea74da1`). The `journal.Journal` boundary gains
`Markdown(exportedOn)` and `ExportMarkdown(path, exportedOn)`. The document has
a title and export date, then the daily log by workday newest first (entries,
today's plan with each task's current state, blockers, and the saved daily
proposal as a block quote with its recent-work source date), tasks grouped open
and completed with their planned dates and O2O origin ("O2O follow-up from the
2026-09-10 meeting"), open O2O topics with their collection date, and every
meeting newest first with notes, addressed topics, agreements, and follow-up
tasks. Tasks read `[ ] Open:` or `[x] Done:` so the state is clear with or
without checkbox rendering. An empty journal produces a short document that says
so. User line breaks are kept as hard breaks; user lines that Markdown would
read as headings, quotes, fences, rules, tables, or HTML are escaped, including
after a list marker, and terminal control characters are dropped. Export writes
atomically, only reads the journal, and refuses the journal and lock paths.

Meeting record rendering moved from `cmd/devjournal/o2o.go` into the journal
package (`MeetingRecord.State`, `Text`, `Details`); the TUI, the `meetings`
command, and the export share its sections. `Task.State` gives the shared
open/done word. Daily proposals now store `source` (the recent-work workday when
saved, `""` when none) so later backdated entries cannot change it; proposals
saved before this change say the source was not recorded and show the current
reference.

The CLI gains `export FILE.md` and `export -` (standard output). The TUI help's
last line now reads "Saved locally after each write · Markdown: devjournal export
FILE.md" (68 columns; help still uses 23 of 24 rows). There is no TUI key for
export: it is a documented command, which fits the offline quick-capture
workflow. README documents the command.

Validation: red/green behavior tests through the public boundary with real
temporary storage cover the empty journal, representative content (every
listed kind, order, states, source date, follow-up origin, addressed topics
only with their meeting), leaving the stored journal byte-identical, replacing
an earlier export, refusing the journal path and a missing directory, a
meeting in progress, a missing source, a saved source surviving backdated
entries, legacy proposals, control characters, and Markdown-looking text. The
full Go suite, vet, gofmt and build passed. New `scripts/export_qa.py` exercises
the built binary: empty export without creating the journal, populated export
to a file and stdout, self-overwrite refusal, journal bytes unchanged, and the
help pointer in a real OS PTY at 80x24. Terminal, Tasks, Daily, O2O and Export
PTY checks passed (startup median 66.07 ms, p95 69.6 ms). A temporary render
probe measured the help at 80x24 within 24 rows and 76 columns including
padding. Native emulator, font and contrast review is still unverified.

Two-axis review:
- Standards: no hard violations. Judgement calls fixed: shared `lockPath`,
  temp-file pattern passed to the atomic writer, one `Task.State` word used by
  the TUI, CLI and record, a shared topic line helper, and setup errors checked
  in tests. The duplicated "None recorded." empty handling between plain and
  Markdown renderers and the mixed direct/accessor reads in the daily log were
  left as minor.
- Spec: a P1 (Markdown-looking user text could break structure, e.g. an
  unclosed fence swallowing a meeting) and a P2 (proposal source recomputed at
  export) were fixed; the recheck's P2 (text after a list marker) and P3s
  (legacy source wording, spaced rules) were fixed. Final recheck: no P1 or P2.
  Deferred: dangling ID references (an addressed topic whose meeting is missing,
  a plan selection whose task is missing) are silently omitted; they cannot be
  produced through the API and belong to ticket 07's restore validation. `export
  -` was kept as a small documented convenience.
