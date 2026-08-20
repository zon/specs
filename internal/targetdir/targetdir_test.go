package targetdir

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/source"
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
	got := Path(root, source.Claude, skill("prose-editor"))
	require.Equal(t, want, got)
}

func TestPathClaudeAgent(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".claude", "agents", "prose-editor.md")
	got := Path(root, source.Claude, agent("prose-editor"))
	require.Equal(t, want, got)
}

func TestPathOpencodeSkill(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".opencode", "skills", "prose-editor", "SKILL.md")
	got := Path(root, source.Opencode, skill("prose-editor"))
	require.Equal(t, want, got)
}

func TestPathOpencodeAgent(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".opencode", "agents", "prose-editor.md")
	got := Path(root, source.Opencode, agent("prose-editor"))
	require.Equal(t, want, got)
}

func TestPathDocsDoc(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, "docs", "zpecs", "architecture.md")
	got := Path(root, source.Docs, doc("architecture"))
	require.Equal(t, want, got)
}

func TestRelPathComposesTargetDirAndSourceLayout(t *testing.T) {
	cases := []struct {
		name   string
		target string
		d      source.Definition
		want   string
	}{
		{name: "claude skill", target: source.Claude, d: skill("prose-editor"), want: filepath.Join(".claude", "skills", "prose-editor", "SKILL.md")},
		{name: "claude agent", target: source.Claude, d: agent("prose-editor"), want: filepath.Join(".claude", "agents", "prose-editor.md")},
		{name: "opencode skill", target: source.Opencode, d: skill("prose-editor"), want: filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")},
		{name: "opencode agent", target: source.Opencode, d: agent("prose-editor"), want: filepath.Join(".opencode", "agents", "prose-editor.md")},
		{name: "docs doc", target: source.Docs, d: doc("architecture"), want: filepath.Join("docs", "zpecs", "architecture.md")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, RelPath(tc.target, tc.d))
		})
	}
}

func TestWriteCreatesDirectoriesAndFile(t *testing.T) {
	root := t.TempDir()

	err := Write(root, source.Claude, agent("prose-editor"), "Review prose.\n", map[string]ownedPath{})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(root, ".claude", "agents", "prose-editor.md"))
	require.NoError(t, err)
	require.Equal(t, "Review prose.\n", string(content))
}

func TestWriteCreatesMissingDirectoriesForASkill(t *testing.T) {
	root := t.TempDir()

	err := Write(root, source.Claude, skill("prose-editor"), "# prose-editor\n", map[string]ownedPath{})
	require.NoError(t, err)

	for _, dir := range []string{
		filepath.Join(root, ".claude"),
		filepath.Join(root, ".claude", "skills"),
		filepath.Join(root, ".claude", "skills", "prose-editor"),
	} {
		info, statErr := os.Stat(dir)
		require.NoError(t, statErr)
		require.True(t, info.IsDir())
	}
}

func TestWriteCreatesMissingDirectoriesForAnAgent(t *testing.T) {
	root := t.TempDir()

	err := Write(root, source.Opencode, agent("code-architect"), "Architect code.\n", map[string]ownedPath{})
	require.NoError(t, err)

	for _, dir := range []string{
		filepath.Join(root, ".opencode"),
		filepath.Join(root, ".opencode", "agents"),
	} {
		info, statErr := os.Stat(dir)
		require.NoError(t, statErr)
		require.True(t, info.IsDir())
	}
}

func TestWriteDocsCreatesDirectoryAndFile(t *testing.T) {
	root := t.TempDir()

	err := Write(root, source.Docs, doc("architecture"), "# Architecture\n", map[string]ownedPath{})
	require.NoError(t, err)

	p := filepath.Join(root, "docs", "zpecs", "architecture.md")
	content, err := os.ReadFile(p)
	require.NoError(t, err)
	require.Equal(t, "# Architecture\n", string(content))
}

func TestSaveOwnedCreatesMissingTargetDirectory(t *testing.T) {
	root := t.TempDir()

	err := SaveOwned(root, source.Opencode, map[string]ownedPath{
		RelPath(source.Opencode, agent("prose-editor")): {kind: source.Agent, known: true},
	})
	require.NoError(t, err)

	dir := filepath.Join(root, ".opencode")
	info, err := os.Stat(dir)
	require.NoError(t, err)
	require.True(t, info.IsDir())
	_, err = os.Stat(filepath.Join(dir, manifestName))
	require.NoError(t, err)
}

func TestWriteWritesUnderRootNotWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	t.Chdir(t.TempDir())

	err := Write(root, source.Opencode, skill("prose-editor"), "# prose-editor\n", map[string]ownedPath{})
	require.NoError(t, err)

	path := filepath.Join(root, ".opencode", "skills", "prose-editor", "SKILL.md")
	_, err = os.Stat(path)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md"))
	require.Error(t, err)
}

func TestWriteLeavesForeignFileAlone(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, ".claude", "agents", "prose-editor.md")
	err := os.MkdirAll(filepath.Dir(p), 0o755)
	require.NoError(t, err)
	err = os.WriteFile(p, []byte("manual\n"), 0o644)
	require.NoError(t, err)

	owned := map[string]ownedPath{}
	err = Write(root, source.Claude, agent("prose-editor"), "rendered\n", owned)
	require.NoError(t, err)
	require.NotContains(t, owned, RelPath(source.Claude, agent("prose-editor")))

	content, err := os.ReadFile(p)
	require.NoError(t, err)
	require.Equal(t, "manual\n", string(content))
}

func TestWriteReplacesOwnedFile(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}

	err := Write(root, source.Claude, agent("prose-editor"), "first\n", owned)
	require.NoError(t, err)

	err = Write(root, source.Claude, agent("prose-editor"), "second\n", owned)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(root, ".claude", "agents", "prose-editor.md"))
	require.NoError(t, err)
	require.Equal(t, "second\n", string(content))
}

func TestOwnedEmptyWithoutManifest(t *testing.T) {
	root := t.TempDir()

	owned, err := Owned(root, source.Claude)
	require.NoError(t, err)
	require.Empty(t, owned)
}

func TestSaveOwnedPersistsWrittenPaths(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, source.Claude, agent("prose-editor"), "content\n", owned)
	require.NoError(t, err)
	err = Write(root, source.Claude, skill("prose-editor"), "# prose-editor\n", owned)
	require.NoError(t, err)
	err = SaveOwned(root, source.Claude, owned)
	require.NoError(t, err)

	got, err := Owned(root, source.Claude)
	require.NoError(t, err)
	require.Contains(t, got, RelPath(source.Claude, agent("prose-editor")))
	require.Contains(t, got, RelPath(source.Claude, skill("prose-editor")))
}

func TestManifestSeparatePerTarget(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, source.Claude, agent("prose-editor"), "content\n", owned)
	require.NoError(t, err)
	err = SaveOwned(root, source.Claude, owned)
	require.NoError(t, err)

	got, err := Owned(root, source.Opencode)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestManifestSeparateForDocs(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, source.Docs, doc("architecture"), "# Architecture\n", owned)
	require.NoError(t, err)
	err = SaveOwned(root, source.Docs, owned)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(root, "docs", "zpecs", manifestName))
	require.NoError(t, err)

	got, err := Owned(root, source.Opencode)
	require.NoError(t, err)
	require.Empty(t, got)

	got, err = Owned(root, source.Docs)
	require.NoError(t, err)
	require.Contains(t, got, RelPath(source.Docs, doc("architecture")))
}

func TestWriteRecordedInOwned(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}

	err := Write(root, source.Opencode, agent("prose-editor"), "content\n", owned)
	require.NoError(t, err)

	require.Contains(t, owned, RelPath(source.Opencode, agent("prose-editor")))
}

func TestWriteAllWritesSeveralDefinitions(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	defs := []source.Definition{
		skill("prose-editor"),
		agent("code-architect"),
		doc("architecture"),
	}

	err := WriteAll(root, source.Claude, defs, func(d source.Definition) (string, error) {
		return d.Name + "\n", nil
	}, owned)
	require.NoError(t, err)

	for _, d := range defs {
		content, err := os.ReadFile(Path(root, source.Claude, d))
		require.NoError(t, err)
		require.Equal(t, d.Name+"\n", string(content))
		require.Contains(t, owned, RelPath(source.Claude, d))
	}
}

func TestWriteAllSkipsForeignFile(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	defs := []source.Definition{
		agent("prose-editor"),
		skill("code-architect"),
	}

	foreign := Path(root, source.Opencode, agent("prose-editor"))
	err := os.MkdirAll(filepath.Dir(foreign), 0o755)
	require.NoError(t, err)
	err = os.WriteFile(foreign, []byte("manual\n"), 0o644)
	require.NoError(t, err)

	err = WriteAll(root, source.Opencode, defs, func(d source.Definition) (string, error) {
		return "rendered\n", nil
	}, owned)
	require.NoError(t, err)

	content, err := os.ReadFile(foreign)
	require.NoError(t, err)
	require.Equal(t, "manual\n", string(content))
	require.NotContains(t, owned, RelPath(source.Opencode, agent("prose-editor")))

	skillPath := Path(root, source.Opencode, skill("code-architect"))
	content, err = os.ReadFile(skillPath)
	require.NoError(t, err)
	require.Equal(t, "rendered\n", string(content))
	require.Contains(t, owned, RelPath(source.Opencode, skill("code-architect")))
}

func TestRemoveStaleRemovesFileNoLongerWritten(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, source.Claude, skill("prose-editor"), "# prose-editor\n", owned)
	require.NoError(t, err)

	removed, err := RemoveStale(root, source.Claude, owned, nil, source.Skill)
	require.NoError(t, err)
	require.Len(t, removed, 1)

	_, err = os.Stat(filepath.Join(root, ".claude", "skills", "prose-editor", "SKILL.md"))
	require.Error(t, err)
	require.Empty(t, owned)
}

func TestRemoveStaleDocsRemovesStaleDoc(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, source.Docs, doc("architecture"), "# Architecture\n", owned)
	require.NoError(t, err)

	removed, err := RemoveStale(root, source.Docs, owned, nil, source.Doc)
	require.NoError(t, err)
	require.Len(t, removed, 1)

	_, err = os.Stat(filepath.Join(root, "docs", "zpecs", "architecture.md"))
	require.Error(t, err)
	require.Empty(t, owned)
}

func TestRemoveStaleKeepsCurrentDefinition(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, source.Claude, agent("prose-editor"), "content\n", owned)
	require.NoError(t, err)
	current := []source.Definition{agent("prose-editor")}

	removed, err := RemoveStale(root, source.Claude, owned, current, source.Agent)
	require.NoError(t, err)
	require.Empty(t, removed)
	_, err = os.Stat(filepath.Join(root, ".claude", "agents", "prose-editor.md"))
	require.NoError(t, err)
}

func TestRemoveStaleScopedLeavesOtherKinds(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, source.Claude, skill("prose-editor"), "# prose-editor\n", owned)
	require.NoError(t, err)
	err = Write(root, source.Claude, agent("code-architect"), "content\n", owned)
	require.NoError(t, err)

	removed, err := RemoveStale(root, source.Claude, owned, nil, source.Skill)
	require.NoError(t, err)
	require.Len(t, removed, 1)

	_, err = os.Stat(filepath.Join(root, ".claude", "agents", "code-architect.md"))
	require.NoError(t, err)
}

func TestRemoveStaleAllKinds(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, source.Opencode, skill("prose-editor"), "# prose-editor\n", owned)
	require.NoError(t, err)
	err = Write(root, source.Opencode, agent("code-architect"), "content\n", owned)
	require.NoError(t, err)

	removed, err := RemoveStale(root, source.Opencode, owned, nil, source.Skill, source.Agent)
	require.NoError(t, err)
	require.Len(t, removed, 2)
	_, err = os.Stat(filepath.Join(root, ".opencode", "skills", "prose-editor", "SKILL.md"))
	require.Error(t, err)
	_, err = os.Stat(filepath.Join(root, ".opencode", "agents", "code-architect.md"))
	require.Error(t, err)
	require.Empty(t, owned)
}

func TestRemoveStaleHandlesMissingFile(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{
		RelPath(source.Claude, skill("prose-editor")): {kind: source.Skill, known: true},
	}

	removed, err := RemoveStale(root, source.Claude, owned, nil, source.Skill)
	require.NoError(t, err)
	require.Len(t, removed, 1)
	require.Empty(t, owned)
}

func TestRemoveStaleWrapsRemovalError(t *testing.T) {
	root := t.TempDir()
	rel := RelPath(source.Claude, skill("prose-editor"))
	owned := map[string]ownedPath{
		rel: {kind: source.Skill, known: true},
	}
	err := os.MkdirAll(filepath.Join(root, rel), 0o755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(root, rel, "file.txt"), []byte("x\n"), 0o644)
	require.NoError(t, err)

	_, err = RemoveStale(root, source.Claude, owned, nil, source.Skill)
	require.Error(t, err)
	require.ErrorContains(t, err, "removing stale definitions")
	var pathErr *fs.PathError
	require.ErrorAs(t, err, &pathErr)
}

func TestManifestRoundTripStoresKinds(t *testing.T) {
	root := t.TempDir()

	claude := map[string]ownedPath{}
	err := Write(root, source.Claude, agent("prose-editor"), "content\n", claude)
	require.NoError(t, err)
	err = Write(root, source.Claude, skill("prose-editor"), "# prose-editor\n", claude)
	require.NoError(t, err)
	err = SaveOwned(root, source.Claude, claude)
	require.NoError(t, err)

	docs := map[string]ownedPath{}
	err = Write(root, source.Docs, doc("architecture"), "# Architecture\n", docs)
	require.NoError(t, err)
	err = SaveOwned(root, source.Docs, docs)
	require.NoError(t, err)

	manifest, err := os.ReadFile(filepath.Join(root, ".claude", manifestName))
	require.NoError(t, err)
	wantClaude := "agent .claude/agents/prose-editor.md\nskill .claude/skills/prose-editor/SKILL.md\n"
	require.Equal(t, wantClaude, string(manifest))

	manifest, err = os.ReadFile(filepath.Join(root, "docs", "zpecs", manifestName))
	require.NoError(t, err)
	require.Equal(t, "doc docs/zpecs/architecture.md\n", string(manifest))

	owned, err := Owned(root, source.Claude)
	require.NoError(t, err)
	agentRel := RelPath(source.Claude, agent("prose-editor"))
	require.Equal(t, ownedPath{kind: source.Agent, known: true}, owned[agentRel])
	skillRel := RelPath(source.Claude, skill("prose-editor"))
	require.Equal(t, ownedPath{kind: source.Skill, known: true}, owned[skillRel])

	removed, err := RemoveStale(root, source.Claude, owned, nil, source.Skill)
	require.NoError(t, err)
	require.Equal(t, []string{skillRel}, removed)
	_, err = os.Stat(Path(root, source.Claude, agent("prose-editor")))
	require.NoError(t, err)

	removed, err = RemoveStale(root, source.Claude, owned, nil, source.Agent)
	require.NoError(t, err)
	require.Equal(t, []string{agentRel}, removed)

	owned, err = Owned(root, source.Docs)
	require.NoError(t, err)
	require.Equal(t, ownedPath{kind: source.Doc, known: true}, owned[RelPath(source.Docs, doc("architecture"))])
	removed, err = RemoveStale(root, source.Docs, owned, nil, source.Doc)
	require.NoError(t, err)
	require.Len(t, removed, 1)
}

func TestSaveOwnedSkipsManifestWhenNothingOwned(t *testing.T) {
	root := t.TempDir()

	err := SaveOwned(root, source.Docs, map[string]ownedPath{})
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(root, "docs", "zpecs", manifestName))
	require.Error(t, err)
}

func TestSaveOwnedRemovesEmptyManifest(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, source.Docs, doc("architecture"), "# Architecture\n", owned)
	require.NoError(t, err)
	err = SaveOwned(root, source.Docs, owned)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(root, "docs", "zpecs", manifestName))
	require.NoError(t, err)

	_, err = RemoveStale(root, source.Docs, owned, nil, source.Doc)
	require.NoError(t, err)
	err = SaveOwned(root, source.Docs, owned)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(root, "docs", "zpecs", manifestName))
	require.Error(t, err)
}

func TestOwnedReadsLegacyManifest(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".opencode")
	err := os.MkdirAll(dir, 0o755)
	require.NoError(t, err)
	rel := RelPath(source.Opencode, agent("prose-editor"))
	err = os.WriteFile(filepath.Join(dir, manifestName), []byte(rel+"\n"), 0o644)
	require.NoError(t, err)

	owned, err := Owned(root, source.Opencode)
	require.NoError(t, err)
	require.Contains(t, owned, rel)
	require.False(t, owned[rel].known)
}

func TestOwnedReadsLegacyPathWithSpace(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".opencode")
	err := os.MkdirAll(dir, 0o755)
	require.NoError(t, err)
	rel := ".opencode/agents/my agent.md"
	err = os.WriteFile(filepath.Join(dir, manifestName), []byte(rel+"\n"), 0o644)
	require.NoError(t, err)

	owned, err := Owned(root, source.Opencode)
	require.NoError(t, err)
	require.Contains(t, owned, rel)
	require.False(t, owned[rel].known)
}

func TestRemoveStaleSkipsUnknownKindEntry(t *testing.T) {
	root := t.TempDir()
	rel := RelPath(source.Claude, agent("prose-editor"))
	owned := map[string]ownedPath{rel: {}}
	err := os.MkdirAll(filepath.Join(root, filepath.Dir(rel)), 0o755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(root, rel), []byte("old\n"), 0o644)
	require.NoError(t, err)

	removed, err := RemoveStale(root, source.Claude, owned, nil, source.Agent)
	require.NoError(t, err)
	require.Empty(t, removed)
	_, err = os.Stat(filepath.Join(root, rel))
	require.NoError(t, err)
	require.Len(t, owned, 1)
}

func TestLegacyManifestUpgradedByNextRun(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".opencode")
	err := os.MkdirAll(dir, 0o755)
	require.NoError(t, err)
	rel := RelPath(source.Opencode, agent("prose-editor"))
	err = os.WriteFile(filepath.Join(dir, manifestName), []byte(rel+"\n"), 0o644)
	require.NoError(t, err)

	owned, err := Owned(root, source.Opencode)
	require.NoError(t, err)

	removed, err := RemoveStale(root, source.Opencode, owned, nil, source.Agent)
	require.NoError(t, err)
	require.Empty(t, removed)

	err = Write(root, source.Opencode, agent("prose-editor"), "content\n", owned)
	require.NoError(t, err)
	require.Equal(t, ownedPath{kind: source.Agent, known: true}, owned[rel])

	err = SaveOwned(root, source.Opencode, owned)
	require.NoError(t, err)
	manifest, err := os.ReadFile(filepath.Join(dir, manifestName))
	require.NoError(t, err)
	require.Equal(t, "agent "+rel+"\n", string(manifest))
}

func TestSaveOwnedKeepsUnknownEntryAsBarePath(t *testing.T) {
	root := t.TempDir()
	rel := RelPath(source.Claude, agent("prose-editor"))
	owned := map[string]ownedPath{rel: {}}

	err := SaveOwned(root, source.Claude, owned)
	require.NoError(t, err)

	manifest, err := os.ReadFile(filepath.Join(root, ".claude", manifestName))
	require.NoError(t, err)
	require.Equal(t, rel+"\n", string(manifest))
}

func TestKindWordsRoundTrip(t *testing.T) {
	for kind, word := range kindNames {
		got, ok := parseKind(word)
		require.True(t, ok)
		require.Equal(t, kind, got)
		require.Equal(t, word, kindName(kind))
	}
	_, ok := parseKind("not-a-kind")
	require.False(t, ok)
	require.Equal(t, "", kindName(source.Kind(99)))
}
