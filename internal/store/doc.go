package store

import (
	"os"
	"strings"
)

type Item struct {
	Text     string
	Checkbox bool
	Done     bool
	Plain    bool
}

type Section struct {
	Title string
	Items []Item
}

type Doc struct {
	Title    string
	Sections []Section
}

func (d *Doc) Section(title string) *Section {
	for i := range d.Sections {
		if strings.EqualFold(d.Sections[i].Title, title) {
			return &d.Sections[i]
		}
	}
	d.Sections = append(d.Sections, Section{Title: title})
	return &d.Sections[len(d.Sections)-1]
}

func ParseDoc(src string) Doc {
	var d Doc
	var cur *Section
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			continue
		case strings.HasPrefix(trimmed, "# ") && d.Title == "" && cur == nil:
			d.Title = strings.TrimPrefix(trimmed, "# ")
		case strings.HasPrefix(trimmed, "## "):
			d.Sections = append(d.Sections, Section{Title: strings.TrimPrefix(trimmed, "## ")})
			cur = &d.Sections[len(d.Sections)-1]
		default:
			if cur == nil {
				d.Sections = append(d.Sections, Section{})
				cur = &d.Sections[len(d.Sections)-1]
			}
			cur.Items = append(cur.Items, parseItem(trimmed))
		}
	}
	return d
}

func parseItem(line string) Item {
	switch {
	case strings.HasPrefix(line, "- [ ] "):
		return Item{Text: line[6:], Checkbox: true}
	case strings.HasPrefix(line, "- [x] "), strings.HasPrefix(line, "- [X] "):
		return Item{Text: line[6:], Checkbox: true, Done: true}
	case strings.HasPrefix(line, "- "), strings.HasPrefix(line, "* "):
		return Item{Text: line[2:]}
	}
	return Item{Text: line, Plain: true}
}

func (d Doc) Render() string {
	var b strings.Builder
	if d.Title != "" {
		b.WriteString("# " + d.Title + "\n")
	}
	for _, s := range d.Sections {
		if s.Title != "" {
			b.WriteString("\n## " + s.Title + "\n")
		}
		for _, it := range s.Items {
			b.WriteString(it.Markdown() + "\n")
		}
	}
	return b.String()
}

func (it Item) Markdown() string {
	switch {
	case it.Plain:
		return it.Text
	case it.Checkbox && it.Done:
		return "- [x] " + it.Text
	case it.Checkbox:
		return "- [ ] " + it.Text
	}
	return "- " + it.Text
}

func LoadDoc(path string) (Doc, bool, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Doc{}, false, nil
	}
	if err != nil {
		return Doc{}, false, err
	}
	return ParseDoc(string(raw)), true, nil
}

func SaveDoc(path string, d Doc) error {
	return writeFile(path, d.Render())
}
