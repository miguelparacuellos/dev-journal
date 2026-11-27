package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miguelparacuellos/dev-journal/internal/store"
)

type screen interface {
	load() error
	listTitle() string
	listView(w, h int) string
	mainTitle() string
	mainView(w, h int) string
	side() []sideBox
	hints() []hint
	capturing() bool
	handleKey(tea.KeyMsg) tea.Cmd
	// file saves pending changes and returns the file `e` should open.
	file() (string, error)
	info() string
}

type sideBox struct{ title, content string }

type (
	statusMsg struct {
		text  string
		isErr bool
	}
	clearStatusMsg struct{ id int }
	reloadMsg      struct{}
	pickResultMsg  struct {
		items []pickItem
		err   error
	}
)

func status(text string) tea.Cmd { return func() tea.Msg { return statusMsg{text: text} } }
func errStatus(err error) tea.Cmd {
	return func() tea.Msg { return statusMsg{text: err.Error(), isErr: true} }
}

type App struct {
	st       *store.Store
	now      func() time.Time
	screens  []screen
	names    []string
	active   int
	w, h     int
	picker   *picker
	prompt   *prompt
	spin     spinner.Model
	status   string
	statusID int
	isErr    bool
}

func New(st *store.Store, now func() time.Time) (*App, error) {
	a := &App{st: st, now: now, spin: spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(sPurple))}
	a.screens = []screen{newDailyScreen(a), newOneOnOneScreen(a), newBacklogScreen(a), newWinsScreen(a)}
	a.names = []string{"Daily", "1:1s", "Backlog", "Wins"}
	for _, s := range a.screens {
		if err := s.load(); err != nil {
			return nil, err
		}
	}
	return a, nil
}

func (a *App) today() time.Time { return store.Day(a.now()) }

func (a *App) Init() tea.Cmd { return nil }

func (a *App) openPrompt(label, initial string, submit func(string) tea.Cmd) tea.Cmd {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 300
	ti.SetValue(initial)
	ti.CursorEnd()
	ti.Width = max(20, a.w-lipgloss.Width(label)-6)
	a.prompt = &prompt{label: label, input: ti, submit: submit}
	return a.prompt.input.Focus()
}

func (a *App) openPicker(title, cmdline string, run func() tea.Cmd, apply func([]pickItem) tea.Cmd) tea.Cmd {
	a.picker = &picker{title: title, cmdline: cmdline, run: run, apply: apply}
	return a.picker.start(a)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
		return a, nil
	case spinner.TickMsg:
		if a.picker != nil && a.picker.loading {
			var cmd tea.Cmd
			a.spin, cmd = a.spin.Update(msg)
			return a, cmd
		}
		return a, nil
	case pickResultMsg:
		if a.picker != nil {
			a.picker.loading = false
			a.picker.items, a.picker.err = msg.items, msg.err
			a.picker.cursor = 0
		}
		return a, nil
	case statusMsg:
		a.statusID++
		a.status, a.isErr = msg.text, msg.isErr
		id := a.statusID
		return a, tea.Tick(5*time.Second, func(time.Time) tea.Msg { return clearStatusMsg{id} })
	case clearStatusMsg:
		if msg.id == a.statusID {
			a.status = ""
		}
		return a, nil
	case reloadMsg:
		for _, s := range a.screens {
			if err := s.load(); err != nil {
				return a, errStatus(err)
			}
		}
		return a, nil
	case tea.KeyMsg:
		return a, a.handleKey(msg)
	}
	return a, nil
}

func (a *App) handleKey(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "ctrl+c" {
		return tea.Quit
	}
	if a.prompt != nil {
		return a.prompt.update(a, msg)
	}
	if a.picker != nil {
		return a.picker.update(a, msg)
	}
	cur := a.screens[a.active]
	if cur.capturing() {
		return cur.handleKey(msg)
	}
	switch k := msg.String(); k {
	case "q":
		return tea.Quit
	case "1", "2", "3", "4":
		a.active = int(k[0] - '1')
		return func() tea.Msg { return reloadMsg{} }
	case "c":
		return a.execIn(claudeBin())
	case "e":
		path, err := cur.file()
		if err != nil {
			return errStatus(err)
		}
		return a.execIn(editor(), path)
	}
	return cur.handleKey(msg)
}

func claudeBin() string {
	if b := os.Getenv("DJ_CLAUDE_BIN"); b != "" {
		return b
	}
	return "claude"
}

func editor() string {
	for _, k := range []string{"VISUAL", "EDITOR"} {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return "vi"
}

// execIn suspends the TUI and runs an interactive program in the journal dir.
func (a *App) execIn(bin string, args ...string) tea.Cmd {
	parts := strings.Fields(bin)
	cmd := exec.Command(parts[0], append(parts[1:], args...)...)
	cmd.Dir = a.st.Root
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return statusMsg{text: fmt.Sprintf("%s: %v", parts[0], err), isErr: true}
		}
		return reloadMsg{}
	})
}

func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			return statusMsg{text: "copy failed: " + err.Error(), isErr: true}
		}
		return statusMsg{text: "copied to clipboard"}
	}
}

const (
	leftW  = 24
	rightW = 42
)

func (a *App) View() string {
	if a.w == 0 {
		return ""
	}
	cur := a.screens[a.active]

	header := sAccent.Bold(true).Render(" dev-journal") + "  " + sDim.Render(cur.info())
	if a.status != "" {
		st := sGreen
		if a.isErr {
			st = sRed
		}
		msg := truncate(a.status, max(10, a.w-lipgloss.Width(header)-4))
		header += strings.Repeat(" ", max(1, a.w-lipgloss.Width(header)-lipgloss.Width(msg)-1)) + st.Render(msg)
	}

	bh := a.h - 2
	rw := 0
	if a.w >= 110 {
		rw = rightW
	}
	mw := a.w - leftW - rw

	var nav []string
	for i, n := range a.names {
		if i == a.active {
			nav = append(nav, sSel.Render(padRight(sAccent.Render(fmt.Sprintf("%d", i+1))+" ▸ "+sBold.Render(n), leftW-4)))
		} else {
			nav = append(nav, sDim.Render(fmt.Sprintf("%d", i+1))+"   "+n)
		}
	}
	navH := len(a.names) + 2
	left := lipgloss.JoinVertical(lipgloss.Left,
		box("sections", strings.Join(nav, "\n"), leftW, navH, false),
		box(cur.listTitle(), cur.listView(leftW-4, bh-navH-2), leftW, bh-navH, false),
	)

	var main string
	if a.picker != nil {
		main = box("✦ "+a.picker.title, a.picker.view(a, mw-4, bh-2), mw, bh, true)
	} else {
		main = box(cur.mainTitle(), cur.mainView(mw-4, bh-2), mw, bh, true)
	}

	cols := []string{left, main}
	if rw > 0 {
		boxes := cur.side()
		var rendered []string
		remaining := bh
		for i, b := range boxes {
			h := remaining / (len(boxes) - i)
			if i == len(boxes)-1 {
				h = remaining
			}
			remaining -= h
			title := b.title
			rendered = append(rendered, box(title, b.content, rw, h, strings.HasPrefix(title, "✦")))
		}
		cols = append(cols, lipgloss.JoinVertical(lipgloss.Left, rendered...))
	}

	var bottom string
	switch {
	case a.prompt != nil:
		bottom = sBar.Width(a.w).Render(" " + sAccent.Render(a.prompt.label) + " " + a.prompt.input.View())
	case a.picker != nil:
		bottom = renderHints(a.picker.hints(), a.w)
	default:
		bottom = renderHints(append(cur.hints(), hint{"e", "editor"}, hint{"c", "claude"}, hint{"q", "quit"}), a.w)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinHorizontal(lipgloss.Top, cols...), bottom)
}

type prompt struct {
	label  string
	input  textinput.Model
	submit func(string) tea.Cmd
}

func (p *prompt) update(a *App, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		a.prompt = nil
		return nil
	case "enter":
		a.prompt = nil
		if v := strings.TrimSpace(p.input.Value()); v != "" {
			return p.submit(v)
		}
		return nil
	}
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	return cmd
}
