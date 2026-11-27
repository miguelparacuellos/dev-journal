package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miguelparacuellos/dev-journal/internal/claude"
	"github.com/miguelparacuellos/dev-journal/internal/store"
)

type dailyScreen struct {
	app    *App
	dates  []time.Time
	idx    int
	ed     docEditor
	dirty  bool
	linear store.LinearCache
}

func newDailyScreen(a *App) *dailyScreen {
	s := &dailyScreen{app: a, ed: newDocEditor(store.SecToday)}
	s.ed.notes = map[string]string{}
	return s
}

func (s *dailyScreen) date() time.Time { return s.dates[s.idx] }

func (s *dailyScreen) load() error {
	var selected time.Time
	if len(s.dates) > 0 {
		selected = s.date()
	}
	s.dates = s.app.st.Dailies(s.app.today())
	s.idx = 0
	for i, d := range s.dates {
		if d.Equal(selected) {
			s.idx = i
		}
	}
	s.linear = s.app.st.LoadLinearCache()
	return s.loadDoc()
}

func (s *dailyScreen) loadDoc() error {
	doc, err := s.app.st.LoadDaily(s.date())
	s.ed.setDoc(doc)
	s.dirty = false
	return err
}

func (s *dailyScreen) save() tea.Cmd {
	if err := store.SaveDoc(s.app.st.DailyPath(s.date()), s.ed.doc); err != nil {
		return errStatus(err)
	}
	s.dirty = false
	return nil
}

func (s *dailyScreen) file() (string, error) {
	if cmd := s.save(); cmd != nil {
		return "", fmt.Errorf("save failed")
	}
	return s.app.st.DailyPath(s.date()), nil
}

func (s *dailyScreen) capturing() bool { return s.ed.editing() }

func (s *dailyScreen) handleKey(msg tea.KeyMsg) tea.Cmd {
	if !s.ed.editing() {
		switch msg.String() {
		case "left", "h", "[":
			if s.idx < len(s.dates)-1 {
				s.idx++
				return s.wrapErr(s.loadDoc())
			}
			return nil
		case "right", "l", "]":
			if s.idx > 0 {
				s.idx--
				return s.wrapErr(s.loadDoc())
			}
			return nil
		case "t":
			s.idx = 0
			return s.wrapErr(s.loadDoc())
		case "g":
			return s.draft()
		case "y":
			return copyToClipboard(slackDaily(s.ed.doc))
		}
	}
	changed, cmd := s.ed.update(msg)
	if changed {
		return tea.Batch(cmd, s.save())
	}
	return cmd
}

func (s *dailyScreen) wrapErr(err error) tea.Cmd {
	if err != nil {
		return errStatus(err)
	}
	return nil
}

// previousWorkday is what "yesterday" means for a standup: Friday on Mondays.
func previousWorkday(d time.Time) time.Time {
	p := d.AddDate(0, 0, -1)
	for p.Weekday() == time.Saturday || p.Weekday() == time.Sunday {
		p = p.AddDate(0, 0, -1)
	}
	return p
}

func (s *dailyScreen) draft() tea.Cmd {
	date := s.date()
	since := previousWorkday(date)
	existing := s.ed.doc.Render()
	root := s.app.st.Root
	st := s.app.st
	run := func() tea.Cmd {
		return func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			d, err := claude.DraftDaily(ctx, root, date, since, existing)
			if err != nil {
				return pickResultMsg{err: err}
			}
			if len(d.Issues) > 0 {
				c := store.LinearCache{SyncedAt: time.Now()}
				for _, i := range d.Issues {
					c.Issues = append(c.Issues, store.LinearIssue(i))
				}
				_ = st.SaveLinearCache(c)
			}
			items := toPickItems(store.SecYesterday, d.Yesterday)
			items = append(items, toPickItems(store.SecToday, d.Today)...)
			items = append(items, toPickItems(store.SecBlockers, d.Blockers)...)
			return pickResultMsg{items: items}
		}
	}
	apply := func(chosen []pickItem) tea.Cmd {
		for _, it := range chosen {
			s.ed.Append(it.section, it.text)
		}
		s.ed.notes[store.SecYesterday] = "✦ drafted from Linear"
		s.linear = st.LoadLinearCache()
		return tea.Batch(s.save(), status(fmt.Sprintf("inserted %d items", len(chosen))))
	}
	cmdline := fmt.Sprintf(`claude -p "daily draft" · Linear %s current cycle · since %s`, claude.LinearTeam(), since.Format("Mon 02"))
	return s.app.openPicker("draft from Linear", cmdline, run, apply)
}

func slackDaily(d store.Doc) string {
	var b strings.Builder
	for _, sec := range []string{store.SecYesterday, store.SecToday, store.SecBlockers} {
		b.WriteString("*" + sec + "*\n")
		items := d.Section(sec).Items
		if len(items) == 0 {
			b.WriteString("• —\n")
		}
		for _, it := range items {
			b.WriteString("• " + it.Text + "\n")
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func (s *dailyScreen) info() string {
	return s.app.today().Format("Mon 02 Jan") + " · Linear " + claude.LinearTeam()
}

func (s *dailyScreen) listTitle() string { return "days" }

func (s *dailyScreen) listView(w, h int) string {
	var lines []string
	start := max(0, s.idx-h+1)
	lastWeek := -1
	for i := start; i < len(s.dates) && len(lines) < h; i++ {
		d := s.dates[i]
		_, wk := d.ISOWeek()
		if lastWeek != -1 && wk != lastWeek && len(lines) < h-1 {
			lines = append(lines, sDim.Render(fmt.Sprintf("── week %d ──", wk)))
		}
		lastWeek = wk
		label := d.Format("Mon 02 Jan")
		if d.Equal(s.app.today()) {
			label += " " + sAccent.Render("today")
		}
		if i == s.idx {
			lines = append(lines, sSel.Render(padRight("▸ "+label, w)))
		} else {
			lines = append(lines, "  "+label)
		}
	}
	return strings.Join(lines, "\n")
}

func (s *dailyScreen) mainTitle() string {
	return "daily · " + strings.ToLower(s.date().Format("Mon 02 Jan"))
}

func (s *dailyScreen) mainView(w, h int) string { return s.ed.view(w, h) }

func (s *dailyScreen) side() []sideBox {
	var lines []string
	if len(s.linear.Issues) == 0 {
		lines = append(lines, sDim.Render("No data yet."), sDim.Render("Press g to draft from Linear."))
	}
	for _, i := range s.linear.Issues {
		lines = append(lines, statusGlyph(i.Status)+" "+sCyan.Render(i.ID)+" "+sDim.Render(i.Status))
		lines = append(lines, "  "+truncate(i.Title, rightW-8))
	}
	title := "linear · my cycle issues"
	if !s.linear.SyncedAt.IsZero() {
		title += " · " + s.linear.SyncedAt.Format("15:04")
	}
	return []sideBox{{title, strings.Join(lines, "\n")}, {"next 1:1", nextOneOnOneSummary(s.app)}}
}

func statusGlyph(status string) string {
	switch st := strings.ToLower(status); {
	case strings.Contains(st, "done"), strings.Contains(st, "complete"):
		return sGreen.Render("●")
	case strings.Contains(st, "review"):
		return sAccent.Render("◐")
	case strings.Contains(st, "progress"):
		return sAccent.Render("◑")
	case strings.Contains(st, "block"):
		return sRed.Render("!")
	}
	return sDim.Render("○")
}

func nextOneOnOneSummary(a *App) string {
	today := a.today()
	var next time.Time
	for _, d := range a.st.OneOnOnes() {
		if !d.Before(today) {
			next = d
		}
	}
	if next.IsZero() {
		return sDim.Render("None scheduled.\nPress 2, then n.")
	}
	doc, _ := a.st.LoadOneOnOne(next)
	days := int(next.Sub(today).Hours() / 24)
	return fmt.Sprintf("%s · in %d days\n%d topics prepped\n%d open actions",
		next.Format("Mon 02 Jan"), days, len(doc.Section(store.SecTopics).Items), len(a.st.OpenActionsBefore(next)))
}

func (s *dailyScreen) hints() []hint {
	if s.ed.editing() {
		return []hint{{"enter", "save"}, {"esc", "cancel"}}
	}
	return []hint{{"g", "✦ draft from Linear"}, {"a", "add"}, {"x", "toggle"}, {"d", "delete"}, {"y", "copy for Slack"}, {"←→", "day"}}
}
