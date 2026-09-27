package main

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const topicPlaceholder = "What do you want to raise with your tech lead?"

// updateTopics handles the O2O review keys. It reports false only for keys
// shared with other views (capture, help, discard, quit, and Esc); every other
// key is consumed so Today and Tasks actions never act on O2O topics.
func (m model) updateTopics(key string) (model, bool) {
	topics := m.app.OpenTopics()
	switch key {
	case "up", "k":
		m.selected = max(0, m.selected-1)
	case "down", "j":
		m.selected = min(max(0, len(topics)-1), m.selected+1)
	case "enter":
		if len(topics) > 0 {
			m.openText("topic-detail", topics[m.selected].Text)
		}
	case "n", "x", "?", "q", "ctrl+c", "esc":
		return m, false
	}
	return m, true
}

// topicsBody lists open topics for meeting preparation below the composer.
func (m model) topicsBody(width int) string {
	topics := m.app.OpenTopics()
	header := m.style("title").Render(fmt.Sprintf("OPEN TOPICS  %d", len(topics))) + m.style("muted").Render(" • for the next one to one")
	if len(topics) == 0 {
		return header + "\n\nNothing to raise yet. n captures a topic whenever it comes up;\nit stays open across days until it is addressed in a one to one.\n"
	}
	listWidth := width
	if m.width >= wideLayoutWidth {
		listWidth = width/2 - 3
	}
	available := max(1, m.height-18)
	start := max(0, m.selected-available/2)
	end := min(len(topics), start+available)
	list := header + "\n"
	for i := start; i < end; i++ {
		marker := "  "
		if i == m.selected {
			marker = "> "
		}
		text := strings.ReplaceAll(displayText(topics[i].Text), "\n", " / ")
		line := ansi.Truncate(marker+topics[i].Day+"  "+text, listWidth, "...")
		if i == m.selected {
			line = m.style("accent").Render(line)
		}
		list += line + "\n"
	}
	list += m.style("muted").Render(fmt.Sprintf("%d-%d of %d · Enter reads the complete topic", start+1, end, len(topics)))
	if m.width < wideLayoutWidth {
		return list
	}
	selected := topics[min(m.selected, len(topics)-1)]
	previewWidth := width - listWidth - 5
	right := m.style("title").Render("SELECTED TOPIC") + m.style("muted").Render(" • captured "+selected.Day) + "\n" + ansi.Wrap(displayText(selected.Text), previewWidth, "")
	if lines := strings.Split(right, "\n"); len(lines) > available+1 {
		right = strings.Join(lines[:available], "\n") + "\n[Enter to read more]"
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(listWidth+2).Render(list), "  ", lipgloss.NewStyle().Width(previewWidth+1).Render(right))
}
