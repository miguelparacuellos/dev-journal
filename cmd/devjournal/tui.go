package main

import (
	"fmt"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"devjournal/journal"
	"github.com/charmbracelet/x/ansi"
)

type model struct {
	app                     *journal.Journal
	today, day, theme       string
	ascii                   bool
	width, height, selected int
	editor                  textarea.Model
	preview                 viewport.Model
	mode                    string
	editingID               string
	status                  string
	help                    bool
}

func newModel(app *journal.Journal, day, theme string, ascii bool) model {
	editor := textarea.New()
	editor.Placeholder = "What moved forward?"
	editor.ShowLineNumbers = false
	editor.CharLimit = 0
	editor.Prompt = ""
	editor.SetHeight(3)
	editor.Focus()
	return model{app: app, today: day, day: day, theme: theme, ascii: ascii, editor: editor, preview: viewport.New(viewport.WithWidth(70), viewport.WithHeight(10)), mode: "capture", status: "Local journal • ready to write"}
}
func (m model) Init() tea.Cmd { return textarea.Blink }
func (m *model) size() {
	m.editor.SetWidth(max(20, m.width-8))
	m.editor.SetHeight(3)
	m.preview.SetWidth(max(20, m.width-8))
	m.preview.SetHeight(max(3, m.height-12))
}
func (m *model) entries() []journal.Entry { return m.app.Entries(m.day) }
func (m *model) save() {
	var err error
	if m.editingID != "" {
		err = m.app.Correct(m.editingID, m.editor.Value())
	} else {
		_, err = m.app.Capture(m.today, m.editor.Value())
	}
	if err != nil {
		m.status = "Not saved: " + err.Error()
		return
	}
	m.status = "Saved • safely stored locally"
	m.editor.Reset()
	m.editingID = ""
	m.day = m.today
	m.mode = "capture"
	m.selected = max(0, len(m.entries())-1)
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.size()
		return m, nil
	case tea.KeyPressMsg:
		key := msg.String()
		if m.help {
			if key == "esc" || key == "?" {
				m.help = false
			}
			return m, nil
		}
		if m.width < 80 || m.height < 24 {
			if key == "ctrl+c" && m.editor.Value() == "" {
				return m, tea.Quit
			}
			return m, nil
		}
		if m.mode == "capture" || m.mode == "edit" {
			switch key {
			case "ctrl+s":
				m.save()
				return m, nil
			case "esc":
				m.editor.Blur()
				m.mode = "browse"
				m.status = "Draft retained • n resumes capture"
				return m, nil
			case "ctrl+c":
				if m.editor.Value() == "" {
					return m, tea.Quit
				}
				m.status = "Draft retained • Ctrl+S saves; Esc browses"
				return m, nil
			}
			var cmd tea.Cmd
			m.editor, cmd = m.editor.Update(msg)
			return m, cmd
		}
		if m.mode == "detail" {
			if key == "esc" {
				m.mode = "browse"
				return m, nil
			}
			if key == "q" {
				if m.editor.Value() == "" {
					return m, tea.Quit
				}
				m.status = "Unsaved draft • Esc back, then n to resume"
				return m, nil
			}
			var cmd tea.Cmd
			m.preview, cmd = m.preview.Update(msg)
			return m, cmd
		}
		switch key {
		case "q", "ctrl+c":
			if m.editor.Value() != "" {
				m.status = "Unsaved draft • n resumes; x discards before quitting"
				return m, nil
			}
			return m, tea.Quit
		case "?":
			m.help = true
		case "n":
			m.mode = "capture"
			if m.editingID != "" {
				m.mode = "edit"
			}
			return m, m.editor.Focus()
		case "x":
			m.editor.Reset()
			m.editingID = ""
			m.status = "Draft discarded"
		case "t":
			m.day = m.today
			m.selected = 0
		case "left", "h", "right", "l":
			days := m.app.Workdays()
			if len(days) == 0 {
				return m, nil
			}
			index := -1
			for i, d := range days {
				if d == m.day {
					index = i
				}
			}
			if key == "left" || key == "h" {
				index = min(len(days)-1, index+1)
			} else {
				index = max(0, index-1)
			}
			m.day = days[index]
			m.selected = 0
		case "up", "k":
			m.selected = max(0, m.selected-1)
		case "down", "j":
			m.selected = min(max(0, len(m.entries())-1), m.selected+1)
		case "e":
			if m.editor.Value() != "" {
				m.status = "Draft retained • resume or discard it before editing"
				break
			}
			entries := m.entries()
			if len(entries) > 0 {
				m.editingID = entries[m.selected].ID
				m.editor.SetValue(entries[m.selected].Text)
				m.mode = "edit"
				return m, m.editor.Focus()
			}
		case "enter":
			entries := m.entries()
			if len(entries) > 0 {
				m.mode = "detail"
				m.preview.SetContent(ansi.Hardwrap(displayText(entries[m.selected].Text), max(20, m.width-8), true))
				m.preview.GotoTop()
			}
		case "esc":
			m.mode = "browse"
		}
		if m.mode == "detail" {
			var cmd tea.Cmd
			m.preview, cmd = m.preview.Update(msg)
			return m, cmd
		}
	}
	if m.mode == "capture" || m.mode == "edit" {
		var cmd tea.Cmd
		m.editor, cmd = m.editor.Update(msg)
		return m, cmd
	}
	return m, nil
}
func (m model) style(role string) lipgloss.Style {
	s := lipgloss.NewStyle()
	if role == "title" || role == "accent" {
		s = s.Bold(true)
	}
	if m.theme == "mono" {
		return s
	}
	colors := map[string]string{"title": "#E8EAF2", "muted": "#A5AEC4", "accent": "#9CB8FF", "success": "#99D5B4"}
	if m.theme == "light" {
		colors = map[string]string{"title": "#18233A", "muted": "#47536B", "accent": "#234FAD", "success": "#17683B"}
	}
	return s.Foreground(lipgloss.Color(colors[role]))
}
func (m model) View() tea.View {
	if m.width < 80 || m.height < 24 {
		v := tea.NewView("Dev Journal\nResize to at least 80 x 24. Your saved data and draft are preserved.\nCtrl+C exits when no draft is pending.")
		v.AltScreen = true
		return v
	}
	w := m.width - 8
	title := m.style("title").Render("DEV JOURNAL") + "    " + m.style("accent").Render("Today / Daily log")
	rule := strings.Repeat("─", w)
	if m.ascii {
		rule = strings.Repeat("-", w)
	}
	date := m.day
	if m.day == m.today {
		date = "Today  " + m.day
	} else {
		date = "History  " + m.day
	}
	header := title + "\n" + m.style("muted").Render(date) + "\n" + m.style("muted").Render(rule)
	body := ""
	if m.help {
		body = "KEYBOARD GUIDE\n\nCapture / edit: Enter inserts a line; Ctrl+S saves.\nEsc pauses editing and keeps your draft.\n\nBrowse: n capture/resume · e correct · Enter full entry\nUp/Down or j/k select · Left/Right or h/l workdays\nt Today · ? help · q quit · x discard retained draft\n\nFull entry: PgUp/PgDown scroll · Esc back\n\nNo network. No mandatory hours. Saved only after a durable write.\n\nEsc closes help"
	} else if m.mode == "detail" {
		body = "FULL ENTRY · " + m.day + "\n" + m.preview.View()
	} else {
		focus := "[ ] Capture"
		if m.mode == "capture" {
			focus = "[FOCUS] Capture"
		}
		if m.mode == "edit" {
			focus = "[FOCUS] Correct entry · original workday retained"
		}
		body = m.style("accent").Render(focus) + "\n" + m.editor.View() + "\n\n"
		entries := m.entries()
		logHeader := m.style("title").Render(fmt.Sprintf("DAILY LOG  %d entries", len(entries)))
		body += logHeader + "\n"
		if len(entries) == 0 {
			body += "\nA clear page. Record an outcome, progress, or an event.\n"
		} else {
			logWidth := w
			if m.width >= 110 {
				logWidth = w/2 - 3
			}
			available := max(2, m.height-15)
			start := max(0, m.selected-available/2)
			end := min(len(entries), start+available)
			for i := start; i < end; i++ {
				marker := "  "
				if i == m.selected {
					marker = "> "
				}
				text := strings.ReplaceAll(displayText(entries[i].Text), "\n", " / ")
				line := marker + ansi.Truncate(text, logWidth-2, "...")
				if i == m.selected {
					line = m.style("accent").Render(line)
				}
				body += line + "\n"
			}
			body += m.style("muted").Render(fmt.Sprintf("%d-%d of %d · Enter reads the complete entry", start+1, end, len(entries)))
			if m.width >= 110 {
				parts := strings.SplitN(body, logHeader, 2)
				right := m.style("title").Render("SELECTED ENTRY") + "\n" + ansi.Hardwrap(displayText(entries[m.selected].Text), w-logWidth-5, true)
				rightLines := strings.Split(right, "\n")
				if len(rightLines) > available+1 {
					right = strings.Join(rightLines[:available], "\n") + "\n[Enter to read more]"
				}
				body = parts[0] + lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(logWidth+2).Render(logHeader+parts[1]), "  ", lipgloss.NewStyle().Width(w-logWidth-4).Render(right))
			}
		}
	}
	hints := "Ctrl+S Save · Enter New line · Esc Browse"
	if m.mode == "browse" {
		hints = "n Capture · e Correct · Enter Read · arrows Navigate · ? Help · q Quit"
	}
	if m.mode == "detail" {
		hints = "PgUp/PgDown Scroll · Esc Back · q Quit"
	}
	if m.help {
		hints = "Esc Back"
	}
	content := header + "\n\n" + body
	lines := strings.Count(content, "\n") + 1
	content += strings.Repeat("\n", max(1, m.height-4-lines))
	content += m.style("success").Render(ansi.Truncate(m.status, w, "...")) + "\n" + m.style("muted").Render(hints)
	if m.ascii {
		content = strings.NewReplacer("•", "|", "·", "|").Replace(content)
	}
	v := tea.NewView(lipgloss.NewStyle().Padding(1, 4).Render(content))
	v.AltScreen = true
	return v
}

func displayText(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, text)
}
