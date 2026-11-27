package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miguelparacuellos/dev-journal/internal/store"
)

type winsScreen struct {
	app *App
	ed  docEditor
}

func newWinsScreen(a *App) *winsScreen {
	s := &winsScreen{app: a, ed: newDocEditor()}
	s.ed.onAdd = func(_, text string) string {
		if _, err := time.Parse(store.DateLayout, strings.SplitN(text, " ", 2)[0]); err == nil {
			return text
		}
		return a.today().Format(store.DateLayout) + " · " + text
	}
	return s
}

func (s *winsScreen) load() error {
	doc, err := s.app.st.LoadWins()
	s.ed.setDoc(doc)
	return err
}

func (s *winsScreen) save() tea.Cmd {
	if err := store.SaveDoc(s.app.st.WinsPath(), s.ed.doc); err != nil {
		return errStatus(err)
	}
	return nil
}

func (s *winsScreen) file() (string, error) {
	s.save()
	return s.app.st.WinsPath(), nil
}

func (s *winsScreen) capturing() bool { return s.ed.editing() }

func (s *winsScreen) handleKey(msg tea.KeyMsg) tea.Cmd {
	changed, cmd := s.ed.update(msg)
	if changed {
		return tea.Batch(cmd, s.save())
	}
	return cmd
}

func (s *winsScreen) info() string {
	return fmt.Sprintf("Wins · %d logged", len(s.ed.doc.Section(store.SecWins).Items))
}

func (s *winsScreen) listTitle() string { return "by month" }

func (s *winsScreen) listView(w, h int) string {
	counts := map[string]int{}
	var order []string
	for _, it := range s.ed.doc.Section(store.SecWins).Items {
		if len(it.Text) < 7 {
			continue
		}
		t, err := time.Parse("2006-01", it.Text[:7])
		if err != nil {
			continue
		}
		k := t.Format("Jan 2006")
		if counts[k] == 0 {
			order = append(order, k)
		}
		counts[k]++
	}
	var lines []string
	for _, k := range order {
		if len(lines) >= h {
			break
		}
		lines = append(lines, fmt.Sprintf("  %s %s", k, sDim.Render(fmt.Sprint(counts[k]))))
	}
	if len(lines) == 0 {
		lines = append(lines, sDim.Render("a to log a win"))
	}
	return strings.Join(lines, "\n")
}

func (s *winsScreen) mainTitle() string        { return "wins" }
func (s *winsScreen) mainView(w, h int) string { return s.ed.view(w, h) }

func (s *winsScreen) side() []sideBox {
	return []sideBox{{"why", "Log 1–2 lines a week: shipped work,\nhelp you gave, things you learned.\n\nThey feed the 1:1 side panel and\n✦ agenda draft, and make perf\nreviews easy."}}
}

func (s *winsScreen) hints() []hint {
	if s.ed.editing() {
		return []hint{{"enter", "save"}, {"esc", "cancel"}}
	}
	return []hint{{"a", "add win"}, {"enter", "edit"}, {"d", "delete"}, {"J/K", "move"}}
}
