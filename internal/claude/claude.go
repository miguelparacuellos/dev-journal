package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Suggestion struct {
	Section string `json:"-"`
	Text    string `json:"text"`
	Why     string `json:"why"`
	Default *bool  `json:"default,omitempty"`
}

type Issue struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type DailyDraft struct {
	Yesterday []Suggestion `json:"yesterday"`
	Today     []Suggestion `json:"today"`
	Blockers  []Suggestion `json:"blockers"`
	Issues    []Issue      `json:"issues"`
}

type AgendaDraft struct {
	Topics []Suggestion `json:"topics"`
}

type Options struct {
	Bin            string
	Dir            string
	LinearTeam     string
	LinearAssignee string
	LinearTools    []string
	Language       string
}

func (o Options) teamScope() string {
	if o.LinearTeam == "" {
		return "each team I belong to"
	}
	return "team " + o.LinearTeam
}

func (o Options) assignee() string {
	if o.LinearAssignee == "" || o.LinearAssignee == "me" {
		return `me ("me")`
	}
	return o.LinearAssignee
}

// Run calls `claude -p` headless and returns the final text result.
func Run(ctx context.Context, o Options, prompt string, allowedTools []string) (string, error) {
	args := []string{"-p", prompt, "--output-format", "json"}
	if len(allowedTools) > 0 {
		args = append(args, "--allowedTools", strings.Join(allowedTools, ","))
	}
	cmd := exec.CommandContext(ctx, o.Bin, args...)
	cmd.Dir = o.Dir
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("claude: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("claude: %w", err)
	}
	var res struct {
		Result  string `json:"result"`
		IsError bool   `json:"is_error"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		return "", fmt.Errorf("claude: unexpected output: %w", err)
	}
	if res.IsError {
		return "", fmt.Errorf("claude: %s", res.Result)
	}
	return res.Result, nil
}

func extractJSON(s string, v any) error {
	start, end := strings.Index(s, "{"), strings.LastIndex(s, "}")
	if start < 0 || end <= start {
		return fmt.Errorf("no JSON in claude reply: %q", truncate(s, 200))
	}
	return json.Unmarshal([]byte(s[start:end+1]), v)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

const suggestionShape = `{"text": "...", "why": "short reason", "default": true|false}`

func DraftDaily(ctx context.Context, o Options, today, since time.Time, existing string) (DailyDraft, error) {
	prompt := fmt.Sprintf(`You prepare my daily standup. Use the Linear tools only (read-only, never modify anything).

1. Find the active cycle of %[1]s and list the issues in it assigned to %[6]s.
2. For each, check what changed since %[2]s (status changes, comments, linked PRs).
3. Today is %[3]s.

Current daily entry (do not repeat what is already there):
---
%[4]s
---

Reply with ONLY a JSON object, no prose, no code fences:
{
  "yesterday": [%[5]s],   // what I moved forward since %[2]s, "ABC-123 short title — what happened"
  "today":     [%[5]s],   // in-progress issues and next by priority, "ABC-123 short title"
  "blockers":  [%[5]s],   // real blockers only: blocked-by relations, waiting on another team/person
  "issues":    [{"id": "ABC-123", "title": "...", "status": "In Progress"}]
}
"issues": only my cycle issues that are not Done/Canceled, plus those completed since %[2]s; active ones first.
Waiting for a normal code review is not a blocker unless it has been waiting more than 2 working days.
Keep each "text" under 90 characters and write it in %[7]s. Set "default": false for noise (e.g. only a comment from me).`,
		o.teamScope(), since.Format("Mon 2006-01-02"), today.Format("Mon 2006-01-02"), existing, suggestionShape, o.assignee(), o.Language)

	text, err := Run(ctx, o, prompt, o.LinearTools)
	if err != nil {
		return DailyDraft{}, err
	}
	var d DailyDraft
	return d, extractJSON(text, &d)
}

func DraftAgenda(ctx context.Context, o Options, notes string) (AgendaDraft, error) {
	prompt := fmt.Sprintf(`Help me prepare my next monthly 1:1 with my manager.
Below are my dailies since the last 1:1, my wins and my open actions from previous 1:1s.

---
%s
---

Propose 3-6 agenda topics: progress worth highlighting, decisions or help I need, feedback to ask for,
growth/career points. Each one short, specific and written in %s.
Reply with ONLY a JSON object, no prose, no code fences:
{"topics": [%s]}`, notes, o.Language, suggestionShape)

	text, err := Run(ctx, o, prompt, nil)
	if err != nil {
		return AgendaDraft{}, err
	}
	var a AgendaDraft
	return a, extractJSON(text, &a)
}
