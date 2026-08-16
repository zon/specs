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
description: Reviews prose: checks it against the guidelines, and fixes it.
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
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Read = %+v, want %+v", got, want)
	}
}

func TestReadReadsInlineTools(t *testing.T) {
	path := write(t, "---\nname: prose-editor\ntools: [read, edit]\n---\n")

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !reflect.DeepEqual(got.Tools, []string{"read", "edit"}) {
		t.Fatalf("tools = %v, want [read edit]", got.Tools)
	}
}

func TestReadYieldsZeroFieldsWithoutFrontmatter(t *testing.T) {
	path := write(t, "# Just a body\n")

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !reflect.DeepEqual(got, Fields{}) {
		t.Fatalf("Read = %+v, want zero fields", got)
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
