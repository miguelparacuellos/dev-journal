package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miguelparacuellos/dev-journal/internal/store"
)

type editMode int

const (
	modeNone editMode = iota
	modeAdd
	modeEdit
)

type row struct {
	sec  int
	item int // -1 = section header
}

// docEditor edits a sectioned Markdown doc as a navigable list of items.
type docEditor struct {
	doc       store.Doc
	checkable map[string]bool
	cursor    int
	offset    int
	mode      editMode
	input     textinput.Model
	// onAdd lets a screen rewrite new items (e.g. prefix a date).
	onAdd func(section, text string) string
	notes map[string]string
}

func newDocEditor(checkable ...string) docEditor {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 500
	ti.Cursor.Style = sAccent
	c := map[string]bool{}
	for _, s := range checkable {
		c[s] = true
	}
	return docEditor{checkable: c, input: ti}
}

func (e *docEditor) setDoc(d store.Doc) {
	e.doc = d
	e.mode = modeNone
	e.clampCursor()
}

func (e *docEditor) editing() bool { return e.mode != modeNone }

func (e *docEditor) rows() []row {
	var rs []row
	for si, s := range e.doc.Sections {
		rs = append(rs, row{si, -1})
		for ii := range s.Items {
			rs = append(rs, row{si, ii})
		}
	}
	return rs
}

func (e *docEditor) clampCursor() {
	n := len(e.rows())
	if e.cursor >= n {
		e.cursor = n - 1
	}
	if e.cursor < 0 {
		e.cursor = 0
	}
}

func (e *docEditor) current() (row, bool) {
	rs := e.rows()
	if e.cursor < 0 || e.cursor >= len(rs) {
		return row{}, false
	}
	return rs[e.cursor], true
}

func (e *docEditor) moveTo(sec, item int) {
	for i, r := range e.rows() {
		if r.sec == sec && r.item == item {
			e.cursor = i
			return
		}
	}
}

// Append adds items to a section by title (used by Claude drafts).
func (e *docEditor) Append(section string, texts ...string) {
	s := e.doc.Section(section)
	for _, t := range texts {
		s.Items = append(s.Items, store.Item{Text: t, Checkbox: e.checkable[s.Title]})
	}
}

// update returns changed=true when the doc must be saved.
func (e *docEditor) update(msg tea.KeyMsg) (changed bool, cmd tea.Cmd) {
	if e.mode != modeNone {
		return e.updateInput(msg)
	}
	r, ok := e.current()
	switch msg.String() {
	case "j", "down":
		if e.cursor < len(e.rows())-1 {
			e.cursor++
		}
	case "k", "up":
		if e.cursor > 0 {
			e.cursor--
		}
	case "tab":
		e.jumpSection(1)
	case "shift+tab":
		e.jumpSection(-1)
	case "a", "o":
		if ok {
			e.mode = modeAdd
			e.input.SetValue("")
			return false, e.input.Focus()
		}
	case "enter", "i":
		if ok && r.item >= 0 {
			e.mode = modeEdit
			e.input.SetValue(e.doc.Sections[r.sec].Items[r.item].Text)
			e.input.CursorEnd()
			return false, e.input.Focus()
		}
	case "x", " ":
		if ok && r.item >= 0 {
			it := &e.doc.Sections[r.sec].Items[r.item]
			if it.Checkbox {
				it.Done = !it.Done
				return true, nil
			}
		}
	case "d":
		if ok && r.item >= 0 {
			items := e.doc.Sections[r.sec].Items
			e.doc.Sections[r.sec].Items = append(items[:r.item:r.item], items[r.item+1:]...)
			e.clampCursor()
			return true, nil
		}
	case "J":
		if ok && r.item >= 0 && r.item < len(e.doc.Sections[r.sec].Items)-1 {
			items := e.doc.Sections[r.sec].Items
			items[r.item], items[r.item+1] = items[r.item+1], items[r.item]
			e.cursor++
			return true, nil
		}
	case "K":
		if ok && r.item > 0 {
			items := e.doc.Sections[r.sec].Items
			items[r.item], items[r.item-1] = items[r.item-1], items[r.item]
			e.cursor--
			return true, nil
		}
	}
	return false, nil
}

func (e *docEditor) jumpSection(dir int) {
	r, ok := e.current()
	if !ok || len(e.doc.Sections) == 0 {
		return
	}
	next := (r.sec + dir + len(e.doc.Sections)) % len(e.doc.Sections)
	e.moveTo(next, -1)
}

func (e *docEditor) updateInput(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.String() {
	case "esc":
		e.mode = modeNone
		e.input.Blur()
		return false, nil
	case "enter":
		text := strings.TrimSpace(e.input.Value())
		mode := e.mode
		e.mode = modeNone
		e.input.Blur()
		r, ok := e.current()
		if text == "" || !ok {
			return false, nil
		}
		sec := &e.doc.Sections[r.sec]
		if mode == modeEdit {
			sec.Items[r.item].Text = text
			return true, nil
		}
		if e.onAdd != nil {
			text = e.onAdd(sec.Title, text)
		}
		at := r.item + 1
		item := store.Item{Text: text, Checkbox: e.checkable[sec.Title]}
		sec.Items = append(sec.Items[:at], append([]store.Item{item}, sec.Items[at:]...)...)
		e.moveTo(r.sec, at)
		return true, nil
	}
	var cmd tea.Cmd
	e.input, cmd = e.input.Update(msg)
	return false, cmd
}

func (e *docEditor) view(w, h int) string {
	var lines []string
	cursorLine := 0
	rs := e.rows()
	inputLine := func(prefix string) string {
		e.input.Width = max(10, w-lipgloss.Width(prefix)-2)
		return prefix + e.input.View()
	}
	for i, r := range rs {
		sec := e.doc.Sections[r.sec]
		selected := i == e.cursor
		if selected {
			cursorLine = len(lines)
		}
		var line string
		if r.item < 0 {
			if len(lines) > 0 {
				lines = append(lines, "")
				if selected {
					cursorLine++
				}
			}
			title := strings.ToUpper(sec.Title)
			if title == "" {
				title = "—"
			}
			line = sHead.Render(title)
			if note := e.notes[sec.Title]; note != "" {
				line += " " + sPurple.Render(note)
			}
		} else {
			it := sec.Items[r.item]
			if selected && e.mode == modeEdit {
				line = inputLine(itemPrefix(it))
			} else {
				line = itemPrefix(it) + itemText(it, w-lipgloss.Width(itemPrefix(it)))
			}
		}
		if selected && e.mode != modeEdit {
			line = selectedRow(line, w)
		}
		lines = append(lines, line)
		if selected && e.mode == modeAdd {
			prefix := "• "
			if e.checkable[sec.Title] {
				prefix = "[ ] "
			}
			lines = append(lines, inputLine(sAccent.Render("+ ")+prefix))
			cursorLine = len(lines) - 1
		}
		if r.item < 0 && len(sec.Items) == 0 && !(selected && e.mode == modeAdd) {
			lines = append(lines, sDim.Render("  (empty — a to add)"))
		}
	}
	if cursorLine < e.offset {
		e.offset = cursorLine
	}
	if cursorLine >= e.offset+h {
		e.offset = cursorLine - h + 1
	}
	end := min(len(lines), e.offset+h)
	return strings.Join(lines[min(e.offset, end):end], "\n")
}

func itemPrefix(it store.Item) string {
	switch {
	case it.Plain:
		return ""
	case it.Checkbox && it.Done:
		return sGreen.Render("[x] ")
	case it.Checkbox:
		return sAccent.Render("[ ] ")
	}
	return sDim.Render("• ")
}

func itemText(it store.Item, w int) string {
	t := truncate(it.Text, w)
	if it.Checkbox && it.Done {
		return sDim.Render(t)
	}
	return highlightTickets(t)
}
