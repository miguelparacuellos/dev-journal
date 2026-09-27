package journal

import (
	"fmt"
	"strings"
)

// State describes whether the meeting is in progress or when it closed.
func (r MeetingRecord) State() string {
	if r.Meeting.ClosedOn != "" {
		return "closed " + r.Meeting.ClosedOn
	}
	return "in progress"
}

// Text renders the complete meeting record as plain text for reading in the
// terminal: a heading line followed by Details.
func (r MeetingRecord) Text() string {
	return "O2O MEETING • " + r.Meeting.Day + " • " + r.State() + "\n\n" + r.Details()
}

// Details renders the notes, addressed topics, agreements, and follow-up tasks
// as plain text.
func (r MeetingRecord) Details() string {
	notes := "NOTES\nNone recorded.\n"
	if r.Meeting.Notes != "" {
		notes = "NOTES\n" + r.Meeting.Notes + "\n"
	}
	text := notes
	for _, section := range r.sections() {
		text += fmt.Sprintf("\n%s  %d\n", strings.ToUpper(section.title), len(section.items))
		if len(section.items) == 0 {
			text += "None recorded.\n"
		}
		for _, item := range section.items {
			if item.task != nil {
				item.text = "[" + item.task.State() + "] " + item.text
			}
			text += "- " + strings.ReplaceAll(item.text, "\n", "\n  ") + "\n"
		}
	}
	return text
}

// recordSection is a titled list in a meeting record, shared by the plain-text
// and Markdown renderings so both read the same way.
type recordSection struct {
	title string
	items []recordItem
}

// recordItem is one line of a section; task is set for follow-up tasks so each
// rendering can show its completion state.
type recordItem struct {
	text string
	task *Task
}

func (r MeetingRecord) sections() []recordSection {
	topics := recordSection{title: "Addressed topics"}
	for _, topic := range r.Addressed {
		topics.items = append(topics.items, recordItem{text: topicLine(topic)})
	}
	agreements := recordSection{title: "Agreements"}
	for _, agreement := range r.Agreements {
		agreements.items = append(agreements.items, recordItem{text: agreement.Text})
	}
	followUps := recordSection{title: "Follow-up tasks"}
	for i := range r.FollowUps {
		followUps.items = append(followUps.items, recordItem{text: r.FollowUps[i].Text, task: &r.FollowUps[i]})
	}
	return []recordSection{topics, agreements, followUps}
}

// topicLine names a topic with the day it was collected.
func topicLine(topic Topic) string { return topic.Text + " (collected " + topic.Day + ")" }
