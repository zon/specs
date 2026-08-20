package source

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
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
	require.NoError(t, err)

	require.Equal(t, []string{"prose-editor"}, namesOfKind(defs, Skill))
	require.Equal(t, []string{"code-architect"}, namesOfKind(defs, Agent))
	require.Empty(t, namesOfKind(defs, Doc))

	kinds := make([]Kind, len(defs))
	for i, d := range defs {
		kinds[i] = d.Kind
	}
	require.Equal(t, []Kind{Skill, Agent}, kinds)
}

func TestReadKindsSelectsSingleKind(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	write(t, filepath.Join(dir, "agents", "code-architect.md"), "# code-architect\n")
	write(t, filepath.Join(dir, "docs", "zpecs", "architecture.md"), "# architecture\n")

	defs, err := ReadKinds([]Kind{Doc}, dir)
	require.NoError(t, err)

	require.Equal(t, 1, len(defs))
	require.Equal(t, []string{"architecture"}, namesOfKind(defs, Doc))
}

func TestReadKindsErrorsOnMissingSource(t *testing.T) {
	_, err := ReadKinds([]Kind{Skill}, filepath.Join(t.TempDir(), "missing"))
	require.Error(t, err)
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
	require.NoError(t, err)

	require.Equal(t, []string{"prose-editor"}, namesOfKind(defs, Skill))
	require.Equal(t, []string{"code-architect"}, namesOfKind(defs, Agent))
	require.Equal(t, []string{"architecture"}, namesOfKind(defs, Doc))
}

func TestReadKindsReturnsSourcePaths(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	write(t, filepath.Join(dir, "agents", "code-architect.md"), "# code-architect\n")
	write(t, filepath.Join(dir, "docs", "zpecs", "architecture.md"), "# architecture\n")

	defs, err := ReadKinds([]Kind{Skill, Agent, Doc}, dir)
	require.NoError(t, err)

	want := []Definition{
		{Kind: Skill, Name: "prose-editor", Path: filepath.Join(dir, "skills", "prose-editor", "SKILL.md")},
		{Kind: Agent, Name: "code-architect", Path: filepath.Join(dir, "agents", "code-architect.md")},
		{Kind: Doc, Name: "architecture", Path: filepath.Join(dir, "docs", "zpecs", "architecture.md")},
	}
	require.Equal(t, want, defs)
}

func TestReadKindsErrorsOnUnknownKind(t *testing.T) {
	_, err := ReadKinds([]Kind{Kind(99)}, t.TempDir())
	require.Error(t, err)
}

func TestUnmarshalScope(t *testing.T) {
	cases := []struct {
		name    string
		s       string
		want    Scope
		wantErr bool
	}{
		{name: "all", s: "all", want: ScopeAll},
		{name: "skills", s: "skills", want: ScopeSkills},
		{name: "agents", s: "agents", want: ScopeAgents},
		{name: "docs", s: "docs", want: ScopeDocs},
		{name: "unknown scope", s: "vscode", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got Scope
			err := got.UnmarshalText([]byte(tc.s))
			require.Equal(t, tc.wantErr, err != nil)
			if err == nil {
				require.Equal(t, tc.want, got)
			}
		})
	}
}
