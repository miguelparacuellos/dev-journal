# 05: Record O2O meetings and follow-up actions

**What to build:** Record an O2O meeting with notes, addressed topics, agreements, and follow-up tasks. Close the meeting while preserving its history and leaving unaddressed topics available for a later meeting.

**Blocked by:** 02: Manage tasks and today's plan; 04: Collect and review O2O topics.

**Status:** ready-for-agent

- [x] Users can start an O2O meeting, consult open topics, and save meeting notes.
- [x] Users can mark topics addressed and distinguish them from topics that remain open.
- [x] Users can record O2O agreements and create follow-up tasks associated with the meeting.
- [x] Follow-up actions are available in Tasks and support the same planning and completion behavior as other tasks.
- [x] Closing a meeting preserves its notes, agreements, addressed topics, and associated actions.
- [x] Unaddressed topics remain open and available for the next meeting.
- [x] Users can browse previous meetings and retrieve their notes, agreements, and actions after reopening.
- [x] The meeting workflow is usable by keyboard with visible actions; behavior tests exercise the complete meeting cycle, persistence, and follow-up task relationships through the public application interface.


## Comments

Implementation completed in `66bf651`, with review fixes in `1a48be4` and `8f749f7`
(review baseline `c3613f3`). The `journal.Journal` boundary gains O2O meetings:
`StartMeeting` (one meeting in progress at a time), `CurrentMeeting`,
`SaveMeetingNotes` (blank notes clear them), `AddressTopic` (reversible while
the meeting is in progress), `RecordAgreement`, `CreateFollowUpTask`,
`CloseMeeting`, `Agenda` (open topics plus those the meeting addressed, in
capture order) and `Meetings` (newest first). Each returns a `MeetingRecord`
with the addressed topics, agreements, and follow-up tasks. The stored JSON stays
at version 1 and gains `meetings`, `agreements`, `Topic.addressed_in` and
`Task.meeting_id`, all omitted when empty. `OpenTopics` excludes addressed
topics, and a closed meeting rejects every change.

The O2O view adds the meeting workflow. **s** starts a meeting; the list then
shows `[open]` and `[addressed]` topics and **a** toggles the state. **w**, **r**
and **f** open the composer for notes, an agreement, or a follow-up task, each
with its own focus label and saved status. **v** reads the full record. **c**
closes after a second **c**; any other key only cancels, and a retained draft
blocks closing. **m** opens a meetings pane with full-record reading. At 110
columns or more, the meeting record and the selected past meeting appear beside
the lists. Follow-up tasks appear in Tasks and Today's plan marked `O2O:` and
use the normal planning and completion. The `meetings` CLI command prints every
record.

Red/green behavior tests through the public boundary with real temporary
storage cover notes and reopening, addressed versus open topics, agreements
and follow-up relationships, follow-up planning and completion shared with
Tasks, closing and immutability, unaddressed topics carried into the next
meeting, history order and isolation, the agenda order, and clearing notes. A
failed-save rollback test fails when the new slices are removed from the
`change()` snapshot. A concurrent-sessions test is included too. The full Go suite, vet, gofmt and build
passed.

`scripts/o2o_qa.py` now also runs the complete meeting cycle in a real OS PTY at
80x24 mono/ASCII. It covers start, addressing and reopening a topic,
multiline notes, agreement, follow-up, unchanged notes leaving no draft, draft
protection on close, reading the record, cancelling and then confirming the
close, and planning and completing the follow-up in Tasks. It then reopens a
month later at 120x40 light/Unicode to browse the meeting history, read the
record across a resize, and hold a second meeting. The JSON is checked for
relationships and CLI output. Two review regressions fail without their fixes.
The terminal, Tasks and Daily checks also passed. A temporary render probe
measured 23 of 24 rows and at most 80 columns at 80x24 (39 rows and 120 columns
at 120x40) for topics, meeting in progress, help, the record, the meetings pane,
and after closing. The probe was removed.

Native emulator, font and contrast review is still unverified, as for earlier
tickets. The earlier computer-use request was rejected and nothing was done to
bypass it. This ticket's acceptance items do not depend on that review.

Two-axis review:
- Standards: no hard violations. Judgement calls fixed: the capture kinds now
  live in one table (key, labels, save function), shared list windowing and
  wide-preview helpers, a `meetingState` helper, and named layout row
  constants. The `currentMeeting` middle man was removed. The close confirmation
  still spans `Update` and the O2O handler through a `closing` flag, because
  every key must be able to cancel it.
- Spec: two P2 findings were fixed. The compact meeting hints and help now show
  **v**, and reading notes with **w** then pressing Esc no longer leaves a draft
  that blocks closing. Three P3s were fixed: blank notes clear, the selection
  after saving a blocker from the meetings pane is clamped, and the key that
  cancels closing is consumed. The P3 about the `meetings` CLI command being
  scope creep was declined: it mirrors the existing `topics` listing and is
  documented.
- Spec recheck of `1a48be4`: no P1 or P2 findings. Two P3s were fixed in `8f749f7`:
  the compact meeting hint shows **? Help** again (70 columns), and the list
  footer names **Enter** for reading a topic. The third P3 is intentional: after
  a close prompt, the next key only cancels, so Ctrl+C then needs a second
  press.
