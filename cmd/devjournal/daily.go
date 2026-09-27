package main

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"strings"
)

func (m model) dailyTexts() []string {
	_, entries := m.app.RecentWork(m.today)
	texts := []string{}
	switch m.dailyFocus {
	case 0:
		for _, entry := range entries {
			texts = append(texts, entry.Text)
		}
	case 1:
		for _, task := range m.app.Plan(m.today) {
			texts = append(texts, task.Text)
		}
	case 2:
		for _, blocker := range m.app.Blockers(m.today) {
			texts = append(texts, blocker.Text)
		}
	}
	return texts
}
func (m model) updateDaily(key string) (tea.Model, tea.Cmd) {
	texts := m.dailyTexts()
	switch key {
	case "left", "h":
		m.dailyFocus = max(0, m.dailyFocus-1)
		m.selected = 0
	case "right", "l":
		m.dailyFocus = min(2, m.dailyFocus+1)
		m.selected = 0
	case "up", "k":
		m.selected = max(0, m.selected-1)
	case "down", "j":
		m.selected = min(max(0, len(texts)-1), m.selected+1)
	case "s", "space":
		_, entries := m.app.RecentWork(m.today)
		if m.dailyFocus == 0 && len(entries) > 0 {
			if m.shared == nil {
				m.shared = map[string]bool{}
			}
			id := entries[m.selected].ID
			m.shared[id] = !m.shared[id]
			m.status = "Selection ready • r prepares; e edits saved preparation"
		}
	case "e", "n", "r":
		if m.editor.Value() != "" {
			m.mode = "capture"
			return m, m.editor.Focus()
		}
		text, ok := m.app.Daily(m.today)
		if !ok || key == "r" {
			ids := []string{}
			for id, selected := range m.shared {
				if selected {
					ids = append(ids, id)
				}
			}
			var err error
			text, err = m.app.PrepareDaily(m.today, ids)
			if err != nil {
				m.actionStatus(err, "")
				return m, nil
			}
		}
		m.captureKind = "proposal"
		m.editor.SetValue(text)
		m.mode = "capture"
		return m, m.editor.Focus()
	case "v":
		text, ok := m.app.Daily(m.today)
		if !ok {
			m.status = "No saved personal preparation • s selects progress; e prepares"
			return m, nil
		}
		m.openDailyText(text)
	case "enter":
		if len(texts) > 0 {
			m.openDailyText(texts[m.selected])
		}
	}
	return m, nil
}
func (m *model) openDailyText(text string) {
	m.detailText = text
	m.mode = "daily-detail"
	m.preview.SetContent(ansi.Hardwrap(displayText(text), max(20, m.width-8), true))
	m.preview.GotoTop()
}
func (m model) dailyBody(width int) string {
	source, entries := m.app.RecentWork(m.today)
	heading := "RECENT WORK"
	if source != "" {
		heading += " • " + source
	} else {
		heading += " • No previous recorded workday"
	}
	labels := []string{heading, "TODAY'S PLAN • " + m.today + " • intended actions", "BLOCKERS • " + m.today}
	body := ""
	limit := max(1, (m.height-17)/3)
	for section, label := range labels {
		if section == m.dailyFocus {
			label = "[FOCUS] " + label
		}
		body += m.style("title").Render(label) + "\n"
		copy := m
		copy.dailyFocus = section
		texts := copy.dailyTexts()
		if len(texts) == 0 {
			body += "None recorded.\n"
			continue
		}
		start := 0
		if section == m.dailyFocus {
			start = max(0, m.selected-limit/2)
		}
		end := min(len(texts), start+limit)
		for i := start; i < end; i++ {
			marker := "  "
			if section == m.dailyFocus && i == m.selected {
				marker = "> "
			}
			if section == 0 {
				if m.shared[entries[i].ID] {
					marker += "[share] "
				} else {
					marker += "[ ] "
				}
			}
			body += ansi.Truncate(marker+strings.ReplaceAll(displayText(texts[i]), "\n", " / "), width, "...") + "\n"
		}
		if len(texts) > limit {
			body += fmt.Sprintf("%d-%d of %d • arrows select\n", start+1, end, len(texts))
		}
	}
	if _, ok := m.app.Daily(m.today); ok {
		body += "\nPERSONAL PREPARATION • saved • v reads; e edits; r starts from selection\n"
	} else {
		body += "\nPERSONAL PREPARATION • not saved • e prepares\n"
	}
	body += "Left/Right selects section; Up/Down selects row; Enter reads."
	return body
}
