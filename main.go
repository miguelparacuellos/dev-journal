package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miguelparacuellos/dev-journal/internal/claude"
	"github.com/miguelparacuellos/dev-journal/internal/store"
	"github.com/miguelparacuellos/dev-journal/internal/ui"
)

func main() {
	st := store.New(store.DefaultRoot())
	if err := st.Scaffold(); err != nil {
		fail(err)
	}
	if len(os.Args) > 1 && os.Args[1] == "draft" {
		debugDraft(st)
		return
	}
	app, err := ui.New(st, time.Now)
	if err != nil {
		fail(err)
	}
	if _, err := tea.NewProgram(app, tea.WithAltScreen()).Run(); err != nil {
		fail(err)
	}
}

// debugDraft runs the Linear daily draft without the TUI and prints the raw result.
func debugDraft(st *store.Store) {
	today := store.Day(time.Now())
	since := today.AddDate(0, 0, -1)
	for since.Weekday() == time.Saturday || since.Weekday() == time.Sunday {
		since = since.AddDate(0, 0, -1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	d, err := claude.DraftDaily(ctx, st.Root, today, since, "")
	if err != nil {
		fail(err)
	}
	out, _ := json.MarshalIndent(d, "", "  ")
	fmt.Println(string(out))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "dj:", err)
	os.Exit(1)
}
