package journal

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Markdown renders the whole journal as a readable Markdown document exported on
// the given day: the daily log by workday, tasks, open O2O topics, and O2O
// meetings. It only reads the journal.
func (j *Journal) Markdown(exportedOn string) string {
	var doc strings.Builder
	doc.WriteString("# Dev Journal\n\nExported on " + exportedOn + " from the local journal. Dates use YYYY-MM-DD.\n")
	d := j.data
	if len(d.Entries)+len(d.Plan)+len(d.Blockers)+len(d.Prepared)+len(d.Tasks)+len(d.Topics)+len(d.Meetings) == 0 {
		doc.WriteString("\nThe journal is empty: no entries, plans, blockers, daily proposals, tasks, O2O topics, or meetings have been recorded yet.\n")
		return doc.String()
	}
	j.writeDailyLog(&doc)
	j.writeTasks(&doc)
	j.writeTopics(&doc)
	j.writeMeetings(&doc)
	return doc.String()
}

// ExportMarkdown saves Markdown(exportedOn) to path, replacing any earlier
// export there. The stored journal is left unchanged, and it is never used as
// the destination.
func (j *Journal) ExportMarkdown(path, exportedOn string) error {
	target, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	journalPath, err := filepath.Abs(j.path)
	if err != nil {
		return err
	}
	lockPath, err := filepath.Abs(j.lockPath())
	if err != nil {
		return err
	}
	if target == journalPath || target == lockPath {
		return errors.New("choose a different file; the export cannot replace the journal")
	}
	return writeFile(target, ".export-*", []byte(j.Markdown(exportedOn)))
}

// writeDailyLog writes every dated record grouped by workday, newest first.
func (j *Journal) writeDailyLog(doc *strings.Builder) {
	set := map[string]bool{}
	for _, e := range j.data.Entries {
		set[e.Workday] = true
	}
	for _, s := range j.data.Plan {
		set[s.Day] = true
	}
	for _, b := range j.data.Blockers {
		set[b.Day] = true
	}
	for _, p := range j.data.Prepared {
		set[p.Day] = true
	}
	days := []string{}
	for day := range set {
		days = append(days, day)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(days)))
	doc.WriteString("\n## Daily log\n\nWorkdays newest first. Entries are completed work; today's plan lists the tasks selected for that day with their current state.\n")
	if len(days) == 0 {
		doc.WriteString("\nNo workdays recorded.\n")
	}
	for _, day := range days {
		doc.WriteString("\n### " + day + "\n")
		if entries := j.Entries(day); len(entries) > 0 {
			doc.WriteString("\n#### Entries\n\n")
			for _, entry := range entries {
				doc.WriteString(listItem(entry.Text))
			}
		}
		if plan := j.Plan(day); len(plan) > 0 {
			doc.WriteString("\n#### Today's plan\n\n")
			for _, task := range plan {
				doc.WriteString(taskItem(task, j.followUpNote(task)))
			}
		}
		if blockers := j.Blockers(day); len(blockers) > 0 {
			doc.WriteString("\n#### Blockers\n\n")
			for _, blocker := range blockers {
				doc.WriteString(listItem(blocker.Text))
			}
		}
		for _, proposal := range j.data.Prepared {
			if proposal.Day != day {
				continue
			}
			described := "none; no earlier workday had entries"
			if proposal.Source == nil {
				// Proposals saved before sources were recorded only have the
				// current reference.
				described = "not recorded with this proposal; no earlier workday has entries now"
				if current, _ := j.RecentWork(day); current != "" {
					described = "not recorded with this proposal; currently " + current + ", the latest workday with entries before " + day
				}
			} else if *proposal.Source != "" {
				described = *proposal.Source + ", the latest workday with entries when it was saved"
			}
			doc.WriteString("\n#### Daily proposal\n\nSaved personal preparation for the daily. Recent work source: " + described + ".\n\n")
			doc.WriteString(quote(proposal.Text))
		}
	}
}

// writeTasks writes open then completed tasks with their planned dates and origin.
func (j *Journal) writeTasks(doc *strings.Builder) {
	doc.WriteString("\n## Tasks\n")
	if len(j.data.Tasks) == 0 {
		doc.WriteString("\nNo tasks recorded.\n")
		return
	}
	for _, group := range []struct {
		title     string
		completed bool
	}{{"Open tasks", false}, {"Completed tasks", true}} {
		tasks := []Task{}
		for _, task := range j.data.Tasks {
			if task.Completed == group.completed {
				tasks = append(tasks, task)
			}
		}
		fmt.Fprintf(doc, "\n### %s (%d)\n\n", group.title, len(tasks))
		if len(tasks) == 0 {
			doc.WriteString("None.\n")
		}
		for _, task := range tasks {
			notes := []string{}
			if days := j.plannedDays(task.ID); len(days) > 0 {
				notes = append(notes, "planned "+strings.Join(days, ", "))
			}
			if note := j.followUpNote(task); note != "" {
				notes = append(notes, note)
			}
			doc.WriteString(taskItem(task, notes...))
		}
	}
}

// writeTopics writes the O2O topics still open for the next one to one.
func (j *Journal) writeTopics(doc *strings.Builder) {
	topics := j.OpenTopics()
	fmt.Fprintf(doc, "\n## Open O2O topics (%d)\n\nWaiting for the next one to one. Addressed topics are listed with the meeting that addressed them.\n\n", len(topics))
	if len(topics) == 0 {
		doc.WriteString("None.\n")
	}
	for _, topic := range topics {
		doc.WriteString(listItem(topicLine(topic)))
	}
}

// writeMeetings writes every O2O meeting record, newest first.
func (j *Journal) writeMeetings(doc *strings.Builder) {
	meetings := j.Meetings()
	fmt.Fprintf(doc, "\n## O2O meetings (%d)\n\nNewest first. Follow-up tasks show their current state.\n", len(meetings))
	if len(meetings) == 0 {
		doc.WriteString("\nNo meetings recorded.\n")
	}
	for _, record := range meetings {
		doc.WriteString("\n### O2O meeting · " + record.Meeting.Day + " · " + record.State() + "\n\n#### Notes\n\n")
		if record.Meeting.Notes == "" {
			doc.WriteString("None recorded.\n")
		} else {
			doc.WriteString(hardBreaks(record.Meeting.Notes, "") + "\n")
		}
		for _, section := range record.sections() {
			fmt.Fprintf(doc, "\n#### %s (%d)\n\n", section.title, len(section.items))
			if len(section.items) == 0 {
				doc.WriteString("None recorded.\n")
			}
			for _, item := range section.items {
				if item.task != nil {
					doc.WriteString(taskItem(*item.task))
				} else {
					doc.WriteString(listItem(item.text))
				}
			}
		}
	}
}

// plannedDays lists the days a task was selected for, oldest first.
func (j *Journal) plannedDays(taskID string) []string {
	days := []string{}
	for _, selection := range j.data.Plan {
		if selection.TaskID == taskID {
			days = append(days, selection.Day)
		}
	}
	sort.Strings(days)
	return days
}

// followUpNote names the meeting a follow-up task came from.
func (j *Journal) followUpNote(task Task) string {
	if task.MeetingID == "" {
		return ""
	}
	for _, meeting := range j.data.Meetings {
		if meeting.ID == task.MeetingID {
			return "O2O follow-up from the " + meeting.Day + " meeting"
		}
	}
	return "O2O follow-up"
}

// taskItem is a task list item with its state spelled out, so it reads the
// same with or without checkbox rendering.
func taskItem(task Task, notes ...string) string {
	box := "[ ] Open: "
	if task.Completed {
		box = "[x] Done: "
	}
	text := box + task.Text
	for _, note := range notes {
		if note != "" {
			text += " · " + note
		}
	}
	return listItem(text)
}

// listItem is a Markdown list item that keeps the text's line breaks.
func listItem(text string) string { return "- " + hardBreaks(text, "  ") + "\n" }

// quote is a Markdown block quote that keeps the text's line breaks.
func quote(text string) string { return "> " + hardBreaks(text, "> ") + "\n" }

// hardBreaks keeps user text readable inside the document: a line followed by
// another non-blank line ends with a hard break, continuation lines start with
// indent so they stay inside their list item or quote, and lines that would
// otherwise start a heading, quote, fence, rule, or HTML block are escaped.
// List markers are kept so written lists still read as lists.
func hardBreaks(text, indent string) string {
	lines := strings.Split(strings.TrimRight(withoutControls(text), "\n"), "\n")
	for i := range lines {
		line := escapeBlock(strings.TrimRight(lines[i], " \t"))
		if line != "" && i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "" {
			line += "  "
		}
		if i > 0 {
			line = indent + line
			if strings.TrimSpace(lines[i]) == "" {
				line = strings.TrimRight(indent, " ")
			}
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

// escapeBlock backslash-escapes a line whose start Markdown would read as
// document structure rather than text.
func escapeBlock(line string) string {
	content := strings.TrimLeft(line, " \t")
	lead := line[:len(line)-len(content)]
	if content == "" {
		return line
	}
	rule := strings.Trim(content, "-=*_ ") == ""
	if marker := listMarker(content); marker != "" && !rule {
		return lead + marker + escapeBlock(content[len(marker):])
	}
	if rule || strings.ContainsRune("#><|", rune(content[0])) || strings.HasPrefix(content, "```") || strings.HasPrefix(content, "~~~") {
		return lead + "\\" + content
	}
	return line
}

// listMarker returns the bullet or ordered-list marker, with its following
// space, that starts content, or "" when content is not a list item.
func listMarker(content string) string {
	if match := listMarkerPattern.FindString(content); match != "" && len(match) < len(content) {
		return match
	}
	return ""
}

var listMarkerPattern = regexp.MustCompile(`^(?:[-*+]|[0-9]{1,9}[.)])[ \t]+`)

// withoutControls drops terminal control characters so reading the export in a
// terminal cannot run escape sequences.
func withoutControls(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, text)
}
