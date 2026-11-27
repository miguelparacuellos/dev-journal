package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miguelparacuellos/dev-journal/internal/store"
)

type backlogScreen struct {
	app      *App
	items    []store.BacklogItem
	typeIdx  int // -1 = all
	showDone bool
	cursor   int
}

func newBacklogScreen(a *App) *backlogScreen { return &backlogScreen{app: a, typeIdx: -1} }

func (s *backlogScreen) load() error {
	items, err := s.app.st.LoadBacklog()
	s.items = items
	s.clamp()
	return err
}

// visible returns indexes into s.items that pass the filters.
func (s *backlogScreen) visible() []int {
	var out []int
	for i, it := range s.items {
		if s.typeIdx >= 0 && it.Type != store.BacklogTypes[s.typeIdx] {
			continue
		}
		if !s.showDone && it.Status == "done" {
			continue
		}
		out = append(out, i)
	}
	return out
}

func (s *backlogScreen) clamp() {
	n := len(s.visible())
	if s.cursor >= n {
		s.cursor = n - 1
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
}

func (s *backlogScreen) selected() (*store.BacklogItem, int) {
	v := s.visible()
	if s.cursor >= len(v) {
		return nil, -1
	}
	return &s.items[v[s.cursor]], v[s.cursor]
}

func (s *backlogScreen) save() tea.Cmd {
	if err := s.app.st.SaveBacklog(s.items); err != nil {
		return errStatus(err)
	}
	return nil
}

func (s *backlogScreen) file() (string, error) {
	if len(s.items) == 0 {
		if err := s.app.st.SaveBacklog(nil); err != nil {
			return "", err
		}
	}
	return s.app.st.BacklogPath(), nil
}

func (s *backlogScreen) capturing() bool { return false }

func (s *backlogScreen) handleKey(msg tea.KeyMsg) tea.Cmd {
	it, idx := s.selected()
	switch msg.String() {
	case "j", "down":
		if s.cursor < len(s.visible())-1 {
			s.cursor++
		}
	case "k", "up":
		if s.cursor > 0 {
			s.cursor--
		}
	case "t":
		s.typeIdx++
		if s.typeIdx >= len(store.BacklogTypes) {
			s.typeIdx = -1
		}
		s.clamp()
	case ".":
		s.showDone = !s.showDone
		s.clamp()
	case "a", "o":
		return s.app.openPrompt("add › type: title #repo !S|M|L ›", "", s.add)
	case "enter", "i":
		if it != nil {
			return s.app.openPrompt("rename ›", it.Title, func(v string) tea.Cmd {
				it.Title = v
				return s.save()
			})
		}
	case "s":
		if it != nil {
			it.Status = cycle(store.BacklogStatuses, it.Status)
			cmd := s.save()
			s.clamp()
			return cmd
		}
	case "T":
		if it != nil {
			it.Type = cycle(store.BacklogTypes, it.Type)
			return s.save()
		}
	case "d":
		if idx >= 0 {
			s.items = append(s.items[:idx:idx], s.items[idx+1:]...)
			s.clamp()
			return s.save()
		}
	}
	return nil
}

func cycle(values []string, cur string) string {
	for i, v := range values {
		if v == cur {
			return values[(i+1)%len(values)]
		}
	}
	return values[0]
}

func (s *backlogScreen) add(input string) tea.Cmd {
	it := store.ParseQuickBacklog(input)
	if it.Title == "" {
		return errStatus(fmt.Errorf("empty title"))
	}
	it.Added = s.app.today().Format(store.DateLayout)
	s.items = append([]store.BacklogItem{it}, s.items...)
	s.typeIdx, s.cursor = -1, 0
	return tea.Batch(s.save(), status("added "+it.Type+": "+it.Title))
}

func (s *backlogScreen) info() string {
	open, started := 0, 0
	for _, it := range s.items {
		switch it.Status {
		case "open":
			open++
		case "started":
			started++
		}
	}
	return fmt.Sprintf("Backlog · %d open · %d started", open, started)
}

func (s *backlogScreen) listTitle() string { return "filter" }

func (s *backlogScreen) listView(w, h int) string {
	count := func(t string) int {
		n := 0
		for _, it := range s.items {
			if (t == "" || it.Type == t) && (s.showDone || it.Status != "done") {
				n++
			}
		}
		return n
	}
	var lines []string
	entry := func(idx int, label string, n int) {
		text := fmt.Sprintf("%s %s", label, sDim.Render(fmt.Sprint(n)))
		if idx == s.typeIdx {
			lines = append(lines, selectedRow("▸ "+text, w))
		} else {
			lines = append(lines, "  "+text)
		}
	}
	entry(-1, "all", count(""))
	for i, t := range store.BacklogTypes {
		entry(i, typeStyle[t].Render(t), count(t))
	}
	doneLabel := "hidden"
	if s.showDone {
		doneLabel = "shown"
	}
	lines = append(lines, "", sDim.Render("done: "+doneLabel+" (.)"))
	return strings.Join(lines, "\n")
}

func (s *backlogScreen) mainTitle() string { return "items" }

func (s *backlogScreen) mainView(w, h int) string {
	v := s.visible()
	if len(v) == 0 {
		return sDim.Render("Nothing here. Press a — e.g. \"refactor: Extract pricing rules #rates-service !M\"")
	}
	effW, stW, typeW := 4, 8, 9
	repoW := min(18, max(8, w/6))
	titleW := max(10, w-repoW-effW-stW-typeW-1)
	header := sDim.Render(padRight("TYPE", typeW) + padRight("TITLE", titleW+1) + padRight("REPO", repoW) + padRight("EFF", effW) + "STATUS")
	lines := []string{header}
	start := max(0, s.cursor-h+2)
	for i := start; i < len(v) && len(lines) < h; i++ {
		it := s.items[v[i]]
		st := sDim
		switch it.Status {
		case "started":
			st = sAccent
		case "done":
			st = sGreen
		}
		repo := it.Repo
		if repo == "" {
			repo = "—"
		}
		line := typeStyle[it.Type].Render(padRight(it.Type, typeW)) +
			padRight(truncate(it.Title, titleW), titleW+1) +
			sDim.Render(padRight(truncate(repo, repoW-1), repoW)) +
			padRight(it.Effort, effW) + st.Render(it.Status)
		if i == s.cursor {
			line = selectedRow(line, w)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (s *backlogScreen) side() []sideBox {
	it, _ := s.selected()
	if it == nil {
		return []sideBox{{"detail", ""}}
	}
	lines := []string{sBold.Render(truncate(it.Title, rightW-4))}
	meta := typeStyle[it.Type].Render(it.Type)
	if it.Added != "" {
		meta += sDim.Render(" · added " + it.Added)
	}
	lines = append(lines, meta, "")
	if it.Notes == "" {
		lines = append(lines, sDim.Render("No notes. Press e to add them"), sDim.Render("in backlog.md."))
	}
	for _, l := range strings.Split(it.Notes, "\n") {
		lines = append(lines, wrap(l, rightW-4)...)
	}
	return []sideBox{{"detail", highlightTickets(strings.Join(lines, "\n"))}}
}

func (s *backlogScreen) hints() []hint {
	return []hint{{"a", "add"}, {"s", "status"}, {"T", "type"}, {"t", "filter"}, {"enter", "rename"}, {"d", "delete"}}
}
