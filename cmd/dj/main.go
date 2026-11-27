package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miguelparacuellos/dev-journal/internal/claude"
	"github.com/miguelparacuellos/dev-journal/internal/config"
	"github.com/miguelparacuellos/dev-journal/internal/store"
	"github.com/miguelparacuellos/dev-journal/internal/ui"
)

func main() {
	st := store.New(config.DefaultHome())
	if err := st.Scaffold(); err != nil {
		fail(err)
	}
	cfg, err := config.Load(st.Root)
	if err != nil {
		fail(err)
	}
	if len(os.Args) > 1 && os.Args[1] == "draft" {
		debugDraft(st, cfg)
		return
	}
	app, err := ui.New(st, cfg, clock())
	if err != nil {
		fail(err)
	}
	if _, err := tea.NewProgram(app, tea.WithAltScreen()).Run(); err != nil {
		fail(err)
	}
}

// clock pins "today" when DJ_DEMO_DATE is set, so the demo recording is reproducible.
func clock() func() time.Time {
	d := os.Getenv("DJ_DEMO_DATE")
	if d == "" {
		return time.Now
	}
	t, err := time.ParseInLocation(store.DateLayout, d, time.Local)
	if err != nil {
		fail(err)
	}
	return func() time.Time { return t.Add(9 * time.Hour) }
}

// debugDraft runs the Linear daily draft without the TUI and prints the raw result.
func debugDraft(st *store.Store, cfg config.Config) {
	today := store.Day(time.Now())
	since := today.AddDate(0, 0, -1)
	for since.Weekday() == time.Saturday || since.Weekday() == time.Sunday {
		since = since.AddDate(0, 0, -1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	d, err := claude.DraftDaily(ctx, claude.Options{
		Bin: cfg.ClaudeBin, Dir: st.Root, LinearTeam: cfg.LinearTeam, LinearAssignee: cfg.LinearAssignee,
		LinearTools: cfg.LinearTools, Language: cfg.Language,
	}, today, since, "")
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
