package targetdir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/target"
)

func skill(name string) source.Definition {
	return source.Definition{Kind: source.Skill, Name: name}
}

func agent(name string) source.Definition {
	return source.Definition{Kind: source.Agent, Name: name}
}

func doc(name string) source.Definition {
	return source.Definition{Kind: source.Doc, Name: name}
}

func TestPathClaudeSkill(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".claude", "skills", "prose-editor", "SKILL.md")
	if got := Path(root, target.Claude, skill("prose-editor")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestPathClaudeAgent(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".claude", "agents", "prose-editor.md")
	if got := Path(root, target.Claude, agent("prose-editor")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestPathOpencodeSkill(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".opencode", "skills", "prose-editor", "SKILL.md")
	if got := Path(root, target.Opencode, skill("prose-editor")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestPathOpencodeAgent(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".opencode", "agents", "prose-editor.md")
	if got := Path(root, target.Opencode, agent("prose-editor")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestPathDocsDoc(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, "docs", "zpecs", "architecture.md")
	if got := Path(root, target.Docs, doc("architecture")); got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func TestWriteCreatesDirectoriesAndFile(t *testing.T) {
	root := t.TempDir()

	written, err := Write(root, target.Claude, agent("prose-editor"), "Review prose.\n", map[string]bool{})
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

func TestWriteCreatesMissingDirectoriesForASkill(t *testing.T) {
	root := t.TempDir()

	written, err := Write(root, target.Claude, skill("prose-editor"), "# prose-editor\n", map[string]bool{})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !written {
		t.Fatal("Write did not write a new file")
	}

	for _, dir := range []string{
		filepath.Join(root, ".claude"),
		filepath.Join(root, ".claude", "skills"),
		filepath.Join(root, ".claude", "skills", "prose-editor"),
	} {
		info, statErr := os.Stat(dir)
		if statErr != nil {
			t.Fatalf("directory %s not created: %v", dir, statErr)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", dir)
		}
	}
}

func TestWriteCreatesMissingDirectoriesForAnAgent(t *testing.T) {
	root := t.TempDir()

	written, err := Write(root, target.Opencode, agent("code-architect"), "Architect code.\n", map[string]bool{})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !written {
		t.Fatal("Write did not write a new file")
	}

	for _, dir := range []string{
		filepath.Join(root, ".opencode"),
		filepath.Join(root, ".opencode", "agents"),
	} {
		info, statErr := os.Stat(dir)
		if statErr != nil {
			t.Fatalf("directory %s not created: %v", dir, statErr)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", dir)
		}
	}
}

func TestWriteDocsCreatesDirectoryAndFile(t *testing.T) {
	root := t.TempDir()

	written, err := Write(root, target.Docs, doc("architecture"), "# Architecture\n", map[string]bool{})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !written {
		t.Fatal("Write did not write a new file")
	}

	p := filepath.Join(root, "docs", "zpecs", "architecture.md")
	content, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("written file: %v", err)
	}
	if string(content) != "# Architecture\n" {
		t.Fatalf("content = %q, want %q", content, "# Architecture\n")
	}
}

func TestSaveOwnedCreatesMissingTargetDirectory(t *testing.T) {
	root := t.TempDir()

	if err := SaveOwned(root, target.Opencode, map[string]bool{}); err != nil {
		t.Fatalf("SaveOwned: %v", err)
	}

	dir := filepath.Join(root, ".opencode")
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("target directory %s not created: %v", dir, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", dir)
	}
	if _, err := os.Stat(filepath.Join(dir, manifestName)); err != nil {
		t.Fatalf("manifest not written: %v", err)
	}
}

func TestWriteWritesUnderRootNotWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	t.Chdir(t.TempDir())

	if _, err := Write(root, target.Opencode, skill("prose-editor"), "# prose-editor\n", map[string]bool{}); err != nil {
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

	written, err := Write(root, target.Claude, agent("prose-editor"), "rendered\n", map[string]bool{})
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

	written, err := Write(root, target.Claude, agent("prose-editor"), "first\n", owned)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !written {
		t.Fatal("Write did not write the first file")
	}

	written, err = Write(root, target.Claude, agent("prose-editor"), "second\n", owned)
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

	owned, err := Owned(root, target.Claude)
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
	if _, err := Write(root, target.Claude, agent("prose-editor"), "content\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := Write(root, target.Claude, skill("prose-editor"), "# prose-editor\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := SaveOwned(root, target.Claude, owned); err != nil {
		t.Fatalf("SaveOwned: %v", err)
	}

	got, err := Owned(root, target.Claude)
	if err != nil {
		t.Fatalf("Owned: %v", err)
	}
	if !got[RelPath(target.Claude, agent("prose-editor"))] {
		t.Fatalf("Owned missing the agent path: %v", got)
	}
	if !got[RelPath(target.Claude, skill("prose-editor"))] {
		t.Fatalf("Owned missing the skill path: %v", got)
	}
}

func TestManifestSeparatePerTarget(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	if _, err := Write(root, target.Claude, agent("prose-editor"), "content\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := SaveOwned(root, target.Claude, owned); err != nil {
		t.Fatalf("SaveOwned: %v", err)
	}

	got, err := Owned(root, target.Opencode)
	if err != nil {
		t.Fatalf("Owned: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Opencode owns claude paths: %v", got)
	}
}

func TestManifestSeparateForDocs(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	if _, err := Write(root, target.Docs, doc("architecture"), "# Architecture\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := SaveOwned(root, target.Docs, owned); err != nil {
		t.Fatalf("SaveOwned: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "docs", "zpecs", manifestName)); err != nil {
		t.Fatalf("docs manifest not written: %v", err)
	}

	got, err := Owned(root, target.Opencode)
	if err != nil {
		t.Fatalf("Owned: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Opencode owns docs paths: %v", got)
	}

	got, err = Owned(root, target.Docs)
	if err != nil {
		t.Fatalf("Owned: %v", err)
	}
	if !got[RelPath(target.Docs, doc("architecture"))] {
		t.Fatalf("Docs missing the doc path: %v", got)
	}
}

func TestWriteRecordedInOwned(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}

	if _, err := Write(root, target.Opencode, agent("prose-editor"), "content\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if !owned[RelPath(target.Opencode, agent("prose-editor"))] {
		t.Fatalf("Write did not record %q as owned: %v", RelPath(target.Opencode, agent("prose-editor")), owned)
	}
}

func TestWriteAllWritesSeveralDefinitions(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	defs := []source.Definition{
		skill("prose-editor"),
		agent("code-architect"),
		doc("architecture"),
	}

	err := WriteAll(root, target.Claude, defs, func(d source.Definition) (string, error) {
		return d.Name + "\n", nil
	}, owned)
	if err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	for _, d := range defs {
		content, err := os.ReadFile(Path(root, target.Claude, d))
		if err != nil {
			t.Fatalf("written file for %s: %v", d.Name, err)
		}
		if string(content) != d.Name+"\n" {
			t.Fatalf("content = %q, want %q", content, d.Name+"\n")
		}
		if !owned[RelPath(target.Claude, d)] {
			t.Fatalf("WriteAll did not record %q as owned: %v", RelPath(target.Claude, d), owned)
		}
	}
}

func TestWriteAllSkipsForeignFile(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	defs := []source.Definition{
		agent("prose-editor"),
		skill("code-architect"),
	}

	foreign := Path(root, target.Opencode, agent("prose-editor"))
	if err := os.MkdirAll(filepath.Dir(foreign), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(foreign, []byte("manual\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := WriteAll(root, target.Opencode, defs, func(d source.Definition) (string, error) {
		return "rendered\n", nil
	}, owned)
	if err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	content, err := os.ReadFile(foreign)
	if err != nil {
		t.Fatalf("foreign file: %v", err)
	}
	if string(content) != "manual\n" {
		t.Fatalf("foreign file changed to %q", content)
	}
	if owned[RelPath(target.Opencode, agent("prose-editor"))] {
		t.Fatalf("WriteAll recorded a foreign file as owned: %v", owned)
	}

	skillPath := Path(root, target.Opencode, skill("code-architect"))
	content, err = os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("other definition not written: %v", err)
	}
	if string(content) != "rendered\n" {
		t.Fatalf("content = %q, want %q", content, "rendered\n")
	}
	if !owned[RelPath(target.Opencode, skill("code-architect"))] {
		t.Fatalf("WriteAll did not record the written skill as owned: %v", owned)
	}
}

func TestRemoveStaleRemovesFileNoLongerWritten(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	if _, err := Write(root, target.Claude, skill("prose-editor"), "# prose-editor\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}

	removed, err := RemoveStale(root, target.Claude, owned, nil, source.Skill)
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

func TestRemoveStaleDocsRemovesStaleDoc(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	if _, err := Write(root, target.Docs, doc("architecture"), "# Architecture\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}

	removed, err := RemoveStale(root, target.Docs, owned, nil, source.Doc)
	if err != nil {
		t.Fatalf("RemoveStale: %v", err)
	}
	if len(removed) != 1 {
		t.Fatalf("RemoveStale removed %v, want one path", removed)
	}

	if _, err := os.Stat(filepath.Join(root, "docs", "zpecs", "architecture.md")); err == nil {
		t.Fatal("stale doc still present")
	}
	if len(owned) != 0 {
		t.Fatalf("owned = %v, want none", owned)
	}
}

func TestRemoveStaleKeepsCurrentDefinition(t *testing.T) {
	root := t.TempDir()
	owned := map[string]bool{}
	if _, err := Write(root, target.Claude, agent("prose-editor"), "content\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	current := []source.Definition{agent("prose-editor")}

	removed, err := RemoveStale(root, target.Claude, owned, current, source.Agent)
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
	if _, err := Write(root, target.Claude, skill("prose-editor"), "# prose-editor\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := Write(root, target.Claude, agent("code-architect"), "content\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}

	removed, err := RemoveStale(root, target.Claude, owned, nil, source.Skill)
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
	if _, err := Write(root, target.Opencode, skill("prose-editor"), "# prose-editor\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := Write(root, target.Opencode, agent("code-architect"), "content\n", owned); err != nil {
		t.Fatalf("Write: %v", err)
	}

	removed, err := RemoveStale(root, target.Opencode, owned, nil, source.Skill, source.Agent)
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
	owned := map[string]bool{RelPath(target.Claude, skill("prose-editor")): true}

	removed, err := RemoveStale(root, target.Claude, owned, nil, source.Skill)
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
