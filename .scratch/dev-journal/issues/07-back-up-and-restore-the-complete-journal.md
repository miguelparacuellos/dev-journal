# 07: Back up and restore the complete journal

**What to build:** Export a complete JSON backup and restore it into an empty installation so the user can recover the journal with its history, states, and relationships intact.

**Blocked by:** 03: Prepare and save a daily proposal; 05: Record O2O meetings and follow-up actions.

**Status:** ready-for-agent

- [x] Users can invoke documented local JSON backup and restore actions without a network connection.
- [x] The backup includes all entries, dates, plans, blockers, daily proposals, tasks, O2O topics, meetings, notes, agreements, actions, states, and their relationships.
- [x] Restoring into an empty installation reproduces the journal’s observable content and behavior, including historical daily preparation and O2O follow-up tasks.
- [x] Restored data survives closing and reopening the application.
- [x] Empty-journal backup and restore work; malformed or unsupported backup input produces a clear error without leaving partially restored data.
- [x] The supported restore behavior and any limits for non-empty installations are documented; unrelated existing data must not be silently overwritten.
- [x] A behavioral round-trip test exports representative data and restores it into fresh temporary storage, then verifies it through the public application interface.


## Comments

Implementation completed in `3879956`, with review fixes in `2ea37cb` (review
baseline `f10a228`). The `journal.Journal` boundary gains `Backup(backedUpOn)`,
`ExportBackup(path, backedUpOn)` and `Restore(backup)`. A backup is the stored
journal (version 1) with a `"format": "devjournal-backup"` marker and a backup
date, so absent and `""` proposal sources stay distinct. Restore decodes
strictly (unknown fields, trailing content and wrong types are rejected), then
checks the format, version, dates, non-empty text, unique IDs, every reference
(follow-up task, agreement and addressed topic to meeting; plan selection to
task), duplicate plan selections or daily proposals, and at most one meeting in
progress, all before taking the lock. It only fills an empty journal (checked
under the write lock); a journal with records is refused, never merged. Writes
go through `change()` and `writeFile`, so a failed save rolls the session back.
`ExportBackup` refuses the journal and lock paths. The CLI adds `backup
FILE.json`, `backup -`, `restore FILE.json` and `restore -`; the last TUI help
line now reads "Saved locally · devjournal export FILE.md · backup/restore
FILE.json" (68 columns, no new row). README documents the commands and the
empty-journal restore limit; CONTEXT.md gains **Backup**.

Validation: red/green behavior tests through the public boundary with real
temporary storage cover a representative round trip into fresh storage
(Markdown, tasks, meetings, topics, workdays, entries, plans, blockers, saved
proposals, historical daily preparation, identical re-backup, after reopening,
then continuing the restored meeting in progress and its follow-up task),
legacy and `""` proposal sources, an empty journal, refusal into a non-empty
journal, 9 malformed or unsupported inputs, 25 inconsistent backups, backup
file replacement and path refusals, an invalid backup date, and a failed
restore save that leaves the session empty and can be retried (verified to fail
when the rollback is disabled). The full Go suite, vet, gofmt and build passed.
New `scripts/backup_qa.py` exercises the built binary: empty and populated
backups to a file and stdout, restore from a file and stdin with CLI output
matching the original, non-empty and invalid restores refused without changes,
and the restored journal and help in a real OS PTY at 80x24. Terminal, Tasks,
Daily, O2O, Export and Backup PTY checks passed (startup median 59.17 ms, p95
69.05 ms); Daily did not flake. Native emulator, font and contrast review is
still unverified.

Two-axis review:
- Standards: no hard violations. Judgement calls fixed: shared helpers moved
  into `journal.go` with `isDate` reused by every date check, a shared
  rejected-backup test assertion, one optional-reference check, consistent
  `err` use in the CLI. Left: copied `until`/`cli` helpers across PTY scripts
  (existing pattern) and the similar `export`/`backup` CLI cases (two uses).
- Spec: no P1. P2 (tracker and ticket not yet updated) is addressed by this
  record; P3s fixed: `Backup` refuses dates `Restore` would reject, glossary
  term added. `backup -` / `restore -` kept as small documented conveniences.
