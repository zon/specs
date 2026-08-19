package frontmatter

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func write(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "definition.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadReadsNameDescriptionAndTools(t *testing.T) {
	path := write(t, `---
name: prose-editor
description: "Reviews prose: checks it against the guidelines, and fixes it."
tools:
  - read
  - edit
---

Body text.
`)
	want := Fields{
		Name:        "prose-editor",
		Description: "Reviews prose: checks it against the guidelines, and fixes it.",
		Tools:       []string{"read", "edit"},
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !reflect.DeepEqual(got.Fields, want) {
		t.Fatalf("Read = %+v, want %+v", got.Fields, want)
	}
}

func TestReadReadsInlineTools(t *testing.T) {
	path := write(t, "---\nname: prose-editor\ntools: [read, edit]\n---\n")

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !reflect.DeepEqual(got.Fields.Tools, []string{"read", "edit"}) {
		t.Fatalf("tools = %v, want [read edit]", got.Fields.Tools)
	}
}

func TestReadReadsMode(t *testing.T) {
	path := write(t, "---\nname: code-architect\nmode: primary\n---\n\nPlan the work.\n")

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Fields.Mode != "primary" {
		t.Fatalf("Mode = %q, want %q", got.Fields.Mode, "primary")
	}
}

func TestReadYieldsZeroFieldsWithoutFrontmatter(t *testing.T) {
	path := write(t, "# Just a body\n")

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !reflect.DeepEqual(got.Fields, Fields{}) {
		t.Fatalf("Read = %+v, want zero fields", got.Fields)
	}
}

func TestReadErrorsOnUnterminatedFrontmatter(t *testing.T) {
	path := write(t, "---\nname: prose-editor\n")

	if _, err := Read(path); err == nil {
		t.Fatal("expected an error for unterminated frontmatter")
	}
}

func TestReadErrorsOnMissingFile(t *testing.T) {
	if _, err := Read(filepath.Join(t.TempDir(), "missing.md")); err == nil {
		t.Fatal("expected an error for a missing file")
	}
}

func TestReadReadsTheBody(t *testing.T) {
	path := write(t, `---
name: prose-editor
description: Reviews prose.
---

You are a prose editor. Review the prose against the guidelines.
`)

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	want := "You are a prose editor. Review the prose against the guidelines."
	if got.Body != want {
		t.Fatalf("Body = %q, want %q", got.Body, want)
	}
}

func TestReadKeepsBodyLines(t *testing.T) {
	path := write(t, `---
name: prose-editor
---

You are a prose editor.

Give numbered instructions.
`)

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	want := "You are a prose editor.\n\nGive numbered instructions."
	if got.Body != want {
		t.Fatalf("Body = %q, want %q", got.Body, want)
	}
}

func TestReadYieldsWholeContentAsBodyWithoutFrontmatter(t *testing.T) {
	path := write(t, "You are a prose editor.\n")

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if want := "You are a prose editor."; got.Body != want {
		t.Fatalf("Body = %q, want %q", got.Body, want)
	}
}

func TestReadYieldsEmptyBodyWithoutOne(t *testing.T) {
	path := write(t, "---\nname: prose-editor\n---\n")

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Body != "" {
		t.Fatalf("Body = %q, want empty", got.Body)
	}
}
