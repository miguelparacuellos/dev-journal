package journal_test

import (
	"devjournal/journal"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

// backupFile saves a backup of app into directory and returns its path.
func backupFile(t *testing.T, app *journal.Journal, directory string) string {
	t.Helper()
	path := directory + "/backup.json"
	if err := app.ExportBackup(path, "2026-09-27"); err != nil {
		t.Fatal(err)
	}
	return path
}

// restoreInto restores the backup at path into a fresh journal at target.
func restoreInto(t *testing.T, backupPath, target string) *journal.Journal {
	t.Helper()
	b, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	app, err := journal.Open(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Restore(b); err != nil {
		t.Fatal(err)
	}
	return app
}

func TestBackupRestoresTheCompleteJournalIntoAnEmptyInstallation(t *testing.T) {
	directory := t.TempDir()
	original := representativeJournal(t, directory+"/original.json")
	// A second meeting stays in progress, with its own addressed topic and follow-up.
	rotation := original.OpenTopics()[0]
	meeting, _ := original.StartMeeting("2026-09-26")
	if err := original.AddressTopic(meeting.ID, rotation.ID, true); err != nil {
		t.Fatal(err)
	}
	followUp, _ := original.CreateFollowUpTask(meeting.ID, "Ask about rotation dates")
	if err := original.PlanTask("2026-09-26", followUp.ID); err != nil {
		t.Fatal(err)
	}
	if err := original.SaveDaily("2026-09-24", "Nothing earlier to share"); err != nil {
		t.Fatal(err)
	}

	restored := restoreInto(t, backupFile(t, original, directory), directory+"/fresh/journal.json")
	// Reopening proves the restore was saved, not only applied to the session.
	reopened, err := journal.Open(directory + "/fresh/journal.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, app := range []*journal.Journal{restored, reopened} {
		if got, want := app.Markdown("2026-09-27"), original.Markdown("2026-09-27"); got != want {
			t.Fatalf("restored journal reads differently:\n%s\nwant:\n%s", got, want)
		}
		if !reflect.DeepEqual(app.Tasks(), original.Tasks()) || !reflect.DeepEqual(app.Meetings(), original.Meetings()) ||
			!reflect.DeepEqual(app.OpenTopics(), original.OpenTopics()) || !reflect.DeepEqual(app.Workdays(), original.Workdays()) {
			t.Fatal("restored tasks, meetings, topics, or workdays differ")
		}
		for _, day := range []string{"2026-09-24", "2026-09-25", "2026-09-26"} {
			if !reflect.DeepEqual(app.Entries(day), original.Entries(day)) || !reflect.DeepEqual(app.Plan(day), original.Plan(day)) ||
				!reflect.DeepEqual(app.Blockers(day), original.Blockers(day)) {
				t.Fatalf("restored %s differs", day)
			}
			gotText, gotSaved := app.Daily(day)
			wantText, wantSaved := original.Daily(day)
			if gotText != wantText || gotSaved != wantSaved {
				t.Fatalf("restored daily proposal for %s differs", day)
			}
			gotDraft, _ := app.PrepareDaily(day, nil)
			wantDraft, _ := original.PrepareDaily(day, nil)
			if gotDraft != wantDraft {
				t.Fatalf("historical daily preparation for %s differs:\n%s\nwant:\n%s", day, gotDraft, wantDraft)
			}
		}
		again, _ := app.Backup("2026-09-27")
		first, _ := original.Backup("2026-09-27")
		if string(again) != string(first) {
			t.Fatal("backing up the restored journal gives a different backup")
		}
	}

	// The restored meeting in progress keeps working, and its follow-up is an ordinary task.
	current, open := reopened.CurrentMeeting()
	if !open || current.Meeting.ID != meeting.ID || len(current.Addressed) != 1 || len(current.FollowUps) != 1 {
		t.Fatalf("restored meeting in progress lost its records: %#v", current)
	}
	if _, err := reopened.RecordAgreement(meeting.ID, "Rotation starts in October"); err != nil {
		t.Fatal(err)
	}
	if err := reopened.CompleteTask(followUp.ID); err != nil {
		t.Fatal(err)
	}
	if err := reopened.CloseMeeting(meeting.ID, "2026-09-27"); err != nil {
		t.Fatal(err)
	}
	if _, open := reopened.CurrentMeeting(); open {
		t.Fatal("restored meeting did not close")
	}
	if _, err := reopened.StartMeeting("2026-10-20"); err != nil {
		t.Fatal(err)
	}
}

func TestRestoreKeepsWhetherAProposalSourceWasRecorded(t *testing.T) {
	directory := t.TempDir()
	path := directory + "/journal.json"
	// A proposal saved before sources were recorded has no source field.
	legacy := `{"version": 1, "entries": [{"id": "e1", "workday": "2026-09-24", "text": "Work"}],
		"prepared": [{"day": "2026-09-25", "text": "Recent work (2026-09-24)"}]}`
	if err := os.WriteFile(path, []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	original, _ := journal.Open(path)
	if err := original.SaveDaily("2026-09-21", "Nothing earlier"); err != nil {
		t.Fatal(err)
	}
	restored := restoreInto(t, backupFile(t, original, directory), directory+"/fresh.json")
	if got, want := restored.Markdown("2026-09-27"), original.Markdown("2026-09-27"); got != want {
		t.Fatalf("restore changed which proposal sources were recorded:\n%s\nwant:\n%s", got, want)
	}
}

func TestRestoreRefusesAJournalThatAlreadyHasRecords(t *testing.T) {
	directory := t.TempDir()
	backup := backupFile(t, representativeJournal(t, directory+"/original.json"), directory)
	b, _ := os.ReadFile(backup)
	path := directory + "/current.json"
	current, _ := journal.Open(path)
	if _, err := current.CreateTask("Unrelated task"); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	if err := current.Restore(b); err == nil {
		t.Fatal("restore overwrote a journal that already had records")
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatal("a refused restore changed the stored journal")
	}
	if tasks := current.Tasks(); len(tasks) != 1 || tasks[0].Text != "Unrelated task" || len(current.Workdays()) != 0 {
		t.Fatalf("a refused restore changed the session: %#v", tasks)
	}
}

func TestAnEmptyJournalBacksUpAndRestores(t *testing.T) {
	directory := t.TempDir()
	empty, _ := journal.Open(directory + "/empty.json")
	restored := restoreInto(t, backupFile(t, empty, directory), directory+"/fresh.json")
	reopened, err := journal.Open(directory + "/fresh.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, app := range []*journal.Journal{restored, reopened} {
		if got, want := app.Markdown("2026-09-27"), empty.Markdown("2026-09-27"); got != want {
			t.Fatalf("restored empty journal is not empty:\n%s", got)
		}
		if app.Entries("2026-09-27") == nil || len(app.Tasks()) != 0 {
			t.Fatal("restored empty journal is not usable")
		}
	}
	if _, err := reopened.Capture("2026-09-27", "First entry after restoring"); err != nil {
		t.Fatal(err)
	}
}

func TestMalformedOrUnsupportedBackupsAreRejectedWithoutRestoringAnything(t *testing.T) {
	valid := `"format": "devjournal-backup", "backed_up_on": "2026-09-27", "version": 1, "entries": []`
	for _, test := range []struct{ name, backup, message string }{
		{"not JSON", `# Dev Journal`, "backup cannot be read"},
		{"truncated", `{` + valid, "backup cannot be read"},
		{"trailing content", `{` + valid + `} {}`, "backup cannot be read"},
		{"trailing brace", `{` + valid + `}}`, "backup cannot be read"},
		{"journal file instead of backup", `{"version": 1, "entries": []}`, "not a Dev Journal backup"},
		{"other format", `{"format": "other", "version": 1, "entries": []}`, "not a Dev Journal backup"},
		{"newer version", `{"format": "devjournal-backup", "backed_up_on": "2026-09-27", "version": 2, "entries": []}`, "unsupported backup version: 2"},
		{"unknown field", `{` + valid + `, "resources": []}`, "backup cannot be read"},
		{"wrong type", `{"format": "devjournal-backup", "version": 1, "entries": {}}`, "backup cannot be read"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir() + "/journal.json"
			app, _ := journal.Open(path)
			err := app.Restore([]byte(test.backup))
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("got error %v, want one mentioning %q", err, test.message)
			}
			if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("a rejected backup created the journal: %v", statErr)
			}
			if !strings.Contains(app.Markdown("2026-09-27"), "The journal is empty") {
				t.Fatal("a rejected backup changed the session")
			}
		})
	}
}

func TestInconsistentBackupsAreRejectedWithoutRestoringAnything(t *testing.T) {
	backup := func(records string) string {
		return `{"format": "devjournal-backup", "backed_up_on": "2026-09-27", "version": 1, ` + records + `}`
	}
	meeting := `"meetings": [{"id": "m1", "day": "2026-09-10", "closed_on": "2026-09-11"}]`
	task := `"tasks": [{"id": "t1", "text": "Ship"}]`
	for _, test := range []struct{ name, backup, message string }{
		{"follow-up names a missing meeting", backup(`"tasks": [{"id": "t1", "text": "Ship", "meeting_id": "gone"}]`), `task t1 names meeting "gone", which is not in the backup`},
		{"agreement names a missing meeting", backup(meeting + `, "agreements": [{"id": "a1", "meeting_id": "gone", "text": "Agreed"}]`), `agreement a1 names meeting "gone", which is not in the backup`},
		{"agreement without a meeting", backup(meeting + `, "agreements": [{"id": "a1", "text": "Agreed"}]`), `agreement a1 names meeting "", which is not in the backup`},
		{"topic addressed in a missing meeting", backup(`"topics": [{"id": "o1", "day": "2026-09-03", "text": "Feedback", "addressed_in": "gone"}]`), `topic o1 names meeting "gone", which is not in the backup`},
		{"plan selects a missing task", backup(task + `, "plan": [{"day": "2026-09-25", "task_id": "gone"}]`), `plan for 2026-09-25 names task "gone", which is not in the backup`},
		{"plan selects a task twice", backup(task + `, "plan": [{"day": "2026-09-25", "task_id": "t1"}, {"day": "2026-09-25", "task_id": "t1"}]`), "plan for 2026-09-25 selects task t1 twice"},
		{"two meetings in progress", backup(`"meetings": [{"id": "m1", "day": "2026-09-10"}, {"id": "m2", "day": "2026-09-20"}]`), "meetings m1 and m2 are both in progress"},
		{"meeting closes before it starts", backup(`"meetings": [{"id": "m1", "day": "2026-09-10", "closed_on": "2026-09-09"}]`), "meeting m1 closes before it starts"},
		{"duplicate task ID", backup(`"tasks": [{"id": "t1", "text": "Ship"}, {"id": "t1", "text": "Review"}]`), "task ID t1 is used twice"},
		{"duplicate entry ID", backup(`"entries": [{"id": "e1", "workday": "2026-09-25", "text": "A"}, {"id": "e1", "workday": "2026-09-26", "text": "B"}]`), "entry ID e1 is used twice"},
		{"missing ID", backup(`"topics": [{"day": "2026-09-03", "text": "Feedback"}]`), "a topic has no ID"},
		{"two proposals for one day", backup(`"prepared": [{"day": "2026-09-25", "text": "A"}, {"day": "2026-09-25", "text": "B"}]`), "two daily proposals for 2026-09-25"},
		{"invalid workday", backup(`"entries": [{"id": "e1", "workday": "Monday", "text": "A"}]`), `entry e1 has date "Monday"; dates must be YYYY-MM-DD`},
		{"invalid plan date", backup(task + `, "plan": [{"day": "25/09", "task_id": "t1"}]`), `plan selection has date "25/09"`},
		{"invalid blocker date", backup(`"blockers": [{"day": "", "text": "Access"}]`), `blocker has date ""`},
		{"invalid proposal source", backup(`"prepared": [{"day": "2026-09-25", "text": "A", "source": "yesterday"}]`), `daily proposal for 2026-09-25 has source "yesterday"`},
		{"invalid topic date", backup(`"topics": [{"id": "o1", "day": "soon", "text": "Feedback"}]`), `topic o1 has date "soon"`},
		{"invalid closing date", backup(`"meetings": [{"id": "m1", "day": "2026-09-10", "closed_on": "later"}]`), `meeting m1 has closing date "later"`},
		{"invalid backup date", `{"format": "devjournal-backup", "backed_up_on": "today", "version": 1}`, `backup date "today"`},
		{"blank entry", backup(`"entries": [{"id": "e1", "workday": "2026-09-25", "text": " \n"}]`), "entry e1 is empty"},
		{"blank task", backup(`"tasks": [{"id": "t1", "text": ""}]`), "task t1 is empty"},
		{"blank agreement", backup(meeting + `, "agreements": [{"id": "a1", "meeting_id": "m1", "text": ""}]`), "agreement a1 is empty"},
		{"blank blocker", backup(`"blockers": [{"day": "2026-09-25", "text": ""}]`), "blocker for 2026-09-25 is empty"},
		{"blank proposal", backup(`"prepared": [{"day": "2026-09-25", "text": ""}]`), "daily proposal for 2026-09-25 is empty"},
		{"blank topic", backup(`"topics": [{"id": "o1", "day": "2026-09-03", "text": ""}]`), "topic o1 is empty"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir() + "/journal.json"
			app, _ := journal.Open(path)
			err := app.Restore([]byte(test.backup))
			if err == nil || !strings.Contains(err.Error(), "backup is inconsistent: ") || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("got error %v, want an inconsistency mentioning %q", err, test.message)
			}
			if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("a rejected backup created the journal: %v", statErr)
			}
			if !strings.Contains(app.Markdown("2026-09-27"), "The journal is empty") {
				t.Fatal("a rejected backup changed the session")
			}
		})
	}
}

func TestBackingUpLeavesTheJournalUnchangedAndNeverReplacesIt(t *testing.T) {
	directory := t.TempDir()
	path := directory + "/journal.json"
	app := representativeJournal(t, path)
	before, _ := os.ReadFile(path)
	backup := backupFile(t, app, directory)
	// Backing up again replaces the earlier backup.
	if err := app.ExportBackup(backup, "2026-09-28"); err != nil {
		t.Fatal(err)
	}
	saved, _ := os.ReadFile(backup)
	if want, _ := app.Backup("2026-09-28"); string(saved) != string(want) {
		t.Fatal("the saved backup differs from the rendered backup")
	}
	for _, own := range []string{path, path + ".lock", directory + "/./journal.json"} {
		if err := app.ExportBackup(own, "2026-09-28"); err == nil || !strings.Contains(err.Error(), "cannot replace the journal") {
			t.Fatalf("backing up over %s: got %v", own, err)
		}
	}
	if err := app.ExportBackup(directory+"/missing/backup.json", "2026-09-28"); err == nil {
		t.Fatal("backing up into a missing directory falsely succeeded")
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatal("backing up changed the stored journal")
	}
	empty, _ := journal.Open(directory + "/empty/journal.json")
	if err := empty.ExportBackup(directory+"/empty.json", "2026-09-27"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(directory + "/empty/journal.json"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("backing up an empty journal created it: %v", err)
	}
}

func TestAFailedRestoreSaveLeavesTheSessionEmptyAndCanBeRetried(t *testing.T) {
	directory := t.TempDir()
	backup, _ := os.ReadFile(backupFile(t, representativeJournal(t, directory+"/original.json"), directory))
	if err := os.Mkdir(directory+"/data", 0700); err != nil {
		t.Fatal(err)
	}
	path := directory + "/data/journal.json"
	app, _ := journal.Open(path)
	// Restoring an empty backup stores an empty journal, so the next restore gets
	// as far as replacing it.
	empty, _ := journal.Open(directory + "/empty.json")
	emptyBackup, _ := empty.Backup("2026-09-27")
	if err := app.Restore(emptyBackup); err != nil {
		t.Fatal(err)
	}
	// A read-only storage directory makes the write fail after the restore is applied.
	if err := os.Chmod(directory+"/data", 0500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(directory+"/data", 0700)
	if err := app.Restore(backup); err == nil {
		t.Fatal("restore falsely reported success")
	}
	if !strings.Contains(app.Markdown("2026-09-27"), "The journal is empty") || len(app.Tasks()) != 0 {
		t.Fatal("a failed restore left partially restored data in the session")
	}
	os.Chmod(directory+"/data", 0700)
	if err := app.Restore(backup); err != nil {
		t.Fatal(err)
	}
	reopened, _ := journal.Open(path)
	if len(reopened.Tasks()) != 3 || len(reopened.Meetings()) != 1 {
		t.Fatal("the retried restore was not saved")
	}
}
