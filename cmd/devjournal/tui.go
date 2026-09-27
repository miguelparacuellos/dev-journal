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
	section                 string
	planFocus               bool
	showCompleted           bool
	captureKind             string
	dailyFocus              int
	shared                  map[string]bool
	detailText              string
}

func newModel(app *journal.Journal, day, theme string, ascii bool) model {
	editor := textarea.New()
	editor.Placeholder = "What moved forward?"
	editor.ShowLineNumbers = false
	editor.CharLimit = 0
	editor.MaxHeight = 0
	editor.MaxWidth = 0
	editor.Prompt = ""
	state := textarea.StyleState{Placeholder: lipgloss.NewStyle().Faint(true), Selection: lipgloss.NewStyle().Reverse(true)}
	editor.SetStyles(textarea.Styles{Focused: state, Blurred: state, Cursor: textarea.CursorStyle{Shape: tea.CursorBlock, Blink: true}})
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
	if m.mode == "detail" {
		entries := m.entries()
		if len(entries) > 0 {
			m.preview.SetContent(ansi.Hardwrap(displayText(entries[m.selected].Text), max(20, m.width-8), true))
		}
	}
	if m.mode == "daily-detail" {
		m.preview.SetContent(ansi.Hardwrap(displayText(m.detailText), max(20, m.width-8), true))
	}
	if m.mode == "task-detail" {
		tasks := m.tasks()
		if m.planFocus {
			tasks = m.app.Plan(m.today)
		}
		if len(tasks) > 0 {
			m.preview.SetContent(ansi.Hardwrap(displayText(tasks[m.selected].Text), max(20, m.width-8), true))
		}
	}
}
func (m *model) entries() []journal.Entry { return m.app.Entries(m.day) }
func (m *model) save() {
	var err error
	if m.captureKind == "blocker" {
		err = m.app.RecordBlocker(m.today, m.editor.Value())
	} else if m.captureKind == "proposal" {
		err = m.app.SaveDaily(m.today, m.editor.Value())
	} else if m.section == "tasks" {
		_, err = m.app.CreateTask(m.editor.Value())
	} else if m.editingID != "" {
		err = m.app.Correct(m.editingID, m.editor.Value())
	} else {
		_, err = m.app.Capture(m.today, m.editor.Value())
	}
	if err != nil {
		m.status = "Not saved: " + err.Error()
		return
	}
	m.status = "Saved • safely stored locally"
	kind := m.captureKind
	m.resetCaptureKind()
	m.editor.Reset()
	m.editingID = ""
	m.day = m.today
	m.mode = "capture"
	if kind != "" {
		m.mode = "browse"
		m.editor.Blur()
	}
	if m.section == "tasks" {
		m.showCompleted = false
		m.selected = max(0, len(m.tasks())-1)
	} else {
		m.selected = max(0, len(m.entries())-1)
		if m.section == "daily" {
			m.selected = 0
		}
	}
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
				if m.editor.Value() == "" {
					m.resetCaptureKind()
				}
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
		if m.isDetailMode() {
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

		if key == "g" || key == "b" {
			if m.editor.Value() != "" {
				m.status = "Draft retained • n resumes; x discards before changing views"
				return m, nil
			}
			if key == "g" {
				m.section = "daily"
				m.planFocus = false
				m.selected = 0
				m.dailyFocus = 0
				return m, nil
			}
			m.planFocus = false
			m.selected = 0
			m.captureKind = "blocker"
			m.editor.Placeholder = "What needs help?"
			m.mode = "capture"
			return m, m.editor.Focus()
		}
		if m.section == "daily" && key != "tab" && key != "t" && key != "q" && key != "ctrl+c" && key != "?" && key != "x" {
			return m.updateDaily(key)
		}
		if key == "tab" || key == "t" {
			if m.editor.Value() != "" {
				m.status = "Draft retained • n resumes; x discards before changing views"
				return m, nil
			}
			if key == "tab" && m.section != "tasks" {
				m.section = "tasks"
				m.editor.Placeholder = "What needs doing?"
			} else {
				m.section = ""
				m.editor.Placeholder = "What moved forward?"
				m.day = m.today
			}
			m.selected = 0
			m.planFocus = false
			return m, nil
		}
		if m.section == "tasks" || m.planFocus {
			tasks := m.tasks()
			if m.planFocus {
				tasks = m.app.Plan(m.today)
			}
			switch key {
			case "up", "k":
				m.selected = max(0, m.selected-1)
				return m, nil
			case "down", "j":
				m.selected = min(max(0, len(tasks)-1), m.selected+1)
				return m, nil
			case "c":
				if m.section == "tasks" {
					m.showCompleted = !m.showCompleted
					m.selected = 0
				}
				return m, nil
			case "p":
				if m.planFocus {
					m.planFocus = false
					m.selected = 0
					return m, nil
				}
				if len(tasks) > 0 {
					err := m.app.PlanTask(m.today, tasks[m.selected].ID)
					m.actionStatus(err, "Selected for today • still open")
				}
				return m, nil
			case "u":
				if m.planFocus && len(tasks) > 0 {
					err := m.app.UnplanTask(m.today, tasks[m.selected].ID)
					m.actionStatus(err, "Removed from today’s plan • task retained")
					m.selected = max(0, min(m.selected, len(m.app.Plan(m.today))-1))
				}
				return m, nil
			case "d":
				if len(tasks) > 0 {
					err := m.app.CompleteTask(tasks[m.selected].ID)
					m.actionStatus(err, "Completed • safely stored locally")
					if !m.planFocus {
						m.selected = max(0, min(m.selected, len(m.tasks())-1))
					}
				}
				return m, nil
			case "enter":
				if len(tasks) > 0 {
					m.mode = "task-detail"
					m.preview.SetContent(ansi.Hardwrap(displayText(tasks[m.selected].Text), max(20, m.width-8), true))
					m.preview.GotoTop()
				}
				return m, nil
			case "e", "left", "right", "h", "l":
				return m, nil
			}
		} else if key == "p" {
			m.day = m.today
			m.planFocus = true
			m.selected = 0
			return m, nil
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
			if m.planFocus {
				m.planFocus = false
				m.selected = 0
			}
			m.mode = "capture"
			if m.editingID != "" {
				m.mode = "edit"
			}
			return m, m.editor.Focus()
		case "x":
			m.editor.Reset()
			m.resetCaptureKind()
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
		if m.isDetailMode() {
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
	viewName := "Today / Daily log"
	if m.section == "daily" {
		viewName = "Daily"
	}
	if m.section == "tasks" {
		viewName = "Tasks"
	}
	title := m.style("title").Render("DEV JOURNAL") + "    " + m.style("accent").Render(viewName)
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
	if m.section == "tasks" {
		date = "Undated actions • deliberately select today’s priorities"
	}
	header := title + "\n" + m.style("muted").Render(date) + "\n" + m.style("muted").Render(rule)
	body := ""
	if m.help {
		body = "KEYBOARD GUIDE\n\nCapture: Enter new line · Ctrl+S save · Esc retain draft\nBrowse: n capture/resume · x discard draft · q quit\nTab Today/Tasks · g Daily · b Blocker · t Today · ? help\n\nTasks: arrows select · p plan · d complete · c open/done\nToday: p focus plan · d complete · u remove · p log\nLog: arrows/j/k select · h/l workdays · e correct\nDaily: arrows sections/rows · s share · e edit · r prepare · v saved\nEnter reads full text · PgUp/PgDown scroll · Esc back\n\nSaved locally after a durable write. Esc closes help."

	} else if m.isDetailMode() {
		body = "FULL TEXT · " + viewName + "\n" + m.preview.View()
	} else if m.section == "daily" && m.mode == "browse" {
		body = m.dailyBody(w)
	} else {
		focus := "[ ] Capture"
		if m.section == "tasks" {
			focus = "[ ] New task"
		}
		if m.mode == "capture" {
			focus = "[FOCUS] Capture"
			if m.section == "tasks" {
				focus = "[FOCUS] New task • no workday assigned"
			}
		}
		if m.mode == "edit" {
			focus = "[FOCUS] Correct entry · original workday retained"
		}
		if m.captureKind == "blocker" {
			focus = "[FOCUS] Record blocker • " + m.today
		}
		if m.captureKind == "proposal" {
			focus = "[FOCUS] Personal preparation • source data unchanged"
		}
		body = m.style("accent").Render(focus) + "\n" + m.editor.View() + "\n\n"
		if m.section == "daily" {
			body += "Ctrl+S saves personal preparation; Esc retains the draft.\n"
		} else if m.section == "tasks" {
			tasks := m.tasks()
			label := "OPEN TASKS"
			if m.showCompleted {
				label = "COMPLETED TASKS"
			}
			body += m.style("title").Render(fmt.Sprintf("%s  %d", label, len(tasks))) + "\n"
			body += m.taskRows(tasks, max(1, m.height-17), w, true)
		} else {
			plan := m.app.Plan(m.today)
			label := "TODAY'S PLAN • " + m.today + " • intended actions"
			if m.planFocus {
				label = "[FOCUS] " + label
			}
			body += m.style("title").Render(label) + "\n"
			planRows := 3
			if m.height < 30 {
				planRows = 2
			}
			body += m.taskRows(plan, planRows, w, m.planFocus) + "\n"
			entries := m.entries()
			logHeader := m.style("title").Render(fmt.Sprintf("DAILY LOG  %d entries", len(entries)))
			body += m.style("title").Render("BLOCKERS • b records • g Daily reads") + "\n"
			blockers := m.app.Blockers(m.today)
			text := "None recorded."
			if len(blockers) > 0 {
				text = fmt.Sprintf("%d recorded: %s", len(blockers), strings.ReplaceAll(displayText(blockers[len(blockers)-1].Text), "\n", " / "))
			}
			body += ansi.Truncate(text, w, "...") + "\n"
			body += logHeader + "\n"
			if len(entries) == 0 {
				body += "A clear page. Record an outcome, progress, or an event.\n"
			} else {
				logWidth := w
				if m.width >= 110 {
					logWidth = w/2 - 3
				}
				available := max(1, m.height-25)
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
					right := m.style("title").Render("SELECTED ENTRY") + "\n" + ansi.Hardwrap(displayText(entries[min(m.selected, len(entries)-1)].Text), w-logWidth-5, true)
					rightLines := strings.Split(right, "\n")
					if len(rightLines) > available+1 {
						right = strings.Join(rightLines[:available], "\n") + "\n[Enter to read more]"
					}
					body = parts[0] + lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(logWidth+2).Render(logHeader+parts[1]), "  ", lipgloss.NewStyle().Width(w-logWidth-4).Render(right))
				}
			}
		}
	}
	hints := "Ctrl+S Save · Enter New line · Esc Browse"
	if m.mode == "browse" {
		hints = "n Capture · e Edit · p Plan · b Blocker · g Daily · Tab Tasks · q Quit"
		if m.section == "tasks" {
			hints = "n New · p Plan · d Done · c Open/Done · Tab Today · ? Help · q Quit"
		}
		if m.section == "daily" {
			hints = "s Share · e Edit · r Prepare · v Saved · b Blocker · t Today · q Quit"
		}
		if m.planFocus {
			hints = "arrows Select · d Complete · u Remove · p Log · Tab Tasks"
		}
	}
	if m.isDetailMode() {
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

func (m model) tasks() []journal.Task {
	result := []journal.Task{}
	for _, task := range m.app.Tasks() {
		if task.Completed == m.showCompleted {
			result = append(result, task)
		}
	}
	return result
}
func (m *model) actionStatus(err error, success string) {
	if err != nil {
		m.status = "Not saved: " + err.Error()
	} else {
		m.status = success
	}
}
func (m model) taskRows(tasks []journal.Task, limit, width int, focused bool) string {
	if len(tasks) == 0 {
		if m.section == "tasks" {
			return "\nNo tasks here. n captures an action; c switches open/completed.\n"
		}
		return "No actions selected. Tab opens Tasks; p selects an open task.\n"
	}
	start := 0
	if focused {
		start = max(0, m.selected-limit/2)
	}
	end := min(len(tasks), start+limit)
	result := ""
	for i := start; i < end; i++ {
		marker := "  "
		if focused && i == m.selected {
			marker = "> "
		}
		state := "[open] "
		if tasks[i].Completed {
			state = "[done] "
		}
		line := marker + state + ansi.Truncate(strings.ReplaceAll(displayText(tasks[i].Text), "\n", " / "), width-11, "...")
		if focused && i == m.selected {
			line = m.style("accent").Render(line)
		}
		result += line + "\n"
	}
	if len(tasks) > limit {
		result += m.style("muted").Render(fmt.Sprintf("%d-%d of %d • arrows select", start+1, end, len(tasks))) + "\n"
	}
	return result
}

func (m model) isDetailMode() bool {
	return m.mode == "detail" || m.mode == "task-detail" || m.mode == "daily-detail"
}
func (m *model) resetCaptureKind() {
	m.captureKind = ""
	m.editor.Placeholder = "What moved forward?"
	if m.section == "tasks" {
		m.editor.Placeholder = "What needs doing?"
	}
}
