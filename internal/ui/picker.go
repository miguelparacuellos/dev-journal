package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miguelparacuellos/dev-journal/internal/claude"
)

type pickItem struct {
	section, text, why string
	on                 bool
}

// picker shows Claude's suggestions so the user chooses what goes into the doc.
type picker struct {
	title, cmdline string
	items          []pickItem
	cursor         int
	loading        bool
	err            error
	started        time.Time
	run            func() tea.Cmd
	apply          func([]pickItem) tea.Cmd
}

func (p *picker) start(a *App) tea.Cmd {
	p.loading, p.err, p.items, p.started = true, nil, nil, time.Now()
	return tea.Batch(a.spin.Tick, p.run())
}

func toPickItems(section string, sugg []claude.Suggestion) []pickItem {
	var out []pickItem
	for _, s := range sugg {
		if strings.TrimSpace(s.Text) == "" {
			continue
		}
		out = append(out, pickItem{section: section, text: s.Text, why: s.Why, on: s.Default == nil || *s.Default})
	}
	return out
}

func (p *picker) update(a *App, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		a.picker = nil
		return nil
	case "r":
		if !p.loading {
			return p.start(a)
		}
	}
	if p.loading || p.err != nil {
		return nil
	}
	switch msg.String() {
	case "j", "down":
		if p.cursor < len(p.items)-1 {
			p.cursor++
		}
	case "k", "up":
		if p.cursor > 0 {
			p.cursor--
		}
	case " ", "x":
		if p.cursor < len(p.items) {
			p.items[p.cursor].on = !p.items[p.cursor].on
		}
	case "enter":
		var chosen []pickItem
		for _, it := range p.items {
			if it.on {
				chosen = append(chosen, it)
			}
		}
		a.picker = nil
		return p.apply(chosen)
	}
	return nil
}

func (p *picker) hints() []hint {
	if p.loading {
		return []hint{{"esc", "cancel"}}
	}
	if p.err != nil {
		return []hint{{"r", "retry"}, {"esc", "close"}}
	}
	return []hint{{"space", "toggle"}, {"enter", "insert"}, {"r", "regenerate"}, {"esc", "cancel"}}
}

func (p *picker) view(a *App, w, h int) string {
	var lines []string
	lines = append(lines, sDim.Render(truncate("$ "+p.cmdline, w)), "")
	switch {
	case p.loading:
		lines = append(lines, a.spin.View()+" asking Claude… "+sDim.Render(fmt.Sprintf("%ds", int(time.Since(p.started).Seconds()))))
		lines = append(lines, "", sDim.Render("Headless `claude -p` run; takes ~10–60s."))
	case p.err != nil:
		lines = append(lines, sRed.Render("Claude failed:"), "")
		for _, l := range wrap(p.err.Error(), w) {
			lines = append(lines, sRed.Render(l))
		}
	case len(p.items) == 0:
		lines = append(lines, sDim.Render("Nothing to suggest."))
	default:
		cursorLine := 0
		section := ""
		for i, it := range p.items {
			if it.section != section {
				section = it.section
				if i > 0 {
					lines = append(lines, "")
				}
				lines = append(lines, sHead.Render(strings.ToUpper(section)))
			}
			box := sDim.Render("[ ] ")
			if it.on {
				box = sAccent.Render("[x] ")
			}
			line := box + highlightTickets(truncate(it.text, w-4))
			if i == p.cursor {
				line = sSel.Render(padRight(line, w))
				cursorLine = len(lines)
			}
			lines = append(lines, line)
			if it.why != "" {
				lines = append(lines, "    "+sDim.Render(truncate(it.why, w-4)))
			}
		}
		if cursorLine >= h-1 {
			lines = lines[cursorLine-h+3:]
		}
	}
	if len(lines) > h {
		lines = lines[:h]
	}
	return strings.Join(lines, "\n")
}

func wrap(s string, w int) []string {
	var out []string
	for _, para := range strings.Split(s, "\n") {
		for len(para) > w && w > 0 {
			out = append(out, para[:w])
			para = para[w:]
		}
		out = append(out, para)
	}
	return out
}
