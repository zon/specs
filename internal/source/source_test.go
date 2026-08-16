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

func TestReadLocalFindsSkillAtSkillsNameSKILLMd(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	defs, err := ReadLocal(dir)
	if err != nil {
		t.Fatalf("ReadLocal: %v", err)
	}

	if got := namesOfKind(defs, Skill); !reflect.DeepEqual(got, []string{"prose-editor"}) {
		t.Fatalf("skills = %v, want [prose-editor]", got)
	}
}

func TestReadLocalDoesNotFindMisplacedSkill(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	write(t, filepath.Join(dir, "skills", "loose.md"), "# loose\n")
	write(t, filepath.Join(dir, "docs", "editor", "SKILL.md"), "# editor\n")
	write(t, filepath.Join(dir, "skills", "nested", "deep", "SKILL.md"), "# deep\n")

	defs, err := ReadLocal(dir)
	if err != nil {
		t.Fatalf("ReadLocal: %v", err)
	}

	if got := namesOfKind(defs, Skill); !reflect.DeepEqual(got, []string{"prose-editor"}) {
		t.Fatalf("skills = %v, want only [prose-editor]", got)
	}
}

func TestReadLocalFindsAgentAtAgentsNameMd(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "agents", "prose-editor.md"), "# prose-editor\n")

	defs, err := ReadLocal(dir)
	if err != nil {
		t.Fatalf("ReadLocal: %v", err)
	}

	if got := namesOfKind(defs, Agent); !reflect.DeepEqual(got, []string{"prose-editor"}) {
		t.Fatalf("agents = %v, want [prose-editor]", got)
	}
}

func TestReadLocalDoesNotFindMisplacedAgent(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "agents", "code-architect.md"), "# code-architect\n")
	write(t, filepath.Join(dir, "agents", "drafts", "note.md"), "# note\n")
	write(t, filepath.Join(dir, "notes", "orphan.md"), "# orphan\n")

	defs, err := ReadLocal(dir)
	if err != nil {
		t.Fatalf("ReadLocal: %v", err)
	}

	if got := namesOfKind(defs, Agent); !reflect.DeepEqual(got, []string{"code-architect"}) {
		t.Fatalf("agents = %v, want only [code-architect]", got)
	}
}

func TestReadLocalErrorsOnMissingSource(t *testing.T) {
	if _, err := ReadLocal(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected an error for a missing source directory")
	}
}

func TestReadSkillsFindsOnlySkills(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	write(t, filepath.Join(dir, "agents", "code-architect.md"), "# code-architect\n")

	defs, err := ReadSkills(dir)
	if err != nil {
		t.Fatalf("ReadSkills: %v", err)
	}

	if got := namesOfKind(defs, Skill); !reflect.DeepEqual(got, []string{"prose-editor"}) {
		t.Fatalf("skills = %v, want [prose-editor]", got)
	}
	if got := namesOfKind(defs, Agent); len(got) != 0 {
		t.Fatalf("ReadSkills returned agents %v", got)
	}
}

func TestReadAgentsFindsOnlyAgents(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	write(t, filepath.Join(dir, "agents", "code-architect.md"), "# code-architect\n")

	defs, err := ReadAgents(dir)
	if err != nil {
		t.Fatalf("ReadAgents: %v", err)
	}

	if got := namesOfKind(defs, Agent); !reflect.DeepEqual(got, []string{"code-architect"}) {
		t.Fatalf("agents = %v, want [code-architect]", got)
	}
	if got := namesOfKind(defs, Skill); len(got) != 0 {
		t.Fatalf("ReadAgents returned skills %v", got)
	}
}

func TestReadSkillsErrorsOnMissingSource(t *testing.T) {
	if _, err := ReadSkills(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected an error for a missing source directory")
	}
}

func TestReadAgentsErrorsOnMissingSource(t *testing.T) {
	if _, err := ReadAgents(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected an error for a missing source directory")
	}
}
