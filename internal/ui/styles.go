package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var (
	cFg     = lipgloss.Color("#c9ced6")
	cDim    = lipgloss.Color("#8a919e")
	cAccent = lipgloss.Color("#e5a54b")
	cCyan   = lipgloss.Color("#6cc4d9")
	cGreen  = lipgloss.Color("#8fc97a")
	cRed    = lipgloss.Color("#e5737d")
	cPurple = lipgloss.Color("#b69be0")
	cBorder = lipgloss.Color("#2b303b")
	cSel    = lipgloss.Color("#262a33")
	cBar    = lipgloss.Color("#1a1d24")
	cBg     = lipgloss.Color("#111318")

	sFg     = lipgloss.NewStyle().Foreground(cFg)
	sDim    = lipgloss.NewStyle().Foreground(cDim)
	sAccent = lipgloss.NewStyle().Foreground(cAccent)
	sCyan   = lipgloss.NewStyle().Foreground(cCyan)
	sGreen  = lipgloss.NewStyle().Foreground(cGreen)
	sRed    = lipgloss.NewStyle().Foreground(cRed)
	sPurple = lipgloss.NewStyle().Foreground(cPurple)
	sHead   = lipgloss.NewStyle().Foreground(cAccent).Bold(true)
	sBold   = lipgloss.NewStyle().Foreground(cFg).Bold(true)
	sSel    = lipgloss.NewStyle().Background(cSel)
	sKey    = lipgloss.NewStyle().Foreground(cBg).Background(cCyan).Padding(0, 1)
	sBar    = lipgloss.NewStyle().Background(cBar)

	typeStyle = map[string]lipgloss.Style{"refactor": sPurple, "idea": sCyan, "improve": sGreen}
)

var ticketRe = regexp.MustCompile(`\b[A-Z][A-Z0-9]+-\d+\b`)

func highlightTickets(s string) string {
	return ticketRe.ReplaceAllStringFunc(s, func(id string) string { return sCyan.Render(id) })
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	return ansi.Truncate(s, w, "…")
}

func padRight(s string, w int) string {
	if gap := w - lipgloss.Width(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return truncate(s, w)
}

// box draws a rounded border with the title embedded in the top edge.
func box(title, content string, w, h int, focused bool) string {
	if w < 4 || h < 2 {
		return ""
	}
	border := sDim.Foreground(cBorder)
	titleStyle := sDim
	if focused {
		border = sAccent
		titleStyle = sAccent
	}
	inner := w - 2
	t := ""
	if title != "" {
		t = titleStyle.Render(" "+truncate(title, inner-4)+" ") + border.Render("")
	}
	top := border.Render("╭─") + t + border.Render(strings.Repeat("─", max(0, inner-1-lipgloss.Width(t)))+"╮")

	lines := strings.Split(content, "\n")
	body := make([]string, 0, h-2)
	for i := 0; i < h-2; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		body = append(body, border.Render("│")+" "+padRight(line, inner-2)+" "+border.Render("│"))
	}
	bottom := border.Render("╰" + strings.Repeat("─", inner) + "╯")
	return strings.Join(append(append([]string{top}, body...), bottom), "\n")
}

type hint struct{ key, desc string }

func renderHints(hints []hint, w int) string {
	var parts []string
	for _, h := range hints {
		parts = append(parts, sKey.Render(h.key)+sBar.Foreground(cFg).Render(" "+h.desc))
	}
	line := strings.Join(parts, sBar.Render("  "))
	return sBar.Width(w).Render(truncate(" "+line, w))
}
