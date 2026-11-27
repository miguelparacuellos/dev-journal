// Package config resolves user-specific settings. Precedence: environment
// variables, then $DJ_HOME/config.env, then built-in defaults.
package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Home           string
	LinearTeam     string
	LinearAssignee string
	LinearTools    []string
	Language       string
	ClaudeBin      string
	Editor         string
	ClipboardCmd   string
}

const FileName = "config.env"

var defaults = map[string]string{
	"DJ_LINEAR_TEAM":     "",
	"DJ_LINEAR_ASSIGNEE": "me",
	"DJ_LINEAR_TOOLS":    "mcp__claude_ai_Linear",
	"DJ_LANGUAGE":        "English",
	"DJ_CLAUDE_BIN":      "claude",
	"DJ_EDITOR":          "",
	"DJ_CLIPBOARD_CMD":   "pbcopy",
}

func DefaultHome() string {
	if h := os.Getenv("DJ_HOME"); h != "" {
		return h
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "dev-journal")
}

func Load(home string) (Config, error) {
	file, err := readEnvFile(filepath.Join(home, FileName))
	if err != nil {
		return Config{}, err
	}
	get := func(key string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		if v, ok := file[key]; ok && v != "" {
			return v
		}
		return defaults[key]
	}
	editor := get("DJ_EDITOR")
	for _, k := range []string{"VISUAL", "EDITOR"} {
		if editor == "" {
			editor = os.Getenv(k)
		}
	}
	if editor == "" {
		editor = "vi"
	}
	return Config{
		Home:           home,
		LinearTeam:     get("DJ_LINEAR_TEAM"),
		LinearAssignee: get("DJ_LINEAR_ASSIGNEE"),
		LinearTools:    splitList(get("DJ_LINEAR_TOOLS")),
		Language:       get("DJ_LANGUAGE"),
		ClaudeBin:      get("DJ_CLAUDE_BIN"),
		Editor:         editor,
		ClipboardCmd:   get("DJ_CLIPBOARD_CMD"),
	}, nil
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// readEnvFile parses KEY=VALUE lines; # comments and optional quotes allowed.
func readEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(strings.TrimPrefix(line, "export "), "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
	}
	return out, sc.Err()
}

// Template is written to a new journal so users see every knob.
const Template = `# dev-journal settings. Environment variables with the same name take precedence.

# Linear team key whose active cycle feeds the daily draft (e.g. ENG). Empty = all my teams.
DJ_LINEAR_TEAM=
# Whose issues to read; "me" is the user authenticated in the Linear connector.
DJ_LINEAR_ASSIGNEE=me
# Comma-separated Claude Code tool/server names allowed during the Linear draft.
DJ_LINEAR_TOOLS=mcp__claude_ai_Linear
# Language Claude writes drafts in.
DJ_LANGUAGE=English
# Commands. DJ_EDITOR falls back to $VISUAL / $EDITOR / vi.
DJ_CLAUDE_BIN=claude
DJ_EDITOR=
DJ_CLIPBOARD_CMD=pbcopy
`
