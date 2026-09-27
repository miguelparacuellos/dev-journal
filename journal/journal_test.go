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
