package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"devjournal/journal"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
func run() error {
	home, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet("devjournal", flag.ContinueOnError)
	path := flags.String("data", filepath.Join(home, "devjournal", "journal.json"), "Local journal file")
	theme := flags.String("theme", "dark", "dark, light, or mono")
	ascii := flags.Bool("ascii", false, "Use ASCII decorations")
	day := flags.String("date", time.Now().Format("2006-01-02"), "Workday YYYY-MM-DD")
	if err = flags.Parse(os.Args[1:]); err != nil {
		return err
	}
	if *theme != "dark" && *theme != "light" && *theme != "mono" {
		return fmt.Errorf("unknown theme: %s", *theme)
	}
	if _, err = time.Parse("2006-01-02", *day); err != nil {
		return fmt.Errorf("date must be YYYY-MM-DD")
	}
	app, err := journal.Open(*path)
	if err != nil {
		return err
	}
	args := flags.Args()
	if len(args) > 0 {
		switch args[0] {
		case "add":
			text := strings.Join(args[1:], " ")
			if text == "-" {
				b, e := io.ReadAll(os.Stdin)
				if e != nil {
					return e
				}
				text = string(b)
			}
			_, err = app.Capture(*day, text)
			if err == nil {
				fmt.Println("Saved to", *day)
			}
			return err
		case "log":
			for _, e := range app.Entries(*day) {
				fmt.Printf("%s  %s\n", e.ID, e.Text)
			}
			return nil
		default:
			return fmt.Errorf("unknown command %q; use add, log, or no command for the TUI", args[0])
		}
	}
	if os.Getenv("NO_COLOR") != "" {
		*theme = "mono"
	}
	_, err = tea.NewProgram(newModel(app, *day, *theme, *ascii)).Run()
	return err
}
