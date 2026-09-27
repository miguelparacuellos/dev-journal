package journal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// backupFormat marks a file as a Dev Journal backup.
const backupFormat = "devjournal-backup"

// backup is the JSON document written by Backup: a format marker and the backup
// date followed by the complete stored journal.
type backup struct {
	Format     string `json:"format"`
	BackedUpOn string `json:"backed_up_on"`
	Data
}

// Backup renders the complete journal as a JSON backup made on the given day. It
// keeps every record, state, and relationship so Restore can recover the
// journal. It only reads the journal.
func (j *Journal) Backup(backedUpOn string) ([]byte, error) {
	if !isDate(backedUpOn) {
		return nil, errors.New("backup date must be YYYY-MM-DD")
	}
	b, err := json.MarshalIndent(backup{Format: backupFormat, BackedUpOn: backedUpOn, Data: j.data}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// ExportBackup saves Backup(backedUpOn) to path, replacing any earlier file
// there. The stored journal is left unchanged, and it is never used as the
// destination.
func (j *Journal) ExportBackup(path, backedUpOn string) error {
	target, err := j.outsideJournal(path, "the backup cannot replace the journal")
	if err != nil {
		return err
	}
	b, err := j.Backup(backedUpOn)
	if err != nil {
		return err
	}
	return writeFile(target, ".backup-*", b)
}

// Restore replaces an empty journal with the contents of a backup. A journal
// that already has records is left untouched, as is the journal when the backup
// is malformed, unsupported, or inconsistent.
func (j *Journal) Restore(b []byte) error {
	restored, err := readBackup(b)
	if err != nil {
		return err
	}
	return j.change(func() error {
		if !j.data.empty() {
			return errors.New("restore needs an empty journal; this one already has records, so restore into a new --data file instead")
		}
		j.data = restored
		return nil
	})
}

// readBackup decodes a backup strictly: fields it does not know would be lost,
// so they make the backup unsupported rather than being ignored.
func readBackup(b []byte) (Data, error) {
	var restored backup
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&restored)
	if err == nil {
		if _, next := decoder.Token(); next != io.EOF {
			err = errors.New("unexpected content after the backup")
		}
	}
	if err != nil {
		return Data{}, fmt.Errorf("backup cannot be read: %w", err)
	}
	if restored.Format != backupFormat {
		return Data{}, errors.New("not a Dev Journal backup; use a file saved by devjournal backup")
	}
	if restored.Version != 1 {
		return Data{}, fmt.Errorf("unsupported backup version: %d", restored.Version)
	}
	if err := restored.consistent(); err != nil {
		return Data{}, fmt.Errorf("backup is inconsistent: %w", err)
	}
	if restored.Entries == nil {
		restored.Entries = []Entry{}
	}
	return restored.Data, nil
}

// consistent checks what the journal itself guarantees for every record it
// saves: valid dates, non-empty text, unique IDs, references to records that
// exist, and at most one meeting in progress.
func (b backup) consistent() error {
	if !isDate(b.BackedUpOn) {
		return fmt.Errorf("backup date %q must be YYYY-MM-DD", b.BackedUpOn)
	}
	ids := map[string]map[string]bool{}
	unique := func(kind, id string) error {
		if id == "" {
			return fmt.Errorf("a %s has no ID", kind)
		}
		if ids[kind] == nil {
			ids[kind] = map[string]bool{}
		}
		if ids[kind][id] {
			return fmt.Errorf("%s ID %s is used twice", kind, id)
		}
		ids[kind][id] = true
		return nil
	}
	dated := func(record, day string) error {
		if !isDate(day) {
			return fmt.Errorf("%s has date %q; dates must be YYYY-MM-DD", record, day)
		}
		return nil
	}
	written := func(record, text string) error {
		if strings.TrimSpace(text) == "" {
			return fmt.Errorf("%s is empty", record)
		}
		return nil
	}
	// Meetings come first so that other records can refer to them.
	inProgress := ""
	for _, m := range b.Meetings {
		record := "meeting " + m.ID
		if err := firstError(unique("meeting", m.ID), dated(record, m.Day)); err != nil {
			return err
		}
		switch {
		case m.ClosedOn == "" && inProgress != "":
			return fmt.Errorf("meetings %s and %s are both in progress; only one can be", inProgress, m.ID)
		case m.ClosedOn == "":
			inProgress = m.ID
		case !isDate(m.ClosedOn):
			return fmt.Errorf("%s has closing date %q; dates must be YYYY-MM-DD", record, m.ClosedOn)
		case m.ClosedOn < m.Day:
			return fmt.Errorf("%s closes before it starts", record)
		}
	}
	names := func(record, kind, id string) error {
		if !ids[kind][id] {
			return fmt.Errorf("%s names %s %q, which is not in the backup", record, kind, id)
		}
		return nil
	}
	// mayName checks an optional meeting reference: empty means none.
	mayName := func(record, meetingID string) error {
		if meetingID == "" {
			return nil
		}
		return names(record, "meeting", meetingID)
	}
	for _, e := range b.Entries {
		record := "entry " + e.ID
		if err := firstError(unique("entry", e.ID), dated(record, e.Workday), written(record, e.Text)); err != nil {
			return err
		}
	}
	for _, t := range b.Tasks {
		record := "task " + t.ID
		if err := firstError(unique("task", t.ID), written(record, t.Text), mayName(record, t.MeetingID)); err != nil {
			return err
		}
	}
	for _, a := range b.Agreements {
		record := "agreement " + a.ID
		if err := firstError(unique("agreement", a.ID), names(record, "meeting", a.MeetingID), written(record, a.Text)); err != nil {
			return err
		}
	}
	for _, t := range b.Topics {
		record := "topic " + t.ID
		if err := firstError(unique("topic", t.ID), dated(record, t.Day), written(record, t.Text), mayName(record, t.AddressedIn)); err != nil {
			return err
		}
	}
	selected := map[PlanSelection]bool{}
	for _, s := range b.Plan {
		if err := dated("a plan selection", s.Day); err != nil {
			return err
		}
		if err := names("plan for "+s.Day, "task", s.TaskID); err != nil {
			return err
		}
		if selected[s] {
			return fmt.Errorf("plan for %s selects task %s twice", s.Day, s.TaskID)
		}
		selected[s] = true
	}
	for _, blocker := range b.Blockers {
		if err := firstError(dated("a blocker", blocker.Day), written("blocker for "+blocker.Day, blocker.Text)); err != nil {
			return err
		}
	}
	proposed := map[string]bool{}
	for _, p := range b.Prepared {
		record := "daily proposal for " + p.Day
		if err := firstError(dated("a daily proposal", p.Day), written(record, p.Text)); err != nil {
			return err
		}
		if proposed[p.Day] {
			return fmt.Errorf("two daily proposals for %s; each day has one", p.Day)
		}
		proposed[p.Day] = true
		// An empty source records that no earlier workday had entries.
		if p.Source != nil && *p.Source != "" && !isDate(*p.Source) {
			return fmt.Errorf("%s has source %q; sources must be YYYY-MM-DD", record, *p.Source)
		}
	}
	return nil
}

// firstError returns the first failed check, so a problem is reported on one line.
func firstError(checks ...error) error {
	for _, err := range checks {
		if err != nil {
			return err
		}
	}
	return nil
}
