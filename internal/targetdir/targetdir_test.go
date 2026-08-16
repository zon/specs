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
	root := t.TempDir()
	want := filepath.Join(root, ".claude", "skills", "prose-editor", "SKILL.md")
	if got := Path(root, Claude, skill("prose-editor")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestPathClaudeAgent(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".claude", "agents", "prose-editor.md")
	if got := Path(root, Claude, agent("prose-editor")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestPathOpencodeSkill(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".opencode", "skills", "prose-editor", "SKILL.md")
	if got := Path(root, Opencode, skill("prose-editor")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestPathOpencodeAgent(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".opencode", "agents", "prose-editor.md")
	if got := Path(root, Opencode, agent("prose-editor")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestWriteCreatesDirectoriesAndFile(t *testing.T) {
	root := t.TempDir()

	if err := Write(root, Claude, agent("prose-editor"), "Review prose.\n"); err != nil {
		t.Fatalf("Write: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(root, ".claude", "agents", "prose-editor.md"))
	if err != nil {
		t.Fatalf("written file: %v", err)
	}
	if string(content) != "Review prose.\n" {
		t.Fatalf("content = %q, want %q", content, "Review prose.\n")
	}
}

func TestWriteWritesUnderRootNotWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	t.Chdir(t.TempDir())

	if err := Write(root, Opencode, skill("prose-editor"), "# prose-editor\n"); err != nil {
		t.Fatalf("Write: %v", err)
	}

	path := filepath.Join(root, ".opencode", "skills", "prose-editor", "SKILL.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not written under root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")); err == nil {
		t.Fatal("file written in the working directory")
	}
}
