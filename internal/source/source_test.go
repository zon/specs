package source

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func namesOfKind(defs []Definition, kind Kind) []string {
	var names []string
	for _, d := range defs {
		if d.Kind == kind {
			names = append(names, d.Name)
		}
	}
	return names
}

func TestReadKindsReadsListedKindsInOrder(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	write(t, filepath.Join(dir, "agents", "code-architect.md"), "# code-architect\n")
	write(t, filepath.Join(dir, "docs", "zpecs", "architecture.md"), "# architecture\n")

	defs, err := ReadKinds([]Kind{Skill, Agent}, dir)
	if err != nil {
		t.Fatalf("ReadKinds: %v", err)
	}

	if got := namesOfKind(defs, Skill); !reflect.DeepEqual(got, []string{"prose-editor"}) {
		t.Fatalf("skills = %v, want [prose-editor]", got)
	}
	if got := namesOfKind(defs, Agent); !reflect.DeepEqual(got, []string{"code-architect"}) {
		t.Fatalf("agents = %v, want [code-architect]", got)
	}
	if got := namesOfKind(defs, Doc); len(got) != 0 {
		t.Fatalf("ReadKinds returned docs %v", got)
	}

	kinds := make([]Kind, len(defs))
	for i, d := range defs {
		kinds[i] = d.Kind
	}
	if want := []Kind{Skill, Agent}; !reflect.DeepEqual(kinds, want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
}

func TestReadKindsSelectsSingleKind(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	write(t, filepath.Join(dir, "agents", "code-architect.md"), "# code-architect\n")
	write(t, filepath.Join(dir, "docs", "zpecs", "architecture.md"), "# architecture\n")

	defs, err := ReadKinds([]Kind{Doc}, dir)
	if err != nil {
		t.Fatalf("ReadKinds: %v", err)
	}

	if len(defs) != 1 {
		t.Fatalf("ReadKinds returned %d defs, want 1", len(defs))
	}
	if got := namesOfKind(defs, Doc); !reflect.DeepEqual(got, []string{"architecture"}) {
		t.Fatalf("docs = %v, want [architecture]", got)
	}
}

func TestReadKindsErrorsOnMissingSource(t *testing.T) {
	if _, err := ReadKinds([]Kind{Skill}, filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected an error for a missing source directory")
	}
}

func TestReadKindsDoesNotReadMisplacedFiles(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "skills", "loose.md"), "# loose\n")
	write(t, filepath.Join(dir, "skills", "editor", "nested", "SKILL.md"), "# nested\n")
	write(t, filepath.Join(dir, "docs", "editor", "SKILL.md"), "# editor\n")
	write(t, filepath.Join(dir, "agents", "drafts", "note.md"), "# note\n")
	write(t, filepath.Join(dir, "docs", "specs", "foo.md"), "# foo\n")
	write(t, filepath.Join(dir, "skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	write(t, filepath.Join(dir, "agents", "code-architect.md"), "# code-architect\n")
	write(t, filepath.Join(dir, "docs", "zpecs", "architecture.md"), "# architecture\n")

	defs, err := ReadKinds([]Kind{Skill, Agent, Doc}, dir)
	if err != nil {
		t.Fatalf("ReadKinds: %v", err)
	}

	if got := namesOfKind(defs, Skill); !reflect.DeepEqual(got, []string{"prose-editor"}) {
		t.Fatalf("skills = %v, want [prose-editor]", got)
	}
	if got := namesOfKind(defs, Agent); !reflect.DeepEqual(got, []string{"code-architect"}) {
		t.Fatalf("agents = %v, want [code-architect]", got)
	}
	if got := namesOfKind(defs, Doc); !reflect.DeepEqual(got, []string{"architecture"}) {
		t.Fatalf("docs = %v, want [architecture]", got)
	}
}

func TestReadKindsReturnsSourcePaths(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	write(t, filepath.Join(dir, "agents", "code-architect.md"), "# code-architect\n")
	write(t, filepath.Join(dir, "docs", "zpecs", "architecture.md"), "# architecture\n")

	defs, err := ReadKinds([]Kind{Skill, Agent, Doc}, dir)
	if err != nil {
		t.Fatalf("ReadKinds: %v", err)
	}

	want := []Definition{
		{Kind: Skill, Name: "prose-editor", Path: filepath.Join(dir, "skills", "prose-editor", "SKILL.md")},
		{Kind: Agent, Name: "code-architect", Path: filepath.Join(dir, "agents", "code-architect.md")},
		{Kind: Doc, Name: "architecture", Path: filepath.Join(dir, "docs", "zpecs", "architecture.md")},
	}
	if !reflect.DeepEqual(defs, want) {
		t.Fatalf("defs = %v, want %v", defs, want)
	}
}

func TestReadKindsErrorsOnUnknownKind(t *testing.T) {
	if _, err := ReadKinds([]Kind{Kind(99)}, t.TempDir()); err == nil {
		t.Fatal("expected an error for an unknown kind")
	}
}
