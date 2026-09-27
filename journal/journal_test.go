package journal_test

import (
	"devjournal/journal"
	"os"
	"testing"
)

func TestCaptureSurvivesReopening(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, err := journal.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := app.Capture("2026-09-25", "Fixed login\nAdded regression coverage")
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := journal.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	entries := reopened.Entries("2026-09-25")
	if len(entries) != 1 || entries[0].ID != entry.ID || entries[0].Text != "Fixed login\nAdded regression coverage" {
		t.Fatalf("unexpected entries: %#v", entries)
	}
}
func TestCorrectionPreservesWorkdayAndOtherEntries(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	first, _ := app.Capture("2026-09-25", "Investigating login")
	app.Capture("2026-09-27", "Another outcome")
	if err := app.Correct(first.ID, "Fixed login\nEspaña 日本語 👩‍💻"); err != nil {
		t.Fatal(err)
	}
	reopened, err := journal.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	old := reopened.Entries("2026-09-25")
	if len(old) != 1 || old[0].Text != "Fixed login\nEspaña 日本語 👩‍💻" || old[0].Workday != "2026-09-25" {
		t.Fatalf("correction lost association: %#v", old)
	}
	if len(reopened.Entries("2026-09-27")) != 1 {
		t.Fatal("other entry lost")
	}
}
func TestTwoOpenSessionsDoNotOverwriteEachOther(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	a, _ := journal.Open(path)
	b, _ := journal.Open(path)
	if _, err := a.Capture("2026-09-27", "First"); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Capture("2026-09-27", "Second"); err != nil {
		t.Fatal(err)
	}
	c, err := journal.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Entries("2026-09-27")) != 2 {
		t.Fatal("a concurrent capture was lost")
	}
}
func TestInvalidCaptureDoesNotChangeSavedLog(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	app.Capture("2026-09-27", "Keep me")
	if _, err := app.Capture("2026-09-27", " \n"); err == nil {
		t.Fatal("blank entry accepted")
	}
	if _, err := app.Capture("Monday", "Oops"); err == nil {
		t.Fatal("invalid workday accepted")
	}
	c, _ := journal.Open(path)
	if len(c.Entries("2026-09-27")) != 1 {
		t.Fatal("saved log changed")
	}
}
func TestFailedSaveCanBeRetriedWithoutLosingSavedEntries(t *testing.T) {
	directory := t.TempDir()
	path := directory + "/data/journal.json"
	app, _ := journal.Open(path)
	// A file at the storage-directory path represents an unavailable destination.
	if err := os.WriteFile(directory+"/data", []byte("unavailable"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Capture("2026-09-27", "Retry me"); err == nil {
		t.Fatal("save falsely reported success")
	}
	if err := os.Remove(directory + "/data"); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Capture("2026-09-27", "Retry me"); err != nil {
		t.Fatal(err)
	}
	reopened, err := journal.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	entries := reopened.Entries("2026-09-27")
	if len(entries) != 1 || entries[0].Text != "Retry me" {
		t.Fatal("retry failed")
	}
}

func TestTaskCaptureSurvivesReopeningWithoutWorkday(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	task, err := app.CreateTask("Review deployment")
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := journal.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	tasks := reopened.Tasks()
	if len(tasks) != 1 || tasks[0].ID != task.ID || tasks[0].Text != "Review deployment" || tasks[0].Completed {
		t.Fatalf("unexpected tasks: %#v", tasks)
	}
	if len(reopened.Workdays()) != 0 {
		t.Fatal("task created a daily log workday")
	}
}

func TestPlanIsDatedPersistentAndDoesNotCompleteOrCarryOver(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	task, _ := app.CreateTask("Review deployment")
	if err := app.PlanTask("2026-09-27", task.ID); err != nil {
		t.Fatal(err)
	}
	if err := app.PlanTask("2026-09-27", task.ID); err != nil {
		t.Fatal(err)
	}
	reopened, _ := journal.Open(path)
	plan := reopened.Plan("2026-09-27")
	if len(plan) != 1 || plan[0].ID != task.ID || plan[0].Completed {
		t.Fatalf("planning changed completion or duplicated selection: %#v", plan)
	}
	if len(reopened.Entries("2026-09-27")) != 0 {
		t.Fatal("intention recorded as completed work")
	}
	if len(reopened.Plan("2026-09-28")) != 0 || len(reopened.Tasks()) != 1 {
		t.Fatal("automatic carryover or task lost")
	}
	if err := reopened.PlanTask("2026-09-28", task.ID); err != nil {
		t.Fatal(err)
	}
	if len(reopened.Plan("2026-09-28")) != 1 {
		t.Fatal("open task cannot be selected again")
	}
}

func TestCompletionPersistsAndRejectsInvalidPlanning(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	task, _ := app.CreateTask("Review deployment")
	app.PlanTask("2026-09-27", task.ID)
	if err := app.CompleteTask(task.ID); err != nil {
		t.Fatal(err)
	}
	reopened, _ := journal.Open(path)
	if !reopened.Tasks()[0].Completed || !reopened.Plan("2026-09-27")[0].Completed {
		t.Fatal("completion lost")
	}
	for _, err := range []error{reopened.PlanTask("2026-09-28", task.ID), reopened.PlanTask("bad", task.ID), reopened.PlanTask("2026-09-28", "missing"), reopened.CompleteTask("missing")} {
		if err == nil {
			t.Fatal("invalid task action accepted")
		}
	}
	if _, err := reopened.CreateTask(" \n"); err == nil {
		t.Fatal("blank task accepted")
	}
	if len(reopened.Plan("2026-09-28")) != 0 {
		t.Fatal("invalid planning changed data")
	}
}

func TestPlanCanBeDeselectedWithoutLosingTaskOrOtherDays(t *testing.T) {
	app, _ := journal.Open(t.TempDir() + "/journal.json")
	task, _ := app.CreateTask("Review deployment")
	app.PlanTask("2026-09-27", task.ID)
	app.PlanTask("2026-09-28", task.ID)
	if err := app.UnplanTask("2026-09-27", task.ID); err != nil {
		t.Fatal(err)
	}
	if len(app.Plan("2026-09-27")) != 0 || len(app.Plan("2026-09-28")) != 1 || app.Tasks()[0].Completed {
		t.Fatal("deselect changed task or other date")
	}
}

func TestDailyUsesLatestRecordedDayAcrossWeekendsAndAbsences(t *testing.T) {
	app, _ := journal.Open(t.TempDir() + "/journal.json")
	if day, entries := app.RecentWork("2026-09-28"); day != "" || len(entries) != 0 {
		t.Fatal("missing work must be empty")
	}
	app.Capture("2026-09-25", "Shipped login")
	app.Capture("2026-09-28", "Today is not recent work")
	app.Capture("2026-10-09", "Future work")
	for _, today := range []string{"2026-09-28", "2026-10-01"} {
		day, entries := app.RecentWork(today)
		expected := "2026-09-25"
		if today == "2026-10-01" {
			expected = "2026-09-28"
		}
		if day != expected || len(entries) != 1 {
			t.Fatalf("%s: %s %#v", today, day, entries)
		}
	}
}

func TestBlockersPersistSeparatelyFromProgressAndPlan(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	if err := app.RecordBlocker("2026-09-28", "Awaiting access"); err != nil {
		t.Fatal(err)
	}
	reopened, _ := journal.Open(path)
	if got := reopened.Blockers("2026-09-28"); len(got) != 1 || got[0].Text != "Awaiting access" {
		t.Fatalf("%#v", got)
	}
	if len(reopened.Entries("2026-09-28")) != 0 || len(reopened.Workdays()) != 0 || len(reopened.Plan("2026-09-28")) != 0 || len(reopened.Blockers("2026-09-29")) != 0 {
		t.Fatal("blocker leaked into another concept or day")
	}
	if app.RecordBlocker("bad", "text") == nil || app.RecordBlocker("2026-09-28", " ") == nil {
		t.Fatal("invalid blocker accepted")
	}
}

func TestSelectedDailyPreparationCanBeEditedWithoutChangingSources(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	first, _ := app.Capture("2026-09-25", "Shipped login")
	app.Capture("2026-09-25", "Unshared investigation")
	task, _ := app.CreateTask("Review deployment")
	app.PlanTask("2026-09-28", task.ID)
	app.RecordBlocker("2026-09-28", "Awaiting access")
	text, err := app.PrepareDaily("2026-09-28", []string{first.ID})
	if err != nil {
		t.Fatal(err)
	}
	if text != "Recent work (2026-09-25)\n- Shipped login\n\nToday's plan\n- Review deployment\n\nBlockers\n- Awaiting access" {
		t.Fatalf("unexpected preparation: %q", text)
	}
	if _, err := app.PrepareDaily("2026-09-28", []string{"missing"}); err == nil {
		t.Fatal("unrelated progress accepted")
	}
	if err := app.SaveDaily("2026-09-28", "Delivered login; asking for access."); err != nil {
		t.Fatal(err)
	}
	app.SaveDaily("2026-09-29", "Next day's update")
	reopened, _ := journal.Open(path)
	text, ok := reopened.Daily("2026-09-28")
	if !ok || text != "Delivered login; asking for access." {
		t.Fatal("personal preparation lost")
	}
	if _, ok := reopened.Daily("2026-09-30"); ok {
		t.Fatal("proposal carried over")
	}
	if reopened.Entries("2026-09-25")[0].Text != "Shipped login" || len(reopened.Entries("2026-09-25")) != 2 || reopened.Tasks()[0].Completed || len(reopened.Plan("2026-09-28")) != 1 || reopened.Blockers("2026-09-28")[0].Text != "Awaiting access" {
		t.Fatal("source modified")
	}
}

func TestO2OTopicsCollectAcrossDaysAndSurviveReopeningWithoutAMeeting(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	if topics := app.OpenTopics(); len(topics) != 0 {
		t.Fatalf("new journal has topics: %#v", topics)
	}
	first, err := app.AddTopic("2026-09-03", "Feedback on the incident review")
	if err != nil {
		t.Fatal(err)
	}
	second, err := app.AddTopic("2026-09-21", "Proposal: rotate on-call\nwith a written handoff")
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := journal.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	topics := reopened.OpenTopics()
	if len(topics) != 2 || topics[0] != first || topics[1] != second {
		t.Fatalf("unexpected open topics: %#v", topics)
	}
	if topics[0].Day != "2026-09-03" || topics[1].Text != "Proposal: rotate on-call\nwith a written handoff" {
		t.Fatalf("topic lost its capture day or text: %#v", topics)
	}
	if len(reopened.Workdays()) != 0 || len(reopened.Tasks()) != 0 || len(reopened.Blockers("2026-09-03")) != 0 {
		t.Fatal("topic leaked into another concept")
	}
}

func TestInvalidTopicsAreRejectedWithoutChangingOpenTopics(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	app.AddTopic("2026-09-03", "Keep me")
	if _, err := app.AddTopic("2026-09-03", " \n"); err == nil {
		t.Fatal("blank topic accepted")
	}
	if _, err := app.AddTopic("September", "Oops"); err == nil {
		t.Fatal("invalid topic date accepted")
	}
	reopened, _ := journal.Open(path)
	if topics := reopened.OpenTopics(); len(topics) != 1 || topics[0].Text != "Keep me" {
		t.Fatalf("saved topics changed: %#v", topics)
	}
}

func TestTopicsFromTwoOpenSessionsAreBothKept(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	tui, _ := journal.Open(path)
	cli, _ := journal.Open(path)
	if _, err := tui.AddTopic("2026-09-03", "From the TUI"); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.AddTopic("2026-09-04", "From quick capture"); err != nil {
		t.Fatal(err)
	}
	if topics := cli.OpenTopics(); len(topics) != 2 {
		t.Fatalf("a concurrent topic was lost: %#v", topics)
	}
}

func TestMeetingNotesAreSavedAndSurviveReopening(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	if _, open := app.CurrentMeeting(); open {
		t.Fatal("new journal has a meeting in progress")
	}
	app.AddTopic("2026-09-03", "Feedback on the incident review")
	meeting, err := app.StartMeeting("2026-09-30")
	if err != nil {
		t.Fatal(err)
	}
	if topics := app.OpenTopics(); len(topics) != 1 {
		t.Fatalf("starting a meeting changed open topics: %#v", topics)
	}
	if err := app.SaveMeetingNotes(meeting.ID, "Talked about on-call\nand the review"); err != nil {
		t.Fatal(err)
	}
	reopened, _ := journal.Open(path)
	current, open := reopened.CurrentMeeting()
	if !open || current.Meeting.ID != meeting.ID || current.Meeting.Day != "2026-09-30" || current.Meeting.Notes != "Talked about on-call\nand the review" || current.Meeting.ClosedOn != "" {
		t.Fatalf("meeting in progress lost: %#v", current)
	}
	if _, err := reopened.StartMeeting("2026-09-30"); err == nil {
		t.Fatal("a second meeting started while one is in progress")
	}
}

func TestAddressedTopicsAreDistinguishedFromTopicsThatRemainOpen(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	first, _ := app.AddTopic("2026-09-03", "Feedback on the incident review")
	second, _ := app.AddTopic("2026-09-21", "Proposal: rotate on-call")
	third, _ := app.AddTopic("2026-09-22", "Promotion path")
	meeting, _ := app.StartMeeting("2026-09-30")
	if err := app.AddressTopic(meeting.ID, first.ID, true); err != nil {
		t.Fatal(err)
	}
	if err := app.AddressTopic(meeting.ID, third.ID, true); err != nil {
		t.Fatal(err)
	}
	// A topic marked by mistake can be reopened while the meeting is in progress.
	if err := app.AddressTopic(meeting.ID, third.ID, false); err != nil {
		t.Fatal(err)
	}
	reopened, _ := journal.Open(path)
	open := reopened.OpenTopics()
	if len(open) != 2 || open[0].ID != second.ID || open[1].ID != third.ID {
		t.Fatalf("addressed topic still open or open topic lost: %#v", open)
	}
	current, _ := reopened.CurrentMeeting()
	if len(current.Addressed) != 1 || current.Addressed[0].ID != first.ID || current.Addressed[0].AddressedIn != meeting.ID || current.Addressed[0].Text != first.Text {
		t.Fatalf("addressed topics not recorded with the meeting: %#v", current.Addressed)
	}
	for _, err := range []error{reopened.AddressTopic(meeting.ID, "missing", true), reopened.AddressTopic("missing", second.ID, true)} {
		if err == nil {
			t.Fatal("invalid addressing accepted")
		}
	}
}

func TestAgreementsAndFollowUpTasksBelongToTheMeeting(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	ordinary, _ := app.CreateTask("Review deployment")
	meeting, _ := app.StartMeeting("2026-09-30")
	agreement, err := app.RecordAgreement(meeting.ID, "Rotate on-call monthly\nstarting in October")
	if err != nil {
		t.Fatal(err)
	}
	followUp, err := app.CreateFollowUpTask(meeting.ID, "Draft the on-call handoff")
	if err != nil {
		t.Fatal(err)
	}
	if followUp.MeetingID != meeting.ID || agreement.MeetingID != meeting.ID || ordinary.MeetingID != "" {
		t.Fatalf("relationships missing: %#v %#v %#v", followUp, agreement, ordinary)
	}
	reopened, _ := journal.Open(path)
	current, _ := reopened.CurrentMeeting()
	if len(current.Agreements) != 1 || current.Agreements[0] != agreement || current.Agreements[0].Text != "Rotate on-call monthly\nstarting in October" {
		t.Fatalf("agreement lost: %#v", current.Agreements)
	}
	if len(current.FollowUps) != 1 || current.FollowUps[0] != followUp {
		t.Fatalf("follow-up task lost: %#v", current.FollowUps)
	}
	// The follow-up is an ordinary task: listed in Tasks, planned without completing, then completed.
	tasks := reopened.Tasks()
	if len(tasks) != 2 || tasks[1] != followUp {
		t.Fatalf("follow-up not available in Tasks: %#v", tasks)
	}
	if err := reopened.PlanTask("2026-10-01", followUp.ID); err != nil {
		t.Fatal(err)
	}
	if plan := reopened.Plan("2026-10-01"); len(plan) != 1 || plan[0].Completed || len(reopened.Plan("2026-10-02")) != 0 {
		t.Fatalf("follow-up planning differs from other tasks: %#v", plan)
	}
	if err := reopened.CompleteTask(followUp.ID); err != nil {
		t.Fatal(err)
	}
	again, _ := journal.Open(path)
	current, _ = again.CurrentMeeting()
	if !current.FollowUps[0].Completed || !again.Plan("2026-10-01")[0].Completed || again.Tasks()[0].Completed {
		t.Fatalf("completion not shared between Tasks and the meeting: %#v", current.FollowUps)
	}
	if _, err := again.RecordAgreement(meeting.ID, " \n"); err == nil {
		t.Fatal("blank agreement accepted")
	}
	if _, err := again.CreateFollowUpTask(meeting.ID, " "); err == nil {
		t.Fatal("blank follow-up accepted")
	}
	if _, err := again.RecordAgreement("missing", "Orphan"); err == nil {
		t.Fatal("agreement without a meeting accepted")
	}
	if _, err := again.CreateFollowUpTask("missing", "Orphan"); err == nil {
		t.Fatal("follow-up without a meeting accepted")
	}
	if len(again.Tasks()) != 2 {
		t.Fatal("rejected follow-up changed Tasks")
	}
}

func TestClosingPreservesTheMeetingAndLeavesUnaddressedTopicsForTheNextOne(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	app, _ := journal.Open(path)
	addressed, _ := app.AddTopic("2026-09-03", "Feedback on the incident review")
	remaining, _ := app.AddTopic("2026-09-21", "Promotion path")
	september, _ := app.StartMeeting("2026-09-30")
	app.SaveMeetingNotes(september.ID, "September notes")
	app.AddressTopic(september.ID, addressed.ID, true)
	agreement, _ := app.RecordAgreement(september.ID, "Rotate on-call monthly")
	followUp, _ := app.CreateFollowUpTask(september.ID, "Draft the on-call handoff")
	if err := app.CloseMeeting(september.ID, "2026-09-29"); err == nil {
		t.Fatal("meeting closed before it started")
	}
	if err := app.CloseMeeting(september.ID, "2026-09-30"); err != nil {
		t.Fatal(err)
	}
	if _, open := app.CurrentMeeting(); open {
		t.Fatal("closed meeting still in progress")
	}
	// A closed meeting is history: it no longer accepts changes.
	for _, err := range []error{
		app.SaveMeetingNotes(september.ID, "Rewritten"),
		app.AddressTopic(september.ID, remaining.ID, true),
		app.AddressTopic(september.ID, addressed.ID, false),
		app.CloseMeeting(september.ID, "2026-10-01"),
	} {
		if err == nil {
			t.Fatal("closed meeting changed")
		}
	}
	if _, err := app.RecordAgreement(september.ID, "Late"); err == nil {
		t.Fatal("agreement added to a closed meeting")
	}
	if _, err := app.CreateFollowUpTask(september.ID, "Late"); err == nil {
		t.Fatal("follow-up added to a closed meeting")
	}

	reopened, _ := journal.Open(path)
	if open := reopened.OpenTopics(); len(open) != 1 || open[0].ID != remaining.ID {
		t.Fatalf("unaddressed topic not available for the next meeting: %#v", open)
	}
	october, err := reopened.StartMeeting("2026-10-28")
	if err != nil {
		t.Fatal(err)
	}
	if err := reopened.AddressTopic(october.ID, addressed.ID, false); err == nil {
		t.Fatal("a later meeting reopened a topic addressed earlier")
	}
	reopened.AddressTopic(october.ID, remaining.ID, true)
	reopened.CloseMeeting(october.ID, "2026-10-28")

	final, _ := journal.Open(path)
	meetings := final.Meetings()
	if len(meetings) != 2 || meetings[0].Meeting.ID != october.ID || meetings[1].Meeting.ID != september.ID {
		t.Fatalf("meeting history not newest first: %#v", meetings)
	}
	past := meetings[1]
	if past.Meeting.Notes != "September notes" || past.Meeting.ClosedOn != "2026-09-30" || past.Meeting.Day != "2026-09-30" {
		t.Fatalf("closed meeting lost its notes or dates: %#v", past.Meeting)
	}
	if len(past.Addressed) != 1 || past.Addressed[0].ID != addressed.ID || len(past.Agreements) != 1 || past.Agreements[0] != agreement || len(past.FollowUps) != 1 || past.FollowUps[0].ID != followUp.ID {
		t.Fatalf("closed meeting lost its records: %#v", past)
	}
	if len(meetings[0].Addressed) != 1 || meetings[0].Addressed[0].ID != remaining.ID || len(meetings[0].Agreements) != 0 || len(meetings[0].FollowUps) != 0 {
		t.Fatalf("records leaked between meetings: %#v", meetings[0])
	}
	if len(final.OpenTopics()) != 0 || len(final.Tasks()) != 1 || len(final.Workdays()) != 0 {
		t.Fatal("meeting history changed unrelated concepts")
	}
	if _, err := final.StartMeeting("October"); err == nil {
		t.Fatal("invalid meeting date accepted")
	}
	if err := final.CloseMeeting("missing", "2026-10-28"); err == nil {
		t.Fatal("unknown meeting closed")
	}
}

func TestFailedMeetingSavesLeaveTheSessionUnchangedAndCanBeRetried(t *testing.T) {
	directory := t.TempDir()
	path := directory + "/data/journal.json"
	app, _ := journal.Open(path)
	topic, _ := app.AddTopic("2026-09-03", "Feedback")
	meeting, _ := app.StartMeeting("2026-09-30")
	// A read-only storage directory makes each write fail after the change is applied.
	if err := os.Chmod(directory+"/data", 0500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(directory+"/data", 0700)
	unchanged := func(action string, err error) {
		t.Helper()
		if err == nil {
			t.Fatalf("%s: save falsely reported success", action)
		}
		current, open := app.CurrentMeeting()
		if !open || current.Meeting.ClosedOn != "" || current.Meeting.Notes != "" || len(current.Addressed)+len(current.Agreements)+len(current.FollowUps) != 0 || len(app.OpenTopics()) != 1 || len(app.Meetings()) != 1 || len(app.Tasks()) != 0 {
			t.Fatalf("%s: failed save changed the session: %#v", action, current)
		}
	}
	unchanged("address", app.AddressTopic(meeting.ID, topic.ID, true))
	unchanged("notes", app.SaveMeetingNotes(meeting.ID, "Notes"))
	_, err := app.RecordAgreement(meeting.ID, "Agreed")
	unchanged("agreement", err)
	_, err = app.CreateFollowUpTask(meeting.ID, "Follow up")
	unchanged("follow-up", err)
	unchanged("close", app.CloseMeeting(meeting.ID, "2026-09-30"))
	os.Chmod(directory+"/data", 0700)
	if _, err := app.RecordAgreement(meeting.ID, "Agreed"); err != nil {
		t.Fatal(err)
	}
	reopened, _ := journal.Open(path)
	if current, _ := reopened.CurrentMeeting(); len(current.Agreements) != 1 {
		t.Fatal("retry was not saved")
	}
}

func TestMeetingChangesFromTwoOpenSessionsAreBothKept(t *testing.T) {
	path := t.TempDir() + "/journal.json"
	tui, _ := journal.Open(path)
	cli, _ := journal.Open(path)
	meeting, _ := tui.StartMeeting("2026-09-30")
	if _, err := cli.CreateFollowUpTask(meeting.ID, "From quick capture"); err != nil {
		t.Fatal(err)
	}
	if _, err := tui.RecordAgreement(meeting.ID, "From the TUI"); err != nil {
		t.Fatal(err)
	}
	current, _ := tui.CurrentMeeting()
	if len(current.Agreements) != 1 || len(current.FollowUps) != 1 {
		t.Fatalf("a concurrent meeting change was lost: %#v", current)
	}
}

func TestMeetingAgendaListsOpenAndAddressedTopicsInCaptureOrder(t *testing.T) {
	app, _ := journal.Open(t.TempDir() + "/journal.json")
	earlier, _ := app.AddTopic("2026-08-01", "Addressed last month")
	first, _ := app.AddTopic("2026-09-03", "Feedback")
	second, _ := app.AddTopic("2026-09-21", "Rotation")
	august, _ := app.StartMeeting("2026-08-28")
	app.AddressTopic(august.ID, earlier.ID, true)
	app.CloseMeeting(august.ID, "2026-08-28")
	meeting, _ := app.StartMeeting("2026-09-30")
	app.AddressTopic(meeting.ID, first.ID, true)
	third, _ := app.AddTopic("2026-09-30", "Raised during the meeting")
	agenda := app.Agenda(meeting.ID)
	if len(agenda) != 3 || agenda[0].ID != first.ID || agenda[0].AddressedIn != meeting.ID || agenda[1].ID != second.ID || agenda[2].ID != third.ID {
		t.Fatalf("unexpected agenda: %#v", agenda)
	}
	if agenda := app.Agenda(""); len(agenda) != 2 || agenda[0].ID != second.ID {
		t.Fatalf("agenda without a meeting must list open topics: %#v", agenda)
	}
}
