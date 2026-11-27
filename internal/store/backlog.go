package store

import (
	"os"
	"strings"
)

var (
	BacklogTypes    = []string{"refactor", "idea", "improve"}
	BacklogStatuses = []string{"open", "started", "done"}
)

type BacklogItem struct {
	Title  string
	Type   string
	Repo   string
	Effort string
	Status string
	Added  string
	Notes  string
}

func ParseBacklog(src string) []BacklogItem {
	var items []BacklogItem
	var cur *BacklogItem
	var notes []string
	flush := func() {
		if cur != nil {
			cur.Notes = strings.TrimSpace(strings.Join(notes, "\n"))
			items = append(items, *cur)
		}
		notes = nil
	}
	for _, line := range strings.Split(src, "\n") {
		switch {
		case strings.HasPrefix(line, "## "):
			flush()
			cur = &BacklogItem{Title: strings.TrimPrefix(line, "## "), Status: "open"}
		case cur == nil:
			continue
		case len(notes) == 0 && isMetaLine(line):
			parseMeta(cur, line)
		default:
			notes = append(notes, line)
		}
	}
	flush()
	return items
}

func isMetaLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "type:")
}

func parseMeta(it *BacklogItem, line string) {
	for _, part := range strings.Split(line, " · ") {
		k, v, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		switch strings.TrimSpace(k) {
		case "type":
			it.Type = v
		case "repo":
			it.Repo = v
		case "effort":
			it.Effort = v
		case "status":
			it.Status = v
		case "added":
			it.Added = v
		}
	}
}

func RenderBacklog(items []BacklogItem) string {
	var b strings.Builder
	b.WriteString("# Backlog\n")
	for _, it := range items {
		b.WriteString("\n## " + it.Title + "\n")
		meta := []string{"type: " + it.Type}
		if it.Repo != "" {
			meta = append(meta, "repo: "+it.Repo)
		}
		if it.Effort != "" {
			meta = append(meta, "effort: "+it.Effort)
		}
		meta = append(meta, "status: "+it.Status)
		if it.Added != "" {
			meta = append(meta, "added: "+it.Added)
		}
		b.WriteString(strings.Join(meta, " · ") + "\n")
		if it.Notes != "" {
			b.WriteString("\n" + it.Notes + "\n")
		}
	}
	return b.String()
}

// ParseQuickBacklog turns "refactor: Extract rules #rates-service !M" into an item.
func ParseQuickBacklog(input string) BacklogItem {
	it := BacklogItem{Type: "idea", Status: "open"}
	if prefix, rest, ok := strings.Cut(input, ":"); ok {
		for _, t := range BacklogTypes {
			if strings.EqualFold(strings.TrimSpace(prefix), t) {
				it.Type = t
				input = rest
			}
		}
	}
	var words []string
	for _, w := range strings.Fields(input) {
		switch {
		case strings.HasPrefix(w, "#") && len(w) > 1:
			it.Repo = w[1:]
		case strings.HasPrefix(w, "!") && len(w) > 1:
			it.Effort = strings.ToUpper(w[1:])
		default:
			words = append(words, w)
		}
	}
	it.Title = strings.Join(words, " ")
	return it
}

func (s *Store) LoadBacklog() ([]BacklogItem, error) {
	raw, err := os.ReadFile(s.BacklogPath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ParseBacklog(string(raw)), nil
}

func (s *Store) SaveBacklog(items []BacklogItem) error {
	return writeFile(s.BacklogPath(), RenderBacklog(items))
}
