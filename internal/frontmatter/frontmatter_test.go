package frontmatter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func write(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "definition.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
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
	require.NoError(t, err)
	require.Equal(t, want, got.Fields)
}

func TestReadReadsInlineTools(t *testing.T) {
	path := write(t, "---\nname: prose-editor\ntools: [read, edit]\n---\n")

	got, err := Read(path)
	require.NoError(t, err)
	require.Equal(t, []string{"read", "edit"}, got.Fields.Tools)
}

func TestReadReadsMode(t *testing.T) {
	path := write(t, "---\nname: code-architect\nmode: primary\n---\n\nPlan the work.\n")

	got, err := Read(path)
	require.NoError(t, err)
	require.Equal(t, "primary", got.Fields.Mode)
}

func TestReadYieldsZeroFieldsWithoutFrontmatter(t *testing.T) {
	path := write(t, "# Just a body\n")

	got, err := Read(path)
	require.NoError(t, err)
	require.Equal(t, Fields{}, got.Fields)
}

func TestReadErrorsOnUnterminatedFrontmatter(t *testing.T) {
	path := write(t, "---\nname: prose-editor\n")

	_, err := Read(path)
	require.Error(t, err)
}

func TestReadErrorsOnMissingFile(t *testing.T) {
	_, err := Read(filepath.Join(t.TempDir(), "missing.md"))
	require.Error(t, err)
}

func TestReadReadsTheBody(t *testing.T) {
	path := write(t, `---
name: prose-editor
description: Reviews prose.
---

You are a prose editor. Review the prose against the guidelines.
`)

	got, err := Read(path)
	require.NoError(t, err)
	want := "You are a prose editor. Review the prose against the guidelines."
	require.Equal(t, want, got.Body)
}

func TestReadKeepsBodyLines(t *testing.T) {
	path := write(t, `---
name: prose-editor
---

You are a prose editor.

Give numbered instructions.
`)

	got, err := Read(path)
	require.NoError(t, err)
	want := "You are a prose editor.\n\nGive numbered instructions."
	require.Equal(t, want, got.Body)
}

func TestReadYieldsWholeContentAsBodyWithoutFrontmatter(t *testing.T) {
	path := write(t, "You are a prose editor.\n")

	got, err := Read(path)
	require.NoError(t, err)
	require.Equal(t, "You are a prose editor.", got.Body)
}

func TestReadYieldsEmptyBodyWithoutOne(t *testing.T) {
	path := write(t, "---\nname: prose-editor\n---\n")

	got, err := Read(path)
	require.NoError(t, err)
	require.Equal(t, "", got.Body)
}

func TestReadYieldsZeroFieldsForEmptyFile(t *testing.T) {
	path := write(t, "")

	got, err := Read(path)
	require.NoError(t, err)
	require.Equal(t, Fields{}, got.Fields)
	require.Equal(t, "", got.Body)
}
