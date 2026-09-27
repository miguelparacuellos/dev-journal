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
