package update

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/target"
	"github.com/zon/specs/internal/testutil"
)

func TestResolveSourceClonesRemote(t *testing.T) {
	src := testutil.GitRepoURL(t, map[string]string{"seed": "content\n"})

	dir, label, cleanup, err := resolveSource(src)
	require.NoError(t, err)
	defer cleanup()
	require.Equal(t, src, label)
	require.DirExists(t, dir)
	testutil.RequireFile(t, dir, "seed")
	cleanup()
	require.NoFileExists(t, dir)
}

func TestResolveSourceReadsLocalInPlace(t *testing.T) {
	dir := t.TempDir()

	gotDir, label, cleanup, err := resolveSource(dir)
	require.NoError(t, err)
	require.Equal(t, dir, gotDir)
	require.Equal(t, dir, label)
	cleanup()
	require.DirExists(t, dir)
}

func TestUpdatePairReportsTheRun(t *testing.T) {
	root := t.TempDir()
	src := testutil.SkillSource(t, "prose-editor")
	reported := testutil.CaptureReport(t)

	err := updatePair(root, src, src, pair{target: target.Opencode, kinds: []source.Kind{source.Skill}})
	require.NoError(t, err)
	require.Contains(t, reported(), src)
}

func TestPairsSelectsRunsPerScope(t *testing.T) {
	const targetName = target.Opencode
	cases := []struct {
		name  string
		scope source.Scope
		want  []pair
	}{
		{name: "skills", scope: source.ScopeSkills, want: []pair{{target: targetName, kinds: []source.Kind{source.Skill}}}},
		{name: "agents", scope: source.ScopeAgents, want: []pair{{target: targetName, kinds: []source.Kind{source.Agent}}}},
		{name: "docs", scope: source.ScopeDocs, want: []pair{{target: target.Docs, kinds: []source.Kind{source.Doc}}}},
		{name: "all", scope: source.ScopeAll, want: []pair{
			{target: targetName, kinds: []source.Kind{source.Skill, source.Agent}},
			{target: target.Docs, kinds: []source.Kind{source.Doc}},
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, pairs(tc.scope, targetName))
		})
	}
}
