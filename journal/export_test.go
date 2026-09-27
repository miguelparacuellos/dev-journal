package journal_test

import (
	"devjournal/journal"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestExportingAnEmptyJournalSaysSoAndCreatesNoJournal(t *testing.T) {
	directory := t.TempDir()
	app, _ := journal.Open(directory + "/journal.json")
	if err := app.ExportMarkdown(directory+"/export.md", "2026-09-27"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(directory + "/export.md")
	if err != nil {
		t.Fatal(err)
	}
	markdown := string(b)
	for _, want := range []string{"# Dev Journal", "Exported on 2026-09-27", "The journal is empty"} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("export lacks %q:\n%s", want, markdown)
		}
	}
	if _, err := os.Stat(directory + "/journal.json"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("exporting created or touched the journal: %v", err)
	}
}

// representativeJournal records one of every kind of journal content through
// the public interface.
func representativeJournal(t *testing.T, path string) *journal.Journal {
	t.Helper()
	app, err := journal.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err = app.Capture("2026-09-24", "Reviewed the payments PR")
	must(err)
	_, err = app.Capture("2026-09-25", "Fixed login\nAdded regression coverage")
	must(err)
	ship, err := app.CreateTask("Ship the login fix")
	must(err)
	later, err := app.CreateTask("Write the design note")
	must(err)
	must(app.PlanTask("2026-09-25", ship.ID))
	must(app.PlanTask("2026-09-25", later.ID))
	must(app.PlanTask("2026-09-26", later.ID))
	must(app.CompleteTask(ship.ID))
	must(app.RecordBlocker("2026-09-25", "Waiting for staging access"))
	must(app.SaveDaily("2026-09-25", "Recent work (2026-09-24)\n- Reviewed the payments PR\n\nBlockers\n- Staging access"))
	feedback, err := app.AddTopic("2026-09-03", "Feedback on the migration")
	must(err)
	_, err = app.AddTopic("2026-09-21", "Team rotation")
	must(err)
	meeting, err := app.StartMeeting("2026-09-10")
	must(err)
	must(app.SaveMeetingNotes(meeting.ID, "Went well\nKeep the pace"))
	must(app.AddressTopic(meeting.ID, feedback.ID, true))
	_, err = app.RecordAgreement(meeting.ID, "Pair on reviews weekly")
	must(err)
	_, err = app.CreateFollowUpTask(meeting.ID, "Book the pairing slot")
	must(err)
	must(app.CloseMeeting(meeting.ID, "2026-09-11"))
	return app
}

// assertInOrder fails unless every part appears in markdown in the given order.
func assertInOrder(t *testing.T, markdown string, parts ...string) {
	t.Helper()
	rest := markdown
	for _, part := range parts {
		i := strings.Index(rest, part)
		if i < 0 {
			t.Fatalf("export lacks %q after the previous parts:\n%s", part, markdown)
		}
		rest = rest[i+len(part):]
	}
}

func TestExportRepresentsTheJournalAsReadableMarkdown(t *testing.T) {
	directory := t.TempDir()
	app := representativeJournal(t, directory+"/journal.json")
	if err := app.ExportMarkdown(directory+"/export.md", "2026-09-27"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(directory + "/export.md")
	if err != nil {
		t.Fatal(err)
	}
	markdown := string(b)
	if strings.Contains(markdown, "The journal is empty") {
		t.Fatalf("a journal with content is reported empty:\n%s", markdown)
	}
	// Workdays are newest first, each with its entries, plan, blockers, and proposal.
	assertInOrder(t, markdown,
		"# Dev Journal", "Exported on 2026-09-27",
		"## Daily log",
		"### 2026-09-26", "#### Today's plan", "- [ ] Open: Write the design note",
		"### 2026-09-25",
		"#### Entries", "- Fixed login  \n  Added regression coverage",
		"#### Today's plan", "- [x] Done: Ship the login fix", "- [ ] Open: Write the design note",
		"#### Blockers", "- Waiting for staging access",
		"#### Daily proposal", "Recent work source: 2026-09-24",
		"> Recent work (2026-09-24)  \n> - Reviewed the payments PR", ">\n> Blockers",
		"### 2026-09-24", "- Reviewed the payments PR",
	)
	// Tasks show their state, planned dates, and O2O origin.
	assertInOrder(t, markdown,
		"## Tasks",
		"### Open tasks (2)",
		"- [ ] Open: Write the design note · planned 2026-09-25, 2026-09-26",
		"- [ ] Open: Book the pairing slot · O2O follow-up from the 2026-09-10 meeting",
		"### Completed tasks (1)",
		"- [x] Done: Ship the login fix · planned 2026-09-25",
	)
	// Open topics stay separate from topics addressed in a meeting.
	assertInOrder(t, markdown,
		"## Open O2O topics (1)", "- Team rotation (collected 2026-09-21)",
		"## O2O meetings (1)",
		"### O2O meeting · 2026-09-10 · closed 2026-09-11",
		"#### Notes", "Went well  \nKeep the pace",
		"#### Addressed topics (1)", "- Feedback on the migration (collected 2026-09-03)",
		"#### Agreements (1)", "- Pair on reviews weekly",
		"#### Follow-up tasks (1)", "- [ ] Open: Book the pairing slot",
	)
	if strings.Count(markdown, "Feedback on the migration") != 1 {
		t.Fatalf("an addressed topic must only appear with its meeting:\n%s", markdown)
	}
}

func TestExportLeavesTheStoredJournalUnchanged(t *testing.T) {
	directory := t.TempDir()
	path := directory + "/journal.json"
	app := representativeJournal(t, path)
	before, _ := os.ReadFile(path)
	if err := app.ExportMarkdown(directory+"/export.md", "2026-09-27"); err != nil {
		t.Fatal(err)
	}
	// Exporting again replaces the earlier export.
	if err := app.ExportMarkdown(directory+"/export.md", "2026-09-28"); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("exporting changed the stored journal")
	}
	exported, _ := os.ReadFile(directory + "/export.md")
	if !strings.Contains(string(exported), "Exported on 2026-09-28") || string(exported) != app.Markdown("2026-09-28") {
		t.Fatal("the saved export differs from the rendered journal")
	}
	if err := app.ExportMarkdown(path, "2026-09-28"); err == nil {
		t.Fatal("exporting over the journal itself must be refused")
	}
	if err := app.ExportMarkdown(directory+"/missing/export.md", "2026-09-28"); err == nil {
		t.Fatal("exporting into a missing directory falsely succeeded")
	}
	if after, _ := os.ReadFile(path); string(before) != string(after) {
		t.Fatal("a refused export changed the stored journal")
	}
}

func TestExportExplainsMissingSourcesAndMeetingsInProgress(t *testing.T) {
	app, _ := journal.Open(t.TempDir() + "/journal.json")
	app.SaveDaily("2026-09-21", "Recent work\n\nToday's plan\n- Start the week")
	meeting, _ := app.StartMeeting("2026-09-30")
	app.CreateFollowUpTask(meeting.ID, "Share the \x1b[31mroadmap")
	markdown := app.Markdown("2026-09-30")
	assertInOrder(t, markdown,
		"### 2026-09-21", "#### Daily proposal", "Recent work source: none; no earlier workday had entries",
		"## Tasks", "- [ ] Open: Share the [31mroadmap · O2O follow-up from the 2026-09-30 meeting",
		"## Open O2O topics (0)", "None.",
		"### O2O meeting · 2026-09-30 · in progress", "#### Notes", "None recorded.",
		"#### Agreements (0)", "None recorded.",
	)
	if strings.Contains(markdown, "\x1b") {
		t.Fatal("terminal control characters must not reach the export")
	}
}
