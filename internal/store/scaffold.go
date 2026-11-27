package store

import (
	"os"
	"path/filepath"

	"github.com/miguelparacuellos/dev-journal/internal/config"
)

const claudeMD = `# dev-journal

Personal work journal of a developer. Plain Markdown, edited by the ` + "`dj`" + ` TUI and by hand.

## Layout

- ` + "`daily/YYYY-MM-DD.md`" + ` — standup entry. Sections: ` + "`## Yesterday`" + ` (bullets), ` + "`## Today`" + ` (checkboxes), ` + "`## Blockers`" + ` (bullets).
- ` + "`one-on-ones/YYYY-MM-DD.md`" + ` — monthly 1:1 with my manager. Sections: ` + "`## Topics`, `## Notes`, `## Feedback`, `## Actions`" + ` (checkboxes; open ones carry to the next 1:1).
- ` + "`backlog.md`" + ` — refactors / ideas / improvements. Each item is ` + "`## Title`" + ` followed by one meta line
  ` + "`type: refactor|idea|improve · repo: <name> · effort: S|M|L · status: open|started|done · added: YYYY-MM-DD`" + `, then free notes.
- ` + "`wins.md`" + ` — brag log, ` + "`## Wins`" + ` with bullets ` + "`- YYYY-MM-DD · text`" + `.
- ` + "`config.env`" + ` — personal settings.
- ` + "`.cache/`" + ` — TUI cache, ignore.

## Rules

- Keep the exact section headings and line formats above: the TUI parses them.
- Settings (Linear team, assignee, language…) live in ` + "`config.env`" + `; read it before using Linear.
- Write concise, standup-ready text in the ` + "`DJ_LANGUAGE`" + ` from ` + "`config.env`" + `.
`

const dailyDraftCmd = `---
description: Draft today's daily from Linear (current cycle)
---
Read config.env for DJ_LINEAR_TEAM (empty = all my teams) and DJ_LINEAR_ASSIGNEE.
Using the Linear connector, look at issues in that team's current cycle assigned to that person.
Find what changed since the previous working day (status changes, comments, PRs) and update
today's daily file in daily/ (create it if missing, keeping the section format from CLAUDE.md):
- Yesterday: what I moved forward, one bullet per issue, "ABC-123 Title — what happened".
- Today: checkboxes for the in-progress issues and anything next by priority.
- Blockers: issues marked blocked or waiting on someone.
Keep existing manual bullets. Show me the diff when done.
`

const oneOnOnePrepCmd = `---
description: Prepare the next 1:1 agenda
---
Read the dailies since the last file in one-on-ones/, wins.md, and the open Actions of previous 1:1s.
Propose 3–6 topics for the next 1:1 (progress worth highlighting, decisions I need, feedback to ask for,
growth/career points). Add them under ## Topics of the next 1:1 file, creating it if needed.
`

func (s *Store) Scaffold() error {
	files := map[string]string{
		"CLAUDE.md":                           claudeMD,
		".gitignore":                          ".cache/\n",
		config.FileName:                       config.Template,
		".claude/commands/daily-draft.md":     dailyDraftCmd,
		".claude/commands/one-on-one-prep.md": oneOnOnePrepCmd,
	}
	for rel, content := range files {
		p := filepath.Join(s.Root, rel)
		if _, err := os.Stat(p); err == nil {
			continue
		}
		if err := writeFile(p, content); err != nil {
			return err
		}
	}
	for _, dir := range []string{"daily", "one-on-ones"} {
		if err := os.MkdirAll(filepath.Join(s.Root, dir), 0o755); err != nil {
			return err
		}
	}
	return nil
}
