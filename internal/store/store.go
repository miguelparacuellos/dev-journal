package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const DateLayout = "2006-01-02"

const (
	SecYesterday = "Yesterday"
	SecToday     = "Today"
	SecBlockers  = "Blockers"

	SecTopics   = "Topics"
	SecNotes    = "Notes"
	SecFeedback = "Feedback"
	SecActions  = "Actions"

	SecWins = "Wins"
)

type Store struct {
	Root string
}

func New(root string) *Store { return &Store{Root: root} }

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func (s *Store) path(parts ...string) string {
	return filepath.Join(append([]string{s.Root}, parts...)...)
}

func (s *Store) DailyPath(d time.Time) string {
	return s.path("daily", d.Format(DateLayout)+".md")
}

func (s *Store) OneOnOnePath(d time.Time) string {
	return s.path("one-on-ones", d.Format(DateLayout)+".md")
}

func (s *Store) BacklogPath() string { return s.path("backlog.md") }
func (s *Store) WinsPath() string    { return s.path("wins.md") }

func (s *Store) listDates(dir string) []time.Time {
	entries, _ := os.ReadDir(s.path(dir))
	var out []time.Time
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".md")
		if t, err := time.ParseInLocation(DateLayout, name, time.Local); err == nil {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].After(out[j]) })
	return out
}

func Day(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

// Dailies returns existing daily dates (newest first), always including today.
func (s *Store) Dailies(today time.Time) []time.Time {
	dates := s.listDates("daily")
	if len(dates) == 0 || !dates[0].Equal(today) {
		dates = append([]time.Time{today}, dates...)
	}
	return dates
}

// LoadDaily returns the saved daily, or a fresh one seeded from the latest
// earlier daily: its done Today items become Yesterday, open ones carry over.
func (s *Store) LoadDaily(d time.Time) (Doc, error) {
	doc, ok, err := LoadDoc(s.DailyPath(d))
	if err != nil || ok {
		ensureSections(&doc, SecYesterday, SecToday, SecBlockers)
		return doc, err
	}
	doc = Doc{Title: "Daily — " + d.Format(DateLayout)}
	ensureSections(&doc, SecYesterday, SecToday, SecBlockers)

	for _, prev := range s.listDates("daily") {
		if !prev.Before(d) {
			continue
		}
		prevDoc, _, err := LoadDoc(s.DailyPath(prev))
		if err != nil {
			return doc, err
		}
		for _, it := range prevDoc.Section(SecToday).Items {
			if it.Done {
				doc.Section(SecYesterday).Items = append(doc.Section(SecYesterday).Items, Item{Text: it.Text})
			} else {
				doc.Section(SecToday).Items = append(doc.Section(SecToday).Items, it)
			}
		}
		doc.Section(SecBlockers).Items = append(doc.Section(SecBlockers).Items, prevDoc.Section(SecBlockers).Items...)
		break
	}
	return doc, nil
}

func ensureSections(d *Doc, titles ...string) {
	for _, t := range titles {
		d.Section(t)
	}
}

func (s *Store) OneOnOnes() []time.Time { return s.listDates("one-on-ones") }

func (s *Store) LoadOneOnOne(d time.Time) (Doc, error) {
	doc, ok, err := LoadDoc(s.OneOnOnePath(d))
	if !ok {
		doc.Title = "1:1 — " + d.Format(DateLayout)
	}
	ensureSections(&doc, SecTopics, SecNotes, SecFeedback, SecActions)
	return doc, err
}

type OpenAction struct {
	From time.Time
	Text string
}

func (s *Store) OpenActionsBefore(d time.Time) []OpenAction {
	var out []OpenAction
	for _, date := range s.OneOnOnes() {
		if !date.Before(d) {
			continue
		}
		doc, _, err := LoadDoc(s.OneOnOnePath(date))
		if err != nil {
			continue
		}
		for _, it := range doc.Section(SecActions).Items {
			if it.Checkbox && !it.Done {
				out = append(out, OpenAction{From: date, Text: it.Text})
			}
		}
	}
	return out
}

func (s *Store) LoadWins() (Doc, error) {
	doc, ok, err := LoadDoc(s.WinsPath())
	if !ok {
		doc.Title = "Wins"
	}
	ensureSections(&doc, SecWins)
	return doc, err
}

// WinsSince expects win items formatted as "YYYY-MM-DD · text".
func (s *Store) WinsSince(d time.Time) []string {
	doc, err := s.LoadWins()
	if err != nil {
		return nil
	}
	var out []string
	for _, it := range doc.Section(SecWins).Items {
		date, text, ok := strings.Cut(it.Text, " · ")
		t, err := time.ParseInLocation(DateLayout, date, time.Local)
		if ok && err == nil && !t.Before(d) {
			out = append(out, text)
		}
	}
	return out
}

type LinearIssue struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type LinearCache struct {
	SyncedAt time.Time     `json:"syncedAt"`
	Issues   []LinearIssue `json:"issues"`
}

func (s *Store) linearCachePath() string { return s.path(".cache", "linear.json") }

func (s *Store) LoadLinearCache() LinearCache {
	var c LinearCache
	raw, err := os.ReadFile(s.linearCachePath())
	if err == nil {
		_ = json.Unmarshal(raw, &c)
	}
	return c
}

func (s *Store) SaveLinearCache(c LinearCache) error {
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(s.linearCachePath(), string(raw))
}
