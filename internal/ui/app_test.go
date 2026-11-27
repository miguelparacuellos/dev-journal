package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/miguelparacuellos/dev-journal/internal/store"
)

func newTestApp(t *testing.T) (*App, *store.Store) {
	t.Helper()
	st := store.New(t.TempDir())
	if err := st.Scaffold(); err != nil {
		t.Fatal(err)
	}
	now := func() time.Time { return time.Date(2026, 11, 27, 9, 0, 0, 0, time.Local) }
	a, err := New(st, now)
	if err != nil {
		t.Fatal(err)
	}
	a.Update(tea.WindowSizeMsg{Width: 150, Height: 36})
	return a, st
}

func send(a *App, keys ...string) {
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "space":
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		_, cmd := a.Update(msg)
		drain(a, cmd)
	}
}

// drain runs synchronous follow-up commands (saves, reloads), skipping timers.
func drain(a *App, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				drain(a, c)
			}
			return
		}
		switch msg.(type) {
		case statusMsg, reloadMsg, pickResultMsg:
			_, next := a.Update(msg)
			_ = next
		}
	case <-time.After(50 * time.Millisecond):
	}
}

func typeText(a *App, s string) {
	for _, r := range s {
		send(a, string(r))
	}
}

func plain(a *App) string {
	v := ansi.Strip(a.View())
	if os.Getenv("DUMP") != "" {
		os.Stdout.WriteString(v + "\n\n")
	}
	return v
}

func TestDailyAddToggleAndPersist(t *testing.T) {
	a, st := newTestApp(t)
	send(a, "tab", "a")
	typeText(a, "Review DEV-412")
	send(a, "enter", "x")

	doc, ok, err := store.LoadDoc(st.DailyPath(a.today()))
	if err != nil || !ok {
		t.Fatalf("daily not saved: %v", err)
	}
	items := doc.Section(store.SecToday).Items
	if len(items) != 1 || items[0].Text != "Review DEV-412" || !items[0].Done {
		t.Fatalf("today = %+v", items)
	}
	if v := plain(a); !strings.Contains(v, "[x] Review DEV-412") {
		t.Fatalf("view missing item:\n%s", v)
	}
}

func TestOneOnOneCreateAndBacklogAdd(t *testing.T) {
	a, st := newTestApp(t)
	send(a, "2", "n", "enter")
	if len(st.OneOnOnes()) != 1 {
		t.Fatal("1:1 not created")
	}
	send(a, "a")
	typeText(a, "Ownership of payments")
	send(a, "enter")
	if v := plain(a); !strings.Contains(v, "Ownership of payments") || !strings.Contains(v, "TOPICS") {
		t.Fatalf("1:1 view:\n%s", v)
	}

	send(a, "3", "a")
	typeText(a, "refactor: Extract pricing #rates !M")
	send(a, "enter")
	items, _ := st.LoadBacklog()
	if len(items) != 1 || items[0].Repo != "rates" {
		t.Fatalf("backlog = %+v", items)
	}
	send(a, "s")
	items, _ = st.LoadBacklog()
	if items[0].Status != "started" {
		t.Fatalf("status = %s", items[0].Status)
	}
	plain(a)
}

func TestWinsGetDated(t *testing.T) {
	a, st := newTestApp(t)
	send(a, "4", "a")
	typeText(a, "Shipped refunds")
	send(a, "enter")
	if w := st.WinsSince(a.today()); len(w) != 1 || w[0] != "Shipped refunds" {
		t.Fatalf("wins = %+v", w)
	}
	plain(a)
}

func TestPickerAppliesChosenItems(t *testing.T) {
	a, st := newTestApp(t)
	scr := a.screens[0].(*dailyScreen)
	run := func() tea.Cmd {
		return func() tea.Msg {
			return pickResultMsg{items: []pickItem{
				{section: store.SecYesterday, text: "DEV-1 moved to review", on: true},
				{section: store.SecToday, text: "DEV-2 start", on: false},
			}}
		}
	}
	apply := func(chosen []pickItem) tea.Cmd {
		for _, it := range chosen {
			scr.ed.Append(it.section, it.text)
		}
		return scr.save()
	}
	drain(a, a.openPicker("draft from Linear", "test", run, apply))
	plain(a)
	send(a, "enter")
	doc, _, _ := store.LoadDoc(st.DailyPath(a.today()))
	if y := doc.Section(store.SecYesterday).Items; len(y) != 1 || len(doc.Section(store.SecToday).Items) != 0 {
		t.Fatalf("doc = %+v", doc)
	}
}
