package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrecedenceEnvOverFileOverDefaults(t *testing.T) {
	home := t.TempDir()
	file := "# comment\nDJ_LINEAR_TEAM=ENG\nexport DJ_LANGUAGE=\"Spanish\"\nDJ_LINEAR_TOOLS=a, b\n"
	if err := os.WriteFile(filepath.Join(home, FileName), []byte(file), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DJ_LANGUAGE", "Catalan")
	t.Setenv("DJ_EDITOR", "")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "nvim")

	c, err := Load(home)
	if err != nil {
		t.Fatal(err)
	}
	if c.LinearTeam != "ENG" || c.Language != "Catalan" || c.LinearAssignee != "me" || c.Editor != "nvim" {
		t.Fatalf("got %+v", c)
	}
	if len(c.LinearTools) != 2 || c.LinearTools[1] != "b" {
		t.Fatalf("tools = %v", c.LinearTools)
	}
}
