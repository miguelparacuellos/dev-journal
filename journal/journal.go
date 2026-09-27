// Package journal is the public application boundary shared by the CLI and TUI.
package journal

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

type Entry struct {
	ID      string `json:"id"`
	Workday string `json:"workday"`
	Text    string `json:"text"`
}

// Task is an action to complete in the future. A follow-up task created during
// an O2O meeting names that meeting in MeetingID and otherwise behaves the same.
type Task struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
	MeetingID string `json:"meeting_id,omitempty"`
}

// State is the word for the task's completion state: "open" or "done".
func (t Task) State() string {
	if t.Completed {
		return "done"
	}
	return "open"
}

type PlanSelection struct {
	Day    string `json:"day"`
	TaskID string `json:"task_id"`
}
type Blocker struct {
	Day  string `json:"day"`
	Text string `json:"text"`
}

// DailyProposal is the saved personal preparation for the daily on Day. Source
// is the recent-work workday Daily showed when it was saved, empty when no
// earlier workday had entries, and nil for proposals saved before sources were
// recorded.
type DailyProposal struct {
	Day    string  `json:"day"`
	Text   string  `json:"text"`
	Source *string `json:"source,omitempty"`
}

// Topic is an O2O topic: an item to discuss in a one to one, captured on Day.
// It stays open until AddressedIn names the meeting that addressed it.
type Topic struct {
	ID          string `json:"id"`
	Day         string `json:"day"`
	Text        string `json:"text"`
	AddressedIn string `json:"addressed_in,omitempty"`
}

// Meeting is an O2O meeting started on Day. It is in progress until ClosedOn
// is set; a closed meeting keeps its notes and related records unchanged.
type Meeting struct {
	ID       string `json:"id"`
	Day      string `json:"day"`
	Notes    string `json:"notes,omitempty"`
	ClosedOn string `json:"closed_on,omitempty"`
}

// Agreement is an O2O agreement: a conclusion or commitment recorded in a meeting.
type Agreement struct {
	ID        string `json:"id"`
	MeetingID string `json:"meeting_id"`
	Text      string `json:"text"`
}

// MeetingRecord is a meeting together with the records associated with it, each
// in capture order. Follow-up tasks show their current completion state.
type MeetingRecord struct {
	Meeting    Meeting
	Addressed  []Topic
	Agreements []Agreement
	FollowUps  []Task
}
type Data struct {
	Version    int             `json:"version"`
	Blockers   []Blocker       `json:"blockers,omitempty"`
	Prepared   []DailyProposal `json:"prepared,omitempty"`
	Entries    []Entry         `json:"entries"`
	Tasks      []Task          `json:"tasks,omitempty"`
	Plan       []PlanSelection `json:"plan,omitempty"`
	Topics     []Topic         `json:"topics,omitempty"`
	Meetings   []Meeting       `json:"meetings,omitempty"`
	Agreements []Agreement     `json:"agreements,omitempty"`
}
type Journal struct {
	path string
	data Data
}

func Open(path string) (*Journal, error) { j := &Journal{path: path}; err := j.reload(); return j, err }
func (j *Journal) reload() error {
	j.data = Data{Version: 1, Entries: []Entry{}}
	b, err := os.ReadFile(j.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, &j.data); err != nil {
		return fmt.Errorf("journal cannot be read: %w", err)
	}
	if j.data.Version != 1 {
		return fmt.Errorf("unsupported journal version: %d", j.data.Version)
	}
	return nil
}
func (j *Journal) Entries(day string) []Entry {
	result := []Entry{}
	for _, e := range j.data.Entries {
		if e.Workday == day {
			result = append(result, e)
		}
	}
	return result
}
func (j *Journal) Workdays() []string {
	set := map[string]bool{}
	for _, e := range j.data.Entries {
		set[e.Workday] = true
	}
	days := []string{}
	for day := range set {
		days = append(days, day)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(days)))
	return days
}
func validate(day, text string) error {
	if _, err := time.Parse("2006-01-02", day); err != nil {
		return errors.New("workday must be YYYY-MM-DD")
	}
	if strings.TrimSpace(text) == "" {
		return errors.New("entry cannot be empty")
	}
	return nil
}
func (j *Journal) Capture(day, text string) (Entry, error) {
	if err := validate(day, text); err != nil {
		return Entry{}, err
	}
	id, err := newID()
	if err != nil {
		return Entry{}, err
	}
	e := Entry{id, day, text}
	err = j.change(func() error { j.data.Entries = append(j.data.Entries, e); return nil })
	return e, err
}
func (j *Journal) Correct(id, text string) error {
	if strings.TrimSpace(text) == "" {
		return errors.New("entry cannot be empty")
	}
	return j.change(func() error {
		for i := range j.data.Entries {
			if j.data.Entries[i].ID == id {
				j.data.Entries[i].Text = text
				return nil
			}
		}
		return errors.New("entry not found")
	})
}
func (j *Journal) change(update func() error) (err error) {
	if err := os.MkdirAll(filepath.Dir(j.path), 0700); err != nil {
		return err
	}
	lock, err := os.OpenFile(j.lockPath(), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	if err = j.reload(); err != nil {
		return err
	}
	previous := j.data
	previous.Prepared = append([]DailyProposal(nil), j.data.Prepared...)
	previous.Blockers = append([]Blocker(nil), j.data.Blockers...)
	previous.Entries = append([]Entry(nil), j.data.Entries...)
	previous.Tasks = append([]Task(nil), j.data.Tasks...)
	previous.Plan = append([]PlanSelection(nil), j.data.Plan...)
	previous.Topics = append([]Topic(nil), j.data.Topics...)
	previous.Meetings = append([]Meeting(nil), j.data.Meetings...)
	previous.Agreements = append([]Agreement(nil), j.data.Agreements...)
	defer func() {
		if err != nil {
			j.data = previous
		}
	}()
	if err = update(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(j.data, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(j.path, ".journal-*", b)
}

// lockPath is the advisory lock file that serializes writes to the journal.
func (j *Journal) lockPath() string { return j.path + ".lock" }

// writeFile durably replaces path with b: it writes a temporary file named by
// the tempPattern in the same directory, syncs it, renames it over path, and
// syncs the directory.
func writeFile(path, tempPattern string, b []byte) (err error) {
	tmp, err := os.CreateTemp(filepath.Dir(path), tempPattern)
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(b); err == nil {
		err = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmp.Name(), path)
	}
	if err != nil {
		return err
	}
	directory, openErr := os.Open(filepath.Dir(path))
	if openErr != nil {
		return openErr
	}
	defer directory.Close()
	return directory.Sync()
}

// Tasks returns both open and completed tasks in capture order.
func (j *Journal) Tasks() []Task                        { return append([]Task{}, j.data.Tasks...) }
func (j *Journal) CreateTask(text string) (Task, error) { return j.createTask(text, "") }

// CreateFollowUpTask creates a task associated with the meeting in progress.
// It appears in Tasks and supports the same planning and completion.
func (j *Journal) CreateFollowUpTask(meetingID, text string) (Task, error) {
	if meetingID == "" {
		return Task{}, errors.New("meeting not found")
	}
	return j.createTask(text, meetingID)
}

func (j *Journal) createTask(text, meetingID string) (Task, error) {
	if strings.TrimSpace(text) == "" {
		return Task{}, errors.New("task cannot be empty")
	}
	id, err := newID()
	if err != nil {
		return Task{}, err
	}
	task := Task{ID: id, Text: text, MeetingID: meetingID}
	err = j.change(func() error {
		if meetingID != "" {
			if _, err := j.openMeeting(meetingID); err != nil {
				return err
			}
		}
		j.data.Tasks = append(j.data.Tasks, task)
		return nil
	})
	return task, err
}

func (j *Journal) Plan(day string) []Task {
	result := []Task{}
	for _, selection := range j.data.Plan {
		if selection.Day == day {
			for _, task := range j.data.Tasks {
				if task.ID == selection.TaskID {
					result = append(result, task)
					break
				}
			}
		}
	}
	return result
}
func (j *Journal) PlanTask(day, id string) error {
	if _, err := time.Parse("2006-01-02", day); err != nil {
		return errors.New("plan date must be YYYY-MM-DD")
	}
	return j.change(func() error {
		found := false
		for _, task := range j.data.Tasks {
			if task.ID == id {
				if task.Completed {
					return errors.New("completed tasks cannot be planned")
				}
				found = true
				break
			}
		}
		if !found {
			return errors.New("task not found")
		}
		for _, selection := range j.data.Plan {
			if selection.Day == day && selection.TaskID == id {
				return nil
			}
		}
		j.data.Plan = append(j.data.Plan, PlanSelection{Day: day, TaskID: id})
		return nil
	})
}

func (j *Journal) CompleteTask(id string) error {
	return j.change(func() error {
		for i := range j.data.Tasks {
			if j.data.Tasks[i].ID == id {
				j.data.Tasks[i].Completed = true
				return nil
			}
		}
		return errors.New("task not found")
	})
}

func (j *Journal) UnplanTask(day, id string) error {
	if _, err := time.Parse("2006-01-02", day); err != nil {
		return errors.New("plan date must be YYYY-MM-DD")
	}
	return j.change(func() error {
		selections := []PlanSelection{}
		for _, selection := range j.data.Plan {
			if selection.Day != day || selection.TaskID != id {
				selections = append(selections, selection)
			}
		}
		j.data.Plan = selections
		return nil
	})
}

func newID() (string, error) {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	return hex.EncodeToString(id), nil
}

// RecentWork returns the latest recorded workday strictly before day.
func (j *Journal) RecentWork(day string) (string, []Entry) {
	for _, recorded := range j.Workdays() {
		if recorded < day {
			return recorded, j.Entries(recorded)
		}
	}
	return "", []Entry{}
}

func (j *Journal) Blockers(day string) []Blocker {
	result := []Blocker{}
	for _, blocker := range j.data.Blockers {
		if blocker.Day == day {
			result = append(result, blocker)
		}
	}
	return result
}
func (j *Journal) RecordBlocker(day, text string) error {
	if err := validate(day, text); err != nil {
		return err
	}
	return j.change(func() error { j.data.Blockers = append(j.data.Blockers, Blocker{Day: day, Text: text}); return nil })
}

// PrepareDaily copies explicitly selected recent work and current sources into editable text.
// It never replaces a saved personal preparation.
func (j *Journal) PrepareDaily(day string, selected []string) (string, error) {
	if _, err := time.Parse("2006-01-02", day); err != nil {
		return "", errors.New("daily date must be YYYY-MM-DD")
	}
	source, entries := j.RecentWork(day)
	chosen := map[string]bool{}
	for _, id := range selected {
		found := false
		for _, entry := range entries {
			if entry.ID == id {
				found = true
				chosen[id] = true
				break
			}
		}
		if !found {
			return "", errors.New("selected progress is not from recent work")
		}
	}
	title := "Recent work"
	if source != "" {
		title += " (" + source + ")"
	}
	text := title
	for _, entry := range entries {
		if chosen[entry.ID] {
			text += "\n- " + entry.Text
		}
	}
	text += "\n\nToday's plan"
	for _, task := range j.Plan(day) {
		text += "\n- " + task.Text
	}
	text += "\n\nBlockers"
	for _, blocker := range j.Blockers(day) {
		text += "\n- " + blocker.Text
	}
	return text, nil
}
func (j *Journal) Daily(day string) (string, bool) {
	for _, proposal := range j.data.Prepared {
		if proposal.Day == day {
			return proposal.Text, true
		}
	}
	return "", false
}
func (j *Journal) SaveDaily(day, text string) error {
	if err := validate(day, text); err != nil {
		return err
	}
	return j.change(func() error {
		source, _ := j.RecentWork(day)
		for i := range j.data.Prepared {
			if j.data.Prepared[i].Day == day {
				j.data.Prepared[i].Text = text
				j.data.Prepared[i].Source = &source
				return nil
			}
		}
		j.data.Prepared = append(j.data.Prepared, DailyProposal{Day: day, Text: text, Source: &source})
		return nil
	})
}

// OpenTopics returns the O2O topics still waiting to be discussed, in capture order.
// Topics stay open across days and restarts until a meeting addresses them;
// no meeting is needed to collect them.
func (j *Journal) OpenTopics() []Topic { return j.Agenda("") }

// Agenda returns the topics to consult during a meeting, in capture order: every
// open topic plus those the meeting has addressed. Without a meeting ID it
// returns only the open topics.
func (j *Journal) Agenda(meetingID string) []Topic {
	result := []Topic{}
	for _, topic := range j.data.Topics {
		if topic.AddressedIn == "" || (meetingID != "" && topic.AddressedIn == meetingID) {
			result = append(result, topic)
		}
	}
	return result
}

// AddTopic collects an O2O topic on the given day without starting a meeting.
func (j *Journal) AddTopic(day, text string) (Topic, error) {
	if _, err := time.Parse("2006-01-02", day); err != nil {
		return Topic{}, errors.New("topic date must be YYYY-MM-DD")
	}
	if strings.TrimSpace(text) == "" {
		return Topic{}, errors.New("topic cannot be empty")
	}
	id, err := newID()
	if err != nil {
		return Topic{}, err
	}
	topic := Topic{ID: id, Day: day, Text: text}
	err = j.change(func() error { j.data.Topics = append(j.data.Topics, topic); return nil })
	return topic, err
}

// CurrentMeeting returns the O2O meeting in progress, if any.
func (j *Journal) CurrentMeeting() (MeetingRecord, bool) {
	for _, meeting := range j.data.Meetings {
		if meeting.ClosedOn == "" {
			return j.record(meeting), true
		}
	}
	return MeetingRecord{}, false
}

func (j *Journal) record(meeting Meeting) MeetingRecord {
	record := MeetingRecord{Meeting: meeting, Addressed: []Topic{}, Agreements: []Agreement{}, FollowUps: []Task{}}
	for _, topic := range j.data.Topics {
		if topic.AddressedIn == meeting.ID {
			record.Addressed = append(record.Addressed, topic)
		}
	}
	for _, agreement := range j.data.Agreements {
		if agreement.MeetingID == meeting.ID {
			record.Agreements = append(record.Agreements, agreement)
		}
	}
	for _, task := range j.data.Tasks {
		if task.MeetingID == meeting.ID {
			record.FollowUps = append(record.FollowUps, task)
		}
	}
	return record
}

// StartMeeting begins an O2O meeting on day. Only one meeting can be in progress.
func (j *Journal) StartMeeting(day string) (Meeting, error) {
	if _, err := time.Parse("2006-01-02", day); err != nil {
		return Meeting{}, errors.New("meeting date must be YYYY-MM-DD")
	}
	id, err := newID()
	if err != nil {
		return Meeting{}, err
	}
	meeting := Meeting{ID: id, Day: day}
	err = j.change(func() error {
		if _, open := j.CurrentMeeting(); open {
			return errors.New("an O2O meeting is already in progress")
		}
		j.data.Meetings = append(j.data.Meetings, meeting)
		return nil
	})
	return meeting, err
}

// SaveMeetingNotes replaces the notes of the meeting in progress. Blank notes
// clear them.
func (j *Journal) SaveMeetingNotes(meetingID, notes string) error {
	if strings.TrimSpace(notes) == "" {
		notes = ""
	}
	return j.change(func() error {
		meeting, err := j.openMeeting(meetingID)
		if err != nil {
			return err
		}
		meeting.Notes = notes
		return nil
	})
}

// openMeeting finds a meeting that can still change. Call it inside change().
func (j *Journal) openMeeting(id string) (*Meeting, error) {
	for i := range j.data.Meetings {
		if j.data.Meetings[i].ID == id {
			if j.data.Meetings[i].ClosedOn != "" {
				return nil, errors.New("the meeting is closed")
			}
			return &j.data.Meetings[i], nil
		}
	}
	return nil, errors.New("meeting not found")
}

// AddressTopic marks a topic addressed in the meeting in progress, or reopens it
// when addressed is false. Topics addressed in earlier meetings cannot change.
func (j *Journal) AddressTopic(meetingID, topicID string, addressed bool) error {
	return j.change(func() error {
		if _, err := j.openMeeting(meetingID); err != nil {
			return err
		}
		for i := range j.data.Topics {
			topic := &j.data.Topics[i]
			if topic.ID != topicID {
				continue
			}
			if topic.AddressedIn != "" && topic.AddressedIn != meetingID {
				return errors.New("the topic was addressed in an earlier meeting")
			}
			topic.AddressedIn = ""
			if addressed {
				topic.AddressedIn = meetingID
			}
			return nil
		}
		return errors.New("topic not found")
	})
}

// RecordAgreement records an O2O agreement in the meeting in progress.
func (j *Journal) RecordAgreement(meetingID, text string) (Agreement, error) {
	if strings.TrimSpace(text) == "" {
		return Agreement{}, errors.New("agreement cannot be empty")
	}
	id, err := newID()
	if err != nil {
		return Agreement{}, err
	}
	agreement := Agreement{ID: id, MeetingID: meetingID, Text: text}
	err = j.change(func() error {
		if _, err := j.openMeeting(meetingID); err != nil {
			return err
		}
		j.data.Agreements = append(j.data.Agreements, agreement)
		return nil
	})
	return agreement, err
}

// CloseMeeting closes the meeting in progress on day. Its notes, addressed topics,
// agreements, and follow-up tasks are preserved; unaddressed topics stay open.
func (j *Journal) CloseMeeting(meetingID, day string) error {
	if _, err := time.Parse("2006-01-02", day); err != nil {
		return errors.New("closing date must be YYYY-MM-DD")
	}
	return j.change(func() error {
		meeting, err := j.openMeeting(meetingID)
		if err != nil {
			return err
		}
		if day < meeting.Day {
			return errors.New("a meeting cannot close before it started")
		}
		meeting.ClosedOn = day
		return nil
	})
}

// Meetings returns every O2O meeting with its records, newest first.
func (j *Journal) Meetings() []MeetingRecord {
	result := []MeetingRecord{}
	for i := len(j.data.Meetings) - 1; i >= 0; i-- {
		result = append(result, j.record(j.data.Meetings[i]))
	}
	return result
}
