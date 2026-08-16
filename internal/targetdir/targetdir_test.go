package targetdir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zon/specs/internal/source"
)

func skill(name string) source.Definition {
	return source.Definition{Kind: source.Skill, Name: name}
}

func agent(name string) source.Definition {
	return source.Definition{Kind: source.Agent, Name: name}
}

func TestPathClaudeSkill(t *testing.T) {
	want := filepath.Join(".claude", "skills", "prose-editor", "SKILL.md")
	if got := Path(Claude, skill("prose-editor")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestPathClaudeAgent(t *testing.T) {
	want := filepath.Join(".claude", "agents", "prose-editor.md")
	if got := Path(Claude, agent("prose-editor")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestPathOpencodeSkill(t *testing.T) {
	want := filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")
	if got := Path(Opencode, skill("prose-editor")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestPathOpencodeAgent(t *testing.T) {
	want := filepath.Join(".opencode", "agents", "prose-editor.md")
	if got := Path(Opencode, agent("prose-editor")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestWriteCreatesDirectoriesAndFile(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := Write(Claude, agent("prose-editor"), "Review prose.\n"); err != nil {
		t.Fatalf("Write: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(".claude", "agents", "prose-editor.md"))
	if err != nil {
		t.Fatalf("written file: %v", err)
	}
	if string(content) != "Review prose.\n" {
		t.Fatalf("content = %q, want %q", content, "Review prose.\n")
	}
}
