package targetdir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
	got := Path(root, target.Claude, skill("prose-editor"))
	require.Equal(t, want, got)
}

func TestPathClaudeAgent(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".claude", "agents", "prose-editor.md")
	got := Path(root, target.Claude, agent("prose-editor"))
	require.Equal(t, want, got)
}

func TestPathOpencodeSkill(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".opencode", "skills", "prose-editor", "SKILL.md")
	got := Path(root, target.Opencode, skill("prose-editor"))
	require.Equal(t, want, got)
}

func TestPathOpencodeAgent(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".opencode", "agents", "prose-editor.md")
	got := Path(root, target.Opencode, agent("prose-editor"))
	require.Equal(t, want, got)
}

func TestPathDocsDoc(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, "docs", "zpecs", "architecture.md")
	got := Path(root, target.Docs, doc("architecture"))
	require.Equal(t, want, got)
}

func TestWriteCreatesDirectoriesAndFile(t *testing.T) {
	root := t.TempDir()

	err := Write(root, target.Claude, agent("prose-editor"), "Review prose.\n", map[string]ownedPath{})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(root, ".claude", "agents", "prose-editor.md"))
	require.NoError(t, err)
	require.Equal(t, "Review prose.\n", string(content))
}

func TestWriteCreatesMissingDirectoriesForASkill(t *testing.T) {
	root := t.TempDir()

	err := Write(root, target.Claude, skill("prose-editor"), "# prose-editor\n", map[string]ownedPath{})
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

	err := Write(root, target.Opencode, agent("code-architect"), "Architect code.\n", map[string]ownedPath{})
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

	err := Write(root, target.Docs, doc("architecture"), "# Architecture\n", map[string]ownedPath{})
	require.NoError(t, err)

	p := filepath.Join(root, "docs", "zpecs", "architecture.md")
	content, err := os.ReadFile(p)
	require.NoError(t, err)
	require.Equal(t, "# Architecture\n", string(content))
}

func TestSaveOwnedCreatesMissingTargetDirectory(t *testing.T) {
	root := t.TempDir()

	err := SaveOwned(root, target.Opencode, map[string]ownedPath{})
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

	err := Write(root, target.Opencode, skill("prose-editor"), "# prose-editor\n", map[string]ownedPath{})
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
	err = Write(root, target.Claude, agent("prose-editor"), "rendered\n", owned)
	require.NoError(t, err)
	require.NotContains(t, owned, RelPath(target.Claude, agent("prose-editor")))

	content, err := os.ReadFile(p)
	require.NoError(t, err)
	require.Equal(t, "manual\n", string(content))
}

func TestWriteReplacesOwnedFile(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}

	err := Write(root, target.Claude, agent("prose-editor"), "first\n", owned)
	require.NoError(t, err)

	err = Write(root, target.Claude, agent("prose-editor"), "second\n", owned)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(root, ".claude", "agents", "prose-editor.md"))
	require.NoError(t, err)
	require.Equal(t, "second\n", string(content))
}

func TestOwnedEmptyWithoutManifest(t *testing.T) {
	root := t.TempDir()

	owned, err := Owned(root, target.Claude)
	require.NoError(t, err)
	require.Empty(t, owned)
}

func TestSaveOwnedPersistsWrittenPaths(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, target.Claude, agent("prose-editor"), "content\n", owned)
	require.NoError(t, err)
	err = Write(root, target.Claude, skill("prose-editor"), "# prose-editor\n", owned)
	require.NoError(t, err)
	err = SaveOwned(root, target.Claude, owned)
	require.NoError(t, err)

	got, err := Owned(root, target.Claude)
	require.NoError(t, err)
	require.Contains(t, got, RelPath(target.Claude, agent("prose-editor")))
	require.Contains(t, got, RelPath(target.Claude, skill("prose-editor")))
}

func TestManifestSeparatePerTarget(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, target.Claude, agent("prose-editor"), "content\n", owned)
	require.NoError(t, err)
	err = SaveOwned(root, target.Claude, owned)
	require.NoError(t, err)

	got, err := Owned(root, target.Opencode)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestManifestSeparateForDocs(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, target.Docs, doc("architecture"), "# Architecture\n", owned)
	require.NoError(t, err)
	err = SaveOwned(root, target.Docs, owned)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(root, "docs", "zpecs", manifestName))
	require.NoError(t, err)

	got, err := Owned(root, target.Opencode)
	require.NoError(t, err)
	require.Empty(t, got)

	got, err = Owned(root, target.Docs)
	require.NoError(t, err)
	require.Contains(t, got, RelPath(target.Docs, doc("architecture")))
}

func TestWriteRecordedInOwned(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}

	err := Write(root, target.Opencode, agent("prose-editor"), "content\n", owned)
	require.NoError(t, err)

	require.Contains(t, owned, RelPath(target.Opencode, agent("prose-editor")))
}

func TestWriteAllWritesSeveralDefinitions(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	defs := []source.Definition{
		skill("prose-editor"),
		agent("code-architect"),
		doc("architecture"),
	}

	err := WriteAll(root, target.Claude, defs, func(d source.Definition) (string, error) {
		return d.Name + "\n", nil
	}, owned)
	require.NoError(t, err)

	for _, d := range defs {
		content, err := os.ReadFile(Path(root, target.Claude, d))
		require.NoError(t, err)
		require.Equal(t, d.Name+"\n", string(content))
		require.Contains(t, owned, RelPath(target.Claude, d))
	}
}

func TestWriteAllSkipsForeignFile(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	defs := []source.Definition{
		agent("prose-editor"),
		skill("code-architect"),
	}

	foreign := Path(root, target.Opencode, agent("prose-editor"))
	err := os.MkdirAll(filepath.Dir(foreign), 0o755)
	require.NoError(t, err)
	err = os.WriteFile(foreign, []byte("manual\n"), 0o644)
	require.NoError(t, err)

	err = WriteAll(root, target.Opencode, defs, func(d source.Definition) (string, error) {
		return "rendered\n", nil
	}, owned)
	require.NoError(t, err)

	content, err := os.ReadFile(foreign)
	require.NoError(t, err)
	require.Equal(t, "manual\n", string(content))
	require.NotContains(t, owned, RelPath(target.Opencode, agent("prose-editor")))

	skillPath := Path(root, target.Opencode, skill("code-architect"))
	content, err = os.ReadFile(skillPath)
	require.NoError(t, err)
	require.Equal(t, "rendered\n", string(content))
	require.Contains(t, owned, RelPath(target.Opencode, skill("code-architect")))
}

func TestRemoveStaleRemovesFileNoLongerWritten(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, target.Claude, skill("prose-editor"), "# prose-editor\n", owned)
	require.NoError(t, err)

	removed, err := RemoveStale(root, target.Claude, owned, nil, source.Skill)
	require.NoError(t, err)
	require.Len(t, removed, 1)

	_, err = os.Stat(filepath.Join(root, ".claude", "skills", "prose-editor", "SKILL.md"))
	require.Error(t, err)
	require.Empty(t, owned)
}

func TestRemoveStaleDocsRemovesStaleDoc(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, target.Docs, doc("architecture"), "# Architecture\n", owned)
	require.NoError(t, err)

	removed, err := RemoveStale(root, target.Docs, owned, nil, source.Doc)
	require.NoError(t, err)
	require.Len(t, removed, 1)

	_, err = os.Stat(filepath.Join(root, "docs", "zpecs", "architecture.md"))
	require.Error(t, err)
	require.Empty(t, owned)
}

func TestRemoveStaleKeepsCurrentDefinition(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, target.Claude, agent("prose-editor"), "content\n", owned)
	require.NoError(t, err)
	current := []source.Definition{agent("prose-editor")}

	removed, err := RemoveStale(root, target.Claude, owned, current, source.Agent)
	require.NoError(t, err)
	require.Empty(t, removed)
	_, err = os.Stat(filepath.Join(root, ".claude", "agents", "prose-editor.md"))
	require.NoError(t, err)
}

func TestRemoveStaleScopedLeavesOtherKinds(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, target.Claude, skill("prose-editor"), "# prose-editor\n", owned)
	require.NoError(t, err)
	err = Write(root, target.Claude, agent("code-architect"), "content\n", owned)
	require.NoError(t, err)

	removed, err := RemoveStale(root, target.Claude, owned, nil, source.Skill)
	require.NoError(t, err)
	require.Len(t, removed, 1)

	_, err = os.Stat(filepath.Join(root, ".claude", "agents", "code-architect.md"))
	require.NoError(t, err)
}

func TestRemoveStaleAllKinds(t *testing.T) {
	root := t.TempDir()
	owned := map[string]ownedPath{}
	err := Write(root, target.Opencode, skill("prose-editor"), "# prose-editor\n", owned)
	require.NoError(t, err)
	err = Write(root, target.Opencode, agent("code-architect"), "content\n", owned)
	require.NoError(t, err)

	removed, err := RemoveStale(root, target.Opencode, owned, nil, source.Skill, source.Agent)
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
		RelPath(target.Claude, skill("prose-editor")): {kind: source.Skill, known: true},
	}

	removed, err := RemoveStale(root, target.Claude, owned, nil, source.Skill)
	require.NoError(t, err)
	require.Len(t, removed, 1)
	require.Empty(t, owned)
}

func TestManifestRoundTripStoresKinds(t *testing.T) {
	root := t.TempDir()

	claude := map[string]ownedPath{}
	err := Write(root, target.Claude, agent("prose-editor"), "content\n", claude)
	require.NoError(t, err)
	err = Write(root, target.Claude, skill("prose-editor"), "# prose-editor\n", claude)
	require.NoError(t, err)
	err = SaveOwned(root, target.Claude, claude)
	require.NoError(t, err)

	docs := map[string]ownedPath{}
	err = Write(root, target.Docs, doc("architecture"), "# Architecture\n", docs)
	require.NoError(t, err)
	err = SaveOwned(root, target.Docs, docs)
	require.NoError(t, err)

	manifest, err := os.ReadFile(filepath.Join(root, ".claude", manifestName))
	require.NoError(t, err)
	wantClaude := "agent .claude/agents/prose-editor.md\nskill .claude/skills/prose-editor/SKILL.md\n"
	require.Equal(t, wantClaude, string(manifest))

	manifest, err = os.ReadFile(filepath.Join(root, "docs", "zpecs", manifestName))
	require.NoError(t, err)
	require.Equal(t, "doc docs/zpecs/architecture.md\n", string(manifest))

	owned, err := Owned(root, target.Claude)
	require.NoError(t, err)
	agentRel := RelPath(target.Claude, agent("prose-editor"))
	require.Equal(t, ownedPath{kind: source.Agent, known: true}, owned[agentRel])
	skillRel := RelPath(target.Claude, skill("prose-editor"))
	require.Equal(t, ownedPath{kind: source.Skill, known: true}, owned[skillRel])

	removed, err := RemoveStale(root, target.Claude, owned, nil, source.Skill)
	require.NoError(t, err)
	require.Equal(t, []string{skillRel}, removed)
	_, err = os.Stat(Path(root, target.Claude, agent("prose-editor")))
	require.NoError(t, err)

	removed, err = RemoveStale(root, target.Claude, owned, nil, source.Agent)
	require.NoError(t, err)
	require.Equal(t, []string{agentRel}, removed)

	owned, err = Owned(root, target.Docs)
	require.NoError(t, err)
	require.Equal(t, ownedPath{kind: source.Doc, known: true}, owned[RelPath(target.Docs, doc("architecture"))])
	removed, err = RemoveStale(root, target.Docs, owned, nil, source.Doc)
	require.NoError(t, err)
	require.Len(t, removed, 1)
}

func TestOwnedReadsLegacyManifest(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".opencode")
	err := os.MkdirAll(dir, 0o755)
	require.NoError(t, err)
	rel := RelPath(target.Opencode, agent("prose-editor"))
	err = os.WriteFile(filepath.Join(dir, manifestName), []byte(rel+"\n"), 0o644)
	require.NoError(t, err)

	owned, err := Owned(root, target.Opencode)
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

	owned, err := Owned(root, target.Opencode)
	require.NoError(t, err)
	require.Contains(t, owned, rel)
	require.False(t, owned[rel].known)
}

func TestRemoveStaleSkipsUnknownKindEntry(t *testing.T) {
	root := t.TempDir()
	rel := RelPath(target.Claude, agent("prose-editor"))
	owned := map[string]ownedPath{rel: {}}
	err := os.MkdirAll(filepath.Join(root, filepath.Dir(rel)), 0o755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(root, rel), []byte("old\n"), 0o644)
	require.NoError(t, err)

	removed, err := RemoveStale(root, target.Claude, owned, nil, source.Agent)
	require.NoError(t, err)
	require.Empty(t, removed)
	_, err = os.Stat(filepath.Join(root, rel))
	require.NoError(t, err)
	require.Len(t, owned, 1)
}

func TestRemoveStaleWrapsRemoveErrorWithPath(t *testing.T) {
	root := t.TempDir()
	rel := RelPath(target.Claude, skill("prose-editor"))
	err := os.MkdirAll(filepath.Join(root, rel, "subdir"), 0o755)
	require.NoError(t, err)
	owned := map[string]ownedPath{
		rel: {kind: source.Skill, known: true},
	}

	removed, err := RemoveStale(root, target.Claude, owned, nil, source.Skill)
	require.Error(t, err)
	require.ErrorContains(t, err, "removing stale "+filepath.Join(root, rel))
	require.Nil(t, removed)
}

func TestLegacyManifestUpgradedByNextRun(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".opencode")
	err := os.MkdirAll(dir, 0o755)
	require.NoError(t, err)
	rel := RelPath(target.Opencode, agent("prose-editor"))
	err = os.WriteFile(filepath.Join(dir, manifestName), []byte(rel+"\n"), 0o644)
	require.NoError(t, err)

	owned, err := Owned(root, target.Opencode)
	require.NoError(t, err)

	removed, err := RemoveStale(root, target.Opencode, owned, nil, source.Agent)
	require.NoError(t, err)
	require.Empty(t, removed)

	err = Write(root, target.Opencode, agent("prose-editor"), "content\n", owned)
	require.NoError(t, err)
	require.Equal(t, ownedPath{kind: source.Agent, known: true}, owned[rel])

	err = SaveOwned(root, target.Opencode, owned)
	require.NoError(t, err)
	manifest, err := os.ReadFile(filepath.Join(dir, manifestName))
	require.NoError(t, err)
	require.Equal(t, "agent "+rel+"\n", string(manifest))
}

func TestSaveOwnedKeepsUnknownEntryAsBarePath(t *testing.T) {
	root := t.TempDir()
	rel := RelPath(target.Claude, agent("prose-editor"))
	owned := map[string]ownedPath{rel: {}}

	err := SaveOwned(root, target.Claude, owned)
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
