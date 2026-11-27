package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miguelparacuellos/dev-journal/internal/claude"
	"github.com/miguelparacuellos/dev-journal/internal/store"
)

type oneOnOneScreen struct {
	app   *App
	dates []time.Time
	idx   int
	ed    docEditor
}

func newOneOnOneScreen(a *App) *oneOnOneScreen {
	return &oneOnOneScreen{app: a, ed: newDocEditor(store.SecActions)}
}

func (s *oneOnOneScreen) empty() bool { return len(s.dates) == 0 }

func (s *oneOnOneScreen) load() error {
	var selected time.Time
	if !s.empty() {
		selected = s.dates[s.idx]
	}
	s.dates = s.app.st.OneOnOnes()
	s.idx = s.defaultIdx()
	for i, d := range s.dates {
		if d.Equal(selected) {
			s.idx = i
		}
	}
	return s.loadDoc()
}

// defaultIdx selects the next upcoming 1:1, else the latest one.
func (s *oneOnOneScreen) defaultIdx() int {
	idx := 0
	for i, d := range s.dates {
		if !d.Before(s.app.today()) {
			idx = i
		}
	}
	return idx
}

func (s *oneOnOneScreen) loadDoc() error {
	if s.empty() {
		s.ed.setDoc(store.Doc{})
		return nil
	}
	doc, err := s.app.st.LoadOneOnOne(s.dates[s.idx])
	s.ed.setDoc(doc)
	return err
}

func (s *oneOnOneScreen) save() tea.Cmd {
	if s.empty() {
		return nil
	}
	if err := store.SaveDoc(s.app.st.OneOnOnePath(s.dates[s.idx]), s.ed.doc); err != nil {
		return errStatus(err)
	}
	return nil
}

func (s *oneOnOneScreen) file() (string, error) {
	if s.empty() {
		return "", fmt.Errorf("no 1:1 yet — press n")
	}
	s.save()
	return s.app.st.OneOnOnePath(s.dates[s.idx]), nil
}

func (s *oneOnOneScreen) capturing() bool { return s.ed.editing() }

func (s *oneOnOneScreen) handleKey(msg tea.KeyMsg) tea.Cmd {
	if !s.ed.editing() {
		switch msg.String() {
		case "n":
			return s.app.openPrompt("new 1:1 date (YYYY-MM-DD) ›", s.app.today().AddDate(0, 0, 7).Format(store.DateLayout), s.create)
		case "left", "h", "[":
			if s.idx < len(s.dates)-1 {
				s.idx++
				return s.wrap(s.loadDoc())
			}
			return nil
		case "right", "l", "]":
			if s.idx > 0 {
				s.idx--
				return s.wrap(s.loadDoc())
			}
			return nil
		case "P":
			if !s.empty() {
				return s.draftAgenda()
			}
			return nil
		}
	}
	if s.empty() {
		return nil
	}
	changed, cmd := s.ed.update(msg)
	if changed {
		return tea.Batch(cmd, s.save())
	}
	return cmd
}

func (s *oneOnOneScreen) wrap(err error) tea.Cmd {
	if err != nil {
		return errStatus(err)
	}
	return nil
}

func (s *oneOnOneScreen) create(input string) tea.Cmd {
	d, err := time.ParseInLocation(store.DateLayout, input, time.Local)
	if err != nil {
		return errStatus(fmt.Errorf("invalid date %q", input))
	}
	doc, err := s.app.st.LoadOneOnOne(d)
	if err != nil {
		return errStatus(err)
	}
	if err := store.SaveDoc(s.app.st.OneOnOnePath(d), doc); err != nil {
		return errStatus(err)
	}
	s.dates = s.app.st.OneOnOnes()
	for i, x := range s.dates {
		if x.Equal(d) {
			s.idx = i
		}
	}
	return tea.Batch(s.wrap(s.loadDoc()), status("1:1 created for "+d.Format("Mon 02 Jan")))
}

func (s *oneOnOneScreen) previous() time.Time {
	if s.idx+1 < len(s.dates) {
		return s.dates[s.idx+1]
	}
	return s.dates[s.idx].AddDate(0, -1, 0)
}

// agendaContext gathers what Claude needs without giving it file access.
func (s *oneOnOneScreen) agendaContext() string {
	st := s.app.st
	from, to := s.previous(), s.dates[s.idx]
	var b strings.Builder
	b.WriteString("## Dailies\n")
	for _, d := range st.Dailies(s.app.today()) {
		if d.Before(from) || d.After(to) {
			continue
		}
		if raw, err := os.ReadFile(st.DailyPath(d)); err == nil {
			b.WriteString(string(raw) + "\n")
		}
	}
	b.WriteString("\n## Wins\n")
	for _, w := range st.WinsSince(from) {
		b.WriteString("- " + w + "\n")
	}
	b.WriteString("\n## Open actions from previous 1:1s\n")
	for _, a := range st.OpenActionsBefore(to) {
		b.WriteString("- " + a.Text + "\n")
	}
	b.WriteString("\n## Topics already on the agenda\n")
	for _, it := range s.ed.doc.Section(store.SecTopics).Items {
		b.WriteString("- " + it.Text + "\n")
	}
	return b.String()
}

func (s *oneOnOneScreen) draftAgenda() tea.Cmd {
	notes := s.agendaContext()
	root := s.app.st.Root
	run := func() tea.Cmd {
		return func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			a, err := claude.DraftAgenda(ctx, root, notes)
			if err != nil {
				return pickResultMsg{err: err}
			}
			return pickResultMsg{items: toPickItems(store.SecTopics, a.Topics)}
		}
	}
	apply := func(chosen []pickItem) tea.Cmd {
		for _, it := range chosen {
			s.ed.Append(store.SecTopics, it.text)
		}
		return tea.Batch(s.save(), status(fmt.Sprintf("added %d topics", len(chosen))))
	}
	cmdline := fmt.Sprintf(`claude -p "1:1 agenda" · dailies, wins, open actions since %s`, s.previous().Format("02 Jan"))
	return s.app.openPicker("draft 1:1 agenda", cmdline, run, apply)
}

func (s *oneOnOneScreen) info() string {
	if s.empty() {
		return "1:1s · none yet"
	}
	d := s.dates[s.idx]
	days := int(d.Sub(s.app.today()).Hours() / 24)
	switch {
	case days > 0:
		return fmt.Sprintf("1:1 · %s (in %d days)", d.Format("Mon 02 Jan"), days)
	case days == 0:
		return "1:1 · today"
	}
	return "1:1 · " + d.Format("Mon 02 Jan 2006")
}

func (s *oneOnOneScreen) listTitle() string { return "meetings" }

func (s *oneOnOneScreen) listView(w, h int) string {
	var lines []string
	today := s.app.today()
	for i, d := range s.dates {
		if len(lines) >= h {
			break
		}
		label := d.Format("02 Jan 2006")
		if !d.Before(today) {
			label += " " + sAccent.Render("next")
		}
		if i == s.idx {
			lines = append(lines, sSel.Render(padRight("▸ "+label, w)))
		} else {
			lines = append(lines, "  "+label)
		}
	}
	if len(lines) == 0 {
		lines = append(lines, sDim.Render("n to create"))
	}
	return strings.Join(lines, "\n")
}

func (s *oneOnOneScreen) mainTitle() string {
	if s.empty() {
		return "1:1"
	}
	return "1:1 · " + strings.ToLower(s.dates[s.idx].Format("Mon 02 Jan"))
}

func (s *oneOnOneScreen) mainView(w, h int) string {
	if s.empty() {
		return sDim.Render("No 1:1s yet. Press n to create the next one.")
	}
	return s.ed.view(w, h)
}

func (s *oneOnOneScreen) side() []sideBox {
	if s.empty() {
		return []sideBox{{"carried actions", ""}, {"wins", ""}}
	}
	var actions []string
	for _, a := range s.app.st.OpenActionsBefore(s.dates[s.idx]) {
		actions = append(actions, sAccent.Render("[ ] ")+truncate(a.Text, rightW-10), "    "+sDim.Render("from "+a.From.Format("02 Jan")))
	}
	if len(actions) == 0 {
		actions = append(actions, sDim.Render("None open."))
	}
	var wins []string
	for _, w := range s.app.st.WinsSince(s.previous()) {
		wins = append(wins, sGreen.Render("+ ")+truncate(w, rightW-6))
	}
	if len(wins) == 0 {
		wins = append(wins, sDim.Render("None logged — press 4 to add."))
	}
	return []sideBox{
		{"carried actions · open", strings.Join(actions, "\n")},
		{"wins since " + s.previous().Format("02 Jan"), strings.Join(wins, "\n")},
	}
}

func (s *oneOnOneScreen) hints() []hint {
	if s.ed.editing() {
		return []hint{{"enter", "save"}, {"esc", "cancel"}}
	}
	return []hint{{"n", "new 1:1"}, {"P", "✦ draft agenda"}, {"a", "add"}, {"tab", "next section"}, {"x", "toggle"}, {"←→", "meeting"}}
}
