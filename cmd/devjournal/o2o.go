package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/miguelparacuellos/dev-journal/journal"
)

const topicPlaceholder = "What do you want to raise with your tech lead?"

// meetingCapture describes a composer that writes into the meeting in progress.
type meetingCapture struct {
	key                       string
	placeholder, focus, saved string
	save                      func(app *journal.Journal, meetingID, text string) error
}

// meetingCaptures lists the meeting capture kinds by captureKind value.
var meetingCaptures = map[string]meetingCapture{
	"notes": {"w", "What was discussed?", "Meeting notes • kept with this meeting", "Notes saved • kept with this meeting",
		func(app *journal.Journal, id, text string) error { return app.SaveMeetingNotes(id, text) }},
	"agreement": {"r", "What did you agree on?", "O2O agreement • kept with this meeting", "Agreement recorded • kept with this meeting",
		func(app *journal.Journal, id, text string) error { _, err := app.RecordAgreement(id, text); return err }},
	"followup": {"f", "What will you follow up on?", "Follow-up task • also listed in Tasks", "Follow-up task created • also listed in Tasks",
		func(app *journal.Journal, id, text string) error {
			_, err := app.CreateFollowUpTask(id, text)
			return err
		}},
}

// meetingsPane is the O2O pane that browses every recorded meeting.
const meetingsPane = "meetings"

// Rows used around the O2O lists at the compact layout: the header, composer,
// list heading and range footer (listChrome), plus the three meeting summary
// lines, or the meetings footer line when no meeting is in progress.
const (
	listChrome      = 18
	meetingsChrome  = 13
	summaryRows     = 3
	meetingsSummary = 1
)

// agenda lists the topics shown in O2O: open topics and, during a meeting, the
// topics it addressed, in capture order so marking a topic never moves its row.
func (m model) agenda() []journal.Topic {
	current, _ := m.app.CurrentMeeting()
	return m.app.Agenda(current.Meeting.ID)
}

// saveMeetingCapture saves the composer into the meeting in progress.
func (m model) saveMeetingCapture(text string) error {
	current, open := m.app.CurrentMeeting()
	if !open {
		return fmt.Errorf("no meeting in progress")
	}
	return meetingCaptures[m.captureKind].save(m.app, current.Meeting.ID, text)
}

// unchangedNotes reports whether the composer only holds the saved meeting
// notes, so reading them with w leaves no draft behind.
func (m model) unchangedNotes() bool {
	current, open := m.app.CurrentMeeting()
	return m.captureKind == "notes" && open && m.editor.Value() == current.Meeting.Notes
}

// updateTopics handles the O2O keys. It reports false only for keys shared with
// other views (capture, help, discard, quit, and Esc from topics); every other key
// is consumed so Today and Tasks actions never act on O2O topics or meetings.
// closing is true when the previous key asked to close the meeting.
func (m model) updateTopics(key string, closing bool) (model, tea.Cmd, bool) {
	if m.o2oPane == meetingsPane {
		return m.updateMeetings(key)
	}
	topics := m.agenda()
	current, open := m.app.CurrentMeeting()
	switch key {
	case "up", "k":
		m.selected = max(0, m.selected-1)
	case "down", "j":
		m.selected = min(max(0, len(topics)-1), m.selected+1)
	case "enter":
		if len(topics) > 0 {
			m.openText("topic-detail", topics[m.selected].Text)
		}
	case "s":
		if open {
			m.status = "Meeting in progress since " + current.Meeting.Day + " • c closes it"
			break
		}
		_, err := m.app.StartMeeting(m.today)
		m.actionStatus(err, "Meeting started • a marks topics addressed; w writes notes")
	case "a":
		if !open {
			m.status = "No meeting in progress • s starts one to mark topics addressed"
			break
		}
		if len(topics) == 0 {
			break
		}
		topic := topics[m.selected]
		addressed := topic.AddressedIn == ""
		err := m.app.AddressTopic(current.Meeting.ID, topic.ID, addressed)
		if addressed {
			m.actionStatus(err, "Addressed in this meeting • a reopens it")
		} else {
			m.actionStatus(err, "Open again • it stays for the next meeting")
		}
	case "w", "r", "f":
		if !open {
			m.status = "No meeting in progress • s starts one"
			break
		}
		if m.editor.Value() != "" {
			m.status = "Draft retained • n resumes; x discards before starting another"
			break
		}
		for kind, capture := range meetingCaptures {
			if capture.key == key {
				m.captureKind = kind
				m.editor.Placeholder = capture.placeholder
			}
		}
		if key == "w" {
			m.editor.SetValue(current.Meeting.Notes)
		}
		m.mode = "capture"
		return m, m.editor.Focus(), true
	case "c":
		if !open {
			m.status = "No meeting in progress • s starts one"
			break
		}
		if m.editor.Value() != "" {
			m.status = "Draft retained • save or discard it before closing the meeting"
			break
		}
		if !closing {
			m.confirmClose = true
			m.status = "Close this meeting? c confirms • any other key keeps it open"
			break
		}
		remaining := len(m.app.OpenTopics())
		err := m.app.CloseMeeting(current.Meeting.ID, m.today)
		m.actionStatus(err, fmt.Sprintf("Meeting closed • %d open %s kept for the next one", remaining, plural(remaining, "topic", "topics")))
		m.selected = min(m.selected, max(0, len(m.agenda())-1))
	case "v":
		if !open {
			m.status = "No meeting in progress • m browses past meetings"
			break
		}
		m.openText("meeting-detail", current.Text())
	case "m":
		m.o2oPane = meetingsPane
		m.selected = 0
	case "n", "x", "?", "q", "ctrl+c", "esc":
		return m, nil, false
	}
	return m, nil, true
}

// updateMeetings handles the meeting history pane.
func (m model) updateMeetings(key string) (model, tea.Cmd, bool) {
	meetings := m.app.Meetings()
	switch key {
	case "up", "k":
		m.selected = max(0, m.selected-1)
	case "down", "j":
		m.selected = min(max(0, len(meetings)-1), m.selected+1)
	case "enter":
		if len(meetings) > 0 {
			m.openText("meeting-detail", meetings[m.selected].Text())
		}
	case "esc", "m", "s", "n":
		// These leave the history for the topics; s then starts a meeting and n
		// falls through to topic capture.
		m.o2oPane = ""
		m.selected = 0
		if key == "s" {
			return m.updateTopics(key, false)
		}
		return m, nil, key != "n"
	case "x", "?", "q", "ctrl+c":
		return m, nil, false
	}
	return m, nil, true
}

// o2oHeading is the O2O line under the view title.
func (m model) o2oHeading() string {
	if m.o2oPane == meetingsPane {
		return "Meetings • notes, agreements, and follow-ups are kept after closing"
	}
	if current, open := m.app.CurrentMeeting(); open {
		return "Meeting in progress • started " + current.Meeting.Day + " • c closes it"
	}
	return "One to one topics • collected any day, open until addressed"
}

// o2oHints are the O2O browsing key hints.
func (m model) o2oHints(width int) string {
	if m.o2oPane == meetingsPane {
		return fitHints(width, "arrows Select · Enter Read record · Esc Topics · s Start meeting · ? Help · q Quit",
			"arrows Select · Enter Read · Esc Topics · s Start meeting · q Quit")
	}
	if _, open := m.app.CurrentMeeting(); open {
		return fitHints(width, "a Addressed · w Notes · r Agreement · f Follow-up · v Record · c Close · n Topic · ? Help · q Quit",
			"a Address · w Notes · r Agree · f Follow · v Record · c Close · ? Help")
	}
	return fitHints(width, "n New topic · s Start meeting · m Meetings · Enter Read · t Today · Tab Tasks · ? Help · q Quit",
		"n Topic · s Start meeting · m Meetings · Enter Read · ? Help · q Quit")
}

// listWindow returns the visible range of a list that keeps selected in view.
func listWindow(selected, available, total int) (int, int) {
	start := max(0, selected-available/2)
	return start, min(total, start+available)
}

// withPreview places a list beside a preview for wide layouts, trimming the
// preview to the list's height and pointing to the key that reads the rest.
func (m model) withPreview(list string, listWidth, width, available int, right, more string) string {
	previewWidth := width - listWidth - 5
	right = ansi.Wrap(right, previewWidth, "")
	if lines := strings.Split(right, "\n"); len(lines) > available+1 {
		right = strings.Join(lines[:available], "\n") + "\n[" + more + " to read more]"
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(listWidth+2).Render(list), "  ", lipgloss.NewStyle().Width(previewWidth+1).Render(right))
}

// listWidthFor is the list column width: full width, or half beside a preview.
func (m model) listWidthFor(width int) int {
	if m.width >= wideLayoutWidth {
		return width/2 - 3
	}
	return width
}

// topicsBody lists the agenda below the composer: open topics for meeting
// preparation, or during a meeting the topics with their addressed state.
func (m model) topicsBody(width int) string {
	current, open := m.app.CurrentMeeting()
	topics := m.agenda()
	var header string
	if open {
		header = m.style("title").Render("MEETING TOPICS") + m.style("muted").Render(fmt.Sprintf(" • %d of %d addressed", len(current.Addressed), len(topics)))
	} else {
		header = m.style("title").Render(fmt.Sprintf("OPEN TOPICS  %d", len(topics))) + m.style("muted").Render(" • for the next one to one")
	}
	footer := m.meetingsSummary()
	if open {
		footer = m.meetingSummary(current, width)
	}
	if len(topics) == 0 {
		empty := "\n\nNothing to raise yet. n captures a topic whenever it comes up;\nit stays open across days until it is addressed in a one to one.\n"
		if open {
			empty = "\nNo topics to consult. n adds one; w, r, and f still record the meeting.\n"
		}
		return header + empty + footer
	}
	listWidth := m.listWidthFor(width)
	available := max(1, m.height-listChrome-meetingsSummary)
	if open && m.width < wideLayoutWidth {
		available = max(1, m.height-listChrome-summaryRows)
	}
	start, end := listWindow(m.selected, available, len(topics))
	list := header + "\n"
	for i := start; i < end; i++ {
		marker := "  "
		if i == m.selected {
			marker = "> "
		}
		if open {
			if topics[i].AddressedIn != "" {
				marker += "[addressed] "
			} else {
				marker += "[open]      "
			}
		}
		line := ansi.Truncate(marker+topics[i].Day+"  "+oneLine(topics[i].Text), listWidth, "...")
		if i == m.selected {
			line = m.style("accent").Render(line)
		}
		list += line + "\n"
	}
	action := "Enter reads the complete topic"
	if open {
		action = "a marks addressed · Enter topic · v record"
	}
	list += m.style("muted").Render(fmt.Sprintf("%d-%d of %d · %s", start+1, end, len(topics), action))
	if m.width < wideLayoutWidth {
		return list + "\n" + footer
	}
	if open {
		right := m.style("title").Render("MEETING RECORD") + m.style("muted").Render(" • v reads it all") + "\n" + displayText(current.Details())
		return m.withPreview(list, listWidth, width, available, right, "v")
	}
	selected := topics[min(m.selected, len(topics)-1)]
	right := m.style("title").Render("SELECTED TOPIC") + m.style("muted").Render(" • captured "+selected.Day) + "\n" + displayText(selected.Text)
	return m.withPreview(list, listWidth, width, available, right, "Enter") + "\n" + footer
}

// meetingSummary condenses the meeting record into three lines for compact layouts.
func (m model) meetingSummary(record journal.MeetingRecord, width int) string {
	notes := "None yet • w writes"
	if record.Meeting.Notes != "" {
		notes = oneLine(record.Meeting.Notes)
	}
	agreements := "None yet • r records"
	if n := len(record.Agreements); n > 0 {
		agreements = fmt.Sprintf("%d • %s", n, oneLine(record.Agreements[n-1].Text))
	}
	followUps := "None yet • f creates"
	if n := len(record.FollowUps); n > 0 {
		followUps = fmt.Sprintf("%d • %s", n, oneLine(record.FollowUps[n-1].Text))
	}
	line := func(label, text string) string {
		return m.style("title").Render(label) + " " + ansi.Truncate(text, width-ansi.StringWidth(label)-1, "...") + "\n"
	}
	return line("NOTES", notes) + line("AGREEMENTS", agreements) + line("FOLLOW-UPS", followUps)
}

// meetingsSummary points to meeting history when no meeting is in progress.
func (m model) meetingsSummary() string {
	meetings := m.app.Meetings()
	if len(meetings) == 0 {
		return m.style("muted").Render("No meetings recorded yet • s starts one when you meet") + "\n"
	}
	return m.style("muted").Render(fmt.Sprintf("%d %s recorded • last on %s • m browses", len(meetings), plural(len(meetings), "meeting", "meetings"), meetings[0].Meeting.Day)) + "\n"
}

// meetingsBody browses every recorded meeting, newest first.
func (m model) meetingsBody(width int) string {
	meetings := m.app.Meetings()
	header := m.style("title").Render(fmt.Sprintf("MEETINGS  %d", len(meetings))) + m.style("muted").Render(" • newest first")
	if len(meetings) == 0 {
		return header + "\n\nNo meetings yet. s starts one; Esc returns to topics.\n"
	}
	listWidth := m.listWidthFor(width)
	available := max(1, m.height-meetingsChrome)
	start, end := listWindow(m.selected, available, len(meetings))
	list := header + "\n"
	for i := start; i < end; i++ {
		marker := "  "
		if i == m.selected {
			marker = "> "
		}
		line := ansi.Truncate(marker+meetingSummaryLine(meetings[i]), listWidth, "...")
		if i == m.selected {
			line = m.style("accent").Render(line)
		}
		list += line + "\n"
	}
	list += m.style("muted").Render(fmt.Sprintf("%d-%d of %d · Enter reads the complete record", start+1, end, len(meetings)))
	if m.width < wideLayoutWidth {
		return list
	}
	right := m.style("title").Render("SELECTED MEETING") + "\n" + displayText(meetings[min(m.selected, len(meetings)-1)].Text())
	return m.withPreview(list, listWidth, width, available, right, "Enter")
}

// meetingSummaryLine is a one-line description of a meeting for lists.
func meetingSummaryLine(record journal.MeetingRecord) string {
	return fmt.Sprintf("%s  %s • %d addressed • %d %s • %d %s", record.Meeting.Day, record.State(),
		len(record.Addressed), len(record.Agreements), plural(len(record.Agreements), "agreement", "agreements"),
		len(record.FollowUps), plural(len(record.FollowUps), "follow-up", "follow-ups"))
}

// oneLine flattens text for single-line list rows.
func oneLine(text string) string { return strings.ReplaceAll(displayText(text), "\n", " / ") }

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
