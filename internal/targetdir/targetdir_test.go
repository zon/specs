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

	written, err := Write(root, Claude, agent("prose-editor"), "Review prose.\n", map[string]bool{})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !written {
		t.Fatal("Write did not write a new file")
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

	if _, err := Write(root, Opencode, skill("prose-editor"), "# prose-editor\n", map[string]bool{}); err != nil {
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

func TestWriteLeavesForeignFileAlone(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, ".claude", "agents", "prose-editor.md")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("manual\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	written, err := Write(root, Claude, agent("prose-editor"), "rendered\n", map[string]bool{})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if written {
		t.Fatal("Write replaced a foreign file")
	}

	content, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("foreign file: %v", err)
	}
	if string(content) != "manual\n" {
		t.Fatalf("foreign file changed to %q", content)
	}
}

func TestWriteReplacesOwnedFile(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}

	written, err := Write(root, Claude, agent("prose-editor"), "first\n", owned)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !written {
		t.Fatal("Write did not write the first file")
	}

	written, err = Write(root, Claude, agent("prose-editor"), "second\n", owned)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !written {
		t.Fatal("Write did not replace an owned file")
	}

	content, err := os.ReadFile(filepath.Join(root, ".claude", "agents", "prose-editor.md"))
	if err != nil {
		t.Fatalf("owned file: %v", err)
	}
	if string(content) != "second\n" {
		t.Fatalf("content = %q, want %q", content, "second\n")
	}
}

func TestOwnedEmptyWithoutManifest(t *testing.T) {
	root := t.TempDir()

	owned, err := Owned(root, Claude)
	if err != nil {
		t.Fatalf("Owned: %v", err)
	}
	if len(owned) != 0 {
		t.Fatalf("Owned = %v, want none", owned)
	}
}

func TestSaveOwnedPersistsWrittenPaths(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	if _, err := Write(root, Claude, agent("prose-editor"), "content\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := Write(root, Claude, skill("prose-editor"), "# prose-editor\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := SaveOwned(root, Claude, owned); err != nil {
		t.Fatalf("SaveOwned: %v", err)
	}

	got, err := Owned(root, Claude)
	if err != nil {
		t.Fatalf("Owned: %v", err)
	}
	if !got[RelPath(Claude, agent("prose-editor"))] {
		t.Fatalf("Owned missing the agent path: %v", got)
	}
	if !got[RelPath(Claude, skill("prose-editor"))] {
		t.Fatalf("Owned missing the skill path: %v", got)
	}
}

func TestManifestSeparatePerTarget(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	if _, err := Write(root, Claude, agent("prose-editor"), "content\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := SaveOwned(root, Claude, owned); err != nil {
		t.Fatalf("SaveOwned: %v", err)
	}

	got, err := Owned(root, Opencode)
	if err != nil {
		t.Fatalf("Owned: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Opencode owns claude paths: %v", got)
	}
}

func TestWriteRecordedInOwned(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}

	if _, err := Write(root, Opencode, agent("prose-editor"), "content\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if !owned[RelPath(Opencode, agent("prose-editor"))] {
		t.Fatalf("Write did not record %q as owned: %v", RelPath(Opencode, agent("prose-editor")), owned)
	}
}

func TestRemoveStaleRemovesFileNoLongerWritten(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	if _, err := Write(root, Claude, skill("prose-editor"), "# prose-editor\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}

	removed, err := RemoveStale(root, Claude, owned, nil, source.Skill)
	if err != nil {
		t.Fatalf("RemoveStale: %v", err)
	}
	if len(removed) != 1 {
		t.Fatalf("RemoveStale removed %v, want one path", removed)
	}

	if _, err := os.Stat(filepath.Join(root, ".claude", "skills", "prose-editor", "SKILL.md")); err == nil {
		t.Fatal("stale skill still present")
	}
	if len(owned) != 0 {
		t.Fatalf("owned = %v, want none", owned)
	}
}

func TestRemoveStaleKeepsCurrentDefinition(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	if _, err := Write(root, Claude, agent("prose-editor"), "content\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	current := []source.Definition{agent("prose-editor")}

	removed, err := RemoveStale(root, Claude, owned, current, source.Agent)
	if err != nil {
		t.Fatalf("RemoveStale: %v", err)
	}
	if len(removed) != 0 {
		t.Fatalf("RemoveStale removed %v, want none", removed)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude", "agents", "prose-editor.md")); err != nil {
		t.Fatalf("current agent removed: %v", err)
	}
}

func TestRemoveStaleScopedLeavesOtherKinds(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	if _, err := Write(root, Claude, skill("prose-editor"), "# prose-editor\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := Write(root, Claude, agent("code-architect"), "content\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}

	removed, err := RemoveStale(root, Claude, owned, nil, source.Skill)
	if err != nil {
		t.Fatalf("RemoveStale: %v", err)
	}
	if len(removed) != 1 {
		t.Fatalf("RemoveStale removed %v, want the skill only", removed)
	}

	if _, err := os.Stat(filepath.Join(root, ".claude", "agents", "code-architect.md")); err != nil {
		t.Fatalf("agent removed by a skills-only run: %v", err)
	}
}

func TestRemoveStaleAllKinds(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	if _, err := Write(root, Opencode, skill("prose-editor"), "# prose-editor\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := Write(root, Opencode, agent("code-architect"), "content\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}

	removed, err := RemoveStale(root, Opencode, owned, nil, source.Skill, source.Agent)
	if err != nil {
		t.Fatalf("RemoveStale: %v", err)
	}
	if len(removed) != 2 {
		t.Fatalf("RemoveStale removed %v, want both paths", removed)
	}
	if _, err := os.Stat(filepath.Join(root, ".opencode", "skills", "prose-editor", "SKILL.md")); err == nil {
		t.Fatal("stale skill still present")
	}
	if _, err := os.Stat(filepath.Join(root, ".opencode", "agents", "code-architect.md")); err == nil {
		t.Fatal("stale agent still present")
	}
	if len(owned) != 0 {
		t.Fatalf("owned = %v, want none", owned)
	}
}

func TestRemoveStaleHandlesMissingFile(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{RelPath(Claude, skill("prose-editor")): true}

	removed, err := RemoveStale(root, Claude, owned, nil, source.Skill)
	if err != nil {
		t.Fatalf("RemoveStale: %v", err)
	}
	if len(removed) != 1 {
		t.Fatalf("RemoveStale removed %v, want one path", removed)
	}
	if len(owned) != 0 {
		t.Fatalf("owned = %v, want none", owned)
	}
}
