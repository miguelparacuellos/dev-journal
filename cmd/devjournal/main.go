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
		case "task":
			task, e := app.CreateTask(strings.Join(args[1:], " "))
			if e == nil {
				fmt.Println("Task saved:", task.ID)
			}
			return e
		case "tasks":
			printTasks(app.Tasks())
			return nil
		case "topic":
			_, e := app.AddTopic(*day, strings.Join(args[1:], " "))
			if e == nil {
				fmt.Println("O2O topic saved; it stays open until addressed")
			}
			return e
		case "topics":
			for _, topic := range app.OpenTopics() {
				fmt.Printf("%s  %s  %s\n", topic.ID, topic.Day, topic.Text)
			}
			return nil
		case "plan", "complete", "unplan":
			if len(args) != 2 {
				return fmt.Errorf("use %s TASK_ID", args[0])
			}
			switch args[0] {
			case "plan":
				err = app.PlanTask(*day, args[1])
			case "unplan":
				err = app.UnplanTask(*day, args[1])
			case "complete":
				err = app.CompleteTask(args[1])
			}
			if err == nil {
				fmt.Println("Saved:", args[0])
			}
			return err
		case "planned":
			fmt.Println("Plan for", *day)
			printTasks(app.Plan(*day))
			return nil
		case "log":
			for _, e := range app.Entries(*day) {
				fmt.Printf("%s  %s\n", e.ID, e.Text)
			}
			return nil
		default:
			return fmt.Errorf("unknown command %q; use add, log, task, tasks, plan, unplan, planned, complete, topic, topics, or no command for the TUI", args[0])
		}
	}
	if os.Getenv("NO_COLOR") != "" {
		*theme = "mono"
	}
	_, err = tea.NewProgram(newModel(app, *day, *theme, *ascii)).Run()
	return err
}

func printTasks(tasks []journal.Task) {
	for _, task := range tasks {
		state := "open"
		if task.Completed {
			state = "done"
		}
		fmt.Printf("%s  [%s] %s\n", task.ID, state, task.Text)
	}
}
