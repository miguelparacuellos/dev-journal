# Dev Journal

A calm, local terminal journal for brief visits throughout the workday.

## Run

Install Go 1.27 or newer, then:

```sh
go build -o bin/devjournal ./cmd/devjournal
./bin/devjournal
```

The TUI opens directly in Today with the composer focused. Type immediately,
use Enter for line breaks, then **Ctrl+S** to save. **Esc** pauses capture and
retains the draft; **q** quits from browsing when no draft remains.

```sh
./bin/devjournal add "Fixed login and added regression coverage"
printf 'Fixed login\nAdded coverage' | ./bin/devjournal add -
./bin/devjournal --date 2026-09-25 add "Reviewed deployment"
./bin/devjournal --date 2026-09-25 log
```

Options precede commands. `--data /path/to/journal.json` selects a local file.
The default is the operating system's user configuration directory followed by
`devjournal/journal.json` (`~/Library/Application Support/devjournal/` on macOS).
No network connection is used by the application.

## Keyboard

| Context | Action |
| --- | --- |
| Capture or correction | Ctrl+S saves; Enter inserts a newline; Esc retains the draft and browses |
| Browsing | n captures/resumes; arrows or j/k select; Enter reads full text; e corrects |
| Browsing | Left/Right or h/l browse recorded workdays; t returns to Today |
| Browsing | g opens Daily; o opens O2O; Tab switches Today/Tasks |
| O2O | n adds a topic; s starts a meeting; m browses meetings |
| O2O meeting | a addressed; w notes; r agreement; f follow-up; v record; c close |
| Browsing | ? opens help; x explicitly discards a retained draft; q quits |
| Full entry | Up/Down and PgUp/PgDown scroll; Esc returns |

Corrections preserve the original workday. Saving reports success only after the
write completes. Failed saves retain the composer for retry. Exiting with a draft
requires saving it or explicitly discarding it; drafts are session-local, so
force-killing the process can lose unsaved text.

The compact layout supports **80×24**. At **110 columns or more**, the log gains
a selected-entry preview. Below 80×24 a resize message preserves the draft.
Use `--theme light`, `--theme dark` (default), or `--theme mono`; a nonempty
`NO_COLOR` selects monochrome. `--ascii` removes decorative Unicode. An ordinary
monospace font works without Nerd Fonts. Theme selection is explicit so terminal
color queries never delay writing.

Today includes a dated plan above the daily log. From browsing, **Tab** switches
between Today and Tasks. In Tasks, **n** captures an undated task, **p** selects
an open task for today, **d** completes it, and **c** switches open/completed
lists. **Enter** reads full task text. In Today, **p** focuses the plan; arrows
select, **d** completes, **u** removes a selection, and **p** returns to the log.
A retained draft must be saved or discarded before switching views.

Planning never completes a task or creates a daily log entry. Unfinished tasks
remain open across dates; each new day's plan starts empty. Reopen to refresh
the current date after midnight. Saved historical plans retain their selections
and show the task's current completion state. Exports remain separate approved
tickets.

```sh
./bin/devjournal task "Review deployment"
./bin/devjournal tasks
./bin/devjournal plan TASK_ID
./bin/devjournal planned
./bin/devjournal unplan TASK_ID
./bin/devjournal complete TASK_ID
```

Task IDs are printed on capture and listing. Use `--date YYYY-MM-DD` before a
planning command to inspect or deliberately select a different day's plan.

From browsing, **b** records a dated blocker visible in Today and Daily, and
**g** opens Daily. Daily identifies the latest recorded workday before today,
even after a weekend or absence; a missing workday has a clear empty state.
Recent work, today's plan, and blockers have separate sections. **Left/Right**
selects a section, **Up/Down** selects a row, and **Enter** reads its full text.
In recent work, **s** toggles progress to share. **e** opens saved personal
preparation, or prepares selected progress, the plan, and blockers when nothing
has been saved. Edit freely and **Ctrl+S** saves. **v** reads saved preparation.
**r** explicitly starts a fresh preparation from the current selection; saving
it replaces that day's preparation. **t** returns to Today; **Tab** opens Tasks.

Preparation is stored independently for each day and never modifies source
entries, tasks, or blockers. Saved words stay unchanged when sources change.
Selections before preparing are session-local; the saved proposal preserves the
chosen text. Draft protection and resize behavior apply to blockers and proposals.
There is no generated draft, Linear connection, or automation in this slice.

From browsing, **o** opens O2O. Collect topics for the next one to one whenever
they come up; no meeting is needed. **n** captures a topic, **Ctrl+S** saves it,
arrows select, and **Enter** reads its full text. Open topics are listed in
capture order with the day they were collected, stay open across days and
restarts, and gain a selected-topic preview at 110 columns or more.

When you meet, **s** starts an O2O meeting. The topic list then shows every open
topic with its state: **a** marks the selected topic `[addressed]`, and pressing
it again reopens a topic marked by mistake. **w** writes the meeting notes (it
reopens the saved notes for editing), **r** records an O2O agreement, and **f**
creates a follow-up task. Each opens the composer; **Ctrl+S** saves it into the
meeting. **v** reads the complete meeting record. Follow-up tasks appear in
Tasks marked `O2O:` and are planned and completed like any other task.

**c** closes the meeting and asks for a second **c** to confirm; a retained
draft must be saved or discarded first. Closing keeps the notes, addressed
topics, agreements, and follow-up tasks unchanged, and every unaddressed topic
stays open for the next meeting. Only one meeting is in progress at a time.
**m** browses all meetings, newest first; **Enter** reads a meeting's record
and **Esc** returns to the topics. At 110 columns or more the meeting record
and the selected past meeting are shown beside the lists.

```sh
./bin/devjournal topic "Feedback on the incident review"
./bin/devjournal topics
./bin/devjournal meetings
```

`topics` lists only open topics. `meetings` prints each meeting record, newest
first, with follow-up tasks showing their current completion state.

## Development

```sh
go test ./journal -run TestCaptureSurvivesReopening
go vet ./...
go test ./...
go build -o bin/devjournal ./cmd/devjournal
python3 scripts/terminal_qa.py bin/devjournal
python3 scripts/tasks_qa.py bin/devjournal
python3 scripts/daily_qa.py bin/devjournal
python3 scripts/o2o_qa.py bin/devjournal
```

Behavior tests use the public `journal.Journal` interface shared by both UI
adapters, with real temporary storage. The PTY check exercises the built binary,
bracketed multiline paste, correction, save/retry, resizing, draft-safe exit, and
30 startup samples. See the design validation document for measured results and
terminal compatibility limits.

Storage uses versioned JSON with atomic replacement, file syncing, and an
advisory lock for concurrent writes. Unsupported versions and malformed files
fail visibly instead of resetting data. Locking currently targets macOS and
Linux; Windows packaging is not supported in this slice.
