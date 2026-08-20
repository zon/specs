package update

import (
	"os"
	"path/filepath"
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

func TestRunRendersSkillsAndAgents(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "code-architect")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Opencode}))
	testutil.RequireWritten(t, root, target.Opencode, "prose-editor", source.Skill)
	testutil.RequireWritten(t, root, target.Opencode, "code-architect", source.Agent)
}

func TestRunRendersOnlyTheGivenSource(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.SkillSource(t, "local-only")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Opencode}))
	testutil.RequireWritten(t, root, target.Opencode, "local-only", source.Skill)
	testutil.RequireNotWritten(t, root, target.Opencode, "prose-editor", source.Skill)
}

func TestUpdateWritesAgentUnderSourceNameForBothTargets(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.AgentSource(t, "prose-editor")

	cases := []string{target.Claude, target.Opencode}
	for _, trgt := range cases {
		t.Run(trgt, func(t *testing.T) {
			require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: trgt}))
			testutil.RequireWritten(t, root, trgt, "prose-editor", source.Agent)
		})
	}
}

func TestUpdateWritesSkillAndAgentToClaude(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "prose-editor")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Claude}))
	testutil.RequireWritten(t, root, target.Claude, "prose-editor", source.Skill)
	testutil.RequireWritten(t, root, target.Claude, "prose-editor", source.Agent)
}

func TestUpdateWritesSkillAndAgentToOpencode(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "prose-editor")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Opencode}))
	testutil.RequireWritten(t, root, target.Opencode, "prose-editor", source.Skill)
	testutil.RequireWritten(t, root, target.Opencode, "prose-editor", source.Agent)
}

func TestUpdateRendersWhatTheCommandNames(t *testing.T) {
	cases := []struct {
		name      string
		scope     source.Scope
		wantSkill bool
		wantAgent bool
		wantDoc   bool
	}{
		{name: "update renders skills, agents, and docs", scope: source.ScopeAll, wantSkill: true, wantAgent: true, wantDoc: true},
		{name: "update skills renders skills only", scope: source.ScopeSkills, wantSkill: true},
		{name: "update agents renders agents only", scope: source.ScopeAgents, wantAgent: true},
		{name: "update docs renders docs only", scope: source.ScopeDocs, wantDoc: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := testutil.GitRepo(t, nil)
			t.Chdir(root)
			src := t.TempDir()
			testutil.WriteSkill(t, src, "prose-editor")
			testutil.WriteAgent(t, src, "code-architect")
			testutil.WriteDoc(t, src, "prose")

			require.NoError(t, Run(Options{Scope: tc.scope, Source: src, Target: target.Opencode}))

			if tc.wantSkill {
				testutil.RequireWritten(t, root, target.Opencode, "prose-editor", source.Skill)
			} else {
				testutil.RequireNotWritten(t, root, target.Opencode, "prose-editor", source.Skill)
			}
			if tc.wantAgent {
				testutil.RequireWritten(t, root, target.Opencode, "code-architect", source.Agent)
			} else {
				testutil.RequireNotWritten(t, root, target.Opencode, "code-architect", source.Agent)
			}
			if tc.wantDoc {
				testutil.RequireWritten(t, root, target.Docs, "prose", source.Doc)
			} else {
				testutil.RequireNotWritten(t, root, target.Docs, "prose", source.Doc)
			}
		})
	}
}

func TestUpdateWritesToRepositoryRoot(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "prose-editor")

	work := filepath.Join(root, "nested", "deep")
	require.NoError(t, os.MkdirAll(work, 0o755))
	t.Chdir(work)

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Opencode}))

	testutil.RequireWritten(t, root, target.Opencode, "prose-editor", source.Skill)
	testutil.RequireNotWritten(t, work, target.Opencode, "prose-editor", source.Skill)
}

func TestUpdateErrorsOutsideRepository(t *testing.T) {
	root := t.TempDir()
	src := testutil.AgentSource(t, "prose-editor")

	t.Chdir(root)
	err := Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Opencode})
	require.Error(t, err)
	testutil.RequireNotWritten(t, root, target.Opencode, "prose-editor", source.Agent)
}

func TestUpdateCreatesMissingDirectories(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")

	require.NoError(t, Run(Options{Scope: source.ScopeSkills, Source: src, Target: target.Claude}))
	testutil.RequireWritten(t, root, target.Claude, "prose-editor", source.Skill)
}

func TestUpdateLeavesForeignFileAlone(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.AgentSource(t, "prose-editor")

	testutil.SeedForeignFile(t, root, target.Claude, "prose-editor", source.Agent, "manual content\n")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Claude}))

	require.Equal(t, "manual content\n", testutil.WrittenContent(t, root, target.Claude, "prose-editor", source.Agent))
}

func TestUpdateReplacesOwnedFiles(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteAgentBody(t, src, "prose-editor", "First.\n")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Opencode}))

	testutil.WriteAgentBody(t, src, "prose-editor", "Second.\n")
	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Opencode}))

	require.Contains(t, testutil.WrittenContent(t, root, target.Opencode, "prose-editor", source.Agent), "Second.")
}

func TestUpdateRemovesStaleSkill(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.SkillSource(t, "prose-editor")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Opencode}))
	testutil.RequireWritten(t, root, target.Opencode, "prose-editor", source.Skill)

	require.NoError(t, os.RemoveAll(filepath.Join(src, "skills")))

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Opencode}))
	testutil.RequireNotWritten(t, root, target.Opencode, "prose-editor", source.Skill)
}

func TestUpdateRemovesStaleAgent(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.AgentSource(t, "prose-editor")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Claude}))
	testutil.RequireWritten(t, root, target.Claude, "prose-editor", source.Agent)

	require.NoError(t, os.RemoveAll(filepath.Join(src, "agents")))

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Claude}))
	testutil.RequireNotWritten(t, root, target.Claude, "prose-editor", source.Agent)
}

func TestUpdateAllWritesDocs(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "code-architect")
	testutil.WriteDoc(t, src, "prose")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Opencode}))
	testutil.RequireWritten(t, root, target.Opencode, "prose-editor", source.Skill)
	testutil.RequireWritten(t, root, target.Opencode, "code-architect", source.Agent)
	testutil.RequireWritten(t, root, target.Docs, "prose", source.Doc)
}

func TestUpdateAllWritesDocsToTheTargetItNames(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteAgent(t, src, "code-architect")
	testutil.WriteDoc(t, src, "prose")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Claude}))
	testutil.RequireWritten(t, root, target.Claude, "code-architect", source.Agent)
	testutil.RequireWritten(t, root, target.Docs, "prose", source.Doc)
}

func TestUpdateDocsWritesFromSource(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "code-architect")
	testutil.WriteDoc(t, src, "architecture")

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: target.Opencode}))

	testutil.RequireWritten(t, root, target.Docs, "architecture", source.Doc)
	testutil.RequireNotWritten(t, root, target.Opencode, "prose-editor", source.Skill)
	testutil.RequireNotWritten(t, root, target.Opencode, "code-architect", source.Agent)
	testutil.RequireNotWritten(t, root, target.Claude, "prose-editor", source.Skill)
	testutil.RequireNotWritten(t, root, target.Claude, "code-architect", source.Agent)
}

func TestUpdateDocsIgnoresTarget(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "code-architect")
	testutil.WriteDoc(t, src, "prose")

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: target.Claude}))

	testutil.RequireWritten(t, root, target.Docs, "prose", source.Doc)
	testutil.RequireNotWritten(t, root, target.Claude, "prose-editor", source.Skill)
	testutil.RequireNotWritten(t, root, target.Claude, "code-architect", source.Agent)
	testutil.RequireNotWritten(t, root, target.Opencode, "prose-editor", source.Skill)
	testutil.RequireNotWritten(t, root, target.Opencode, "code-architect", source.Agent)
}

func TestUpdateDocsLeavesForeignFileAlone(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.DocSource(t, "architecture")

	testutil.SeedForeignFile(t, root, target.Docs, "prose", source.Doc, "manual content\n")

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: target.Opencode}))

	require.Equal(t, "manual content\n", testutil.WrittenContent(t, root, target.Docs, "prose", source.Doc))
}

func TestUpdateDocsReplacesOwned(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.DocSource(t, "architecture")

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: target.Opencode}))

	testutil.WriteDocBody(t, src, "architecture", "# Architecture, second\n")
	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: target.Opencode}))

	require.Contains(t, testutil.WrittenContent(t, root, target.Docs, "architecture", source.Doc), "second")
}

func TestUpdateDocsRemovesStale(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.DocSource(t, "architecture")

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: target.Opencode}))
	testutil.RequireWritten(t, root, target.Docs, "architecture", source.Doc)

	require.NoError(t, os.RemoveAll(filepath.Join(src, "docs")))

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: target.Opencode}))
	testutil.RequireNotWritten(t, root, target.Docs, "architecture", source.Doc)
}

func TestUpdateDocsWritesToRepositoryRoot(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	src := testutil.DocSource(t, "architecture")

	work := filepath.Join(root, "nested", "deep")
	require.NoError(t, os.MkdirAll(work, 0o755))
	t.Chdir(work)

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: target.Opencode}))

	testutil.RequireWritten(t, root, target.Docs, "architecture", source.Doc)
	testutil.RequireNotWritten(t, work, target.Docs, "architecture", source.Doc)
}

func TestUpdateDocsErrorsOutsideRepository(t *testing.T) {
	root := t.TempDir()
	src := testutil.DocSource(t, "architecture")

	t.Chdir(root)
	err := Run(Options{Scope: source.ScopeDocs, Source: src, Target: target.Opencode})
	require.Error(t, err)
	testutil.RequireNotWritten(t, root, target.Docs, "architecture", source.Doc)
}

func TestUpdateScopedRemovalLeavesOtherKinds(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "code-architect")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: target.Opencode}))
	testutil.RequireWritten(t, root, target.Opencode, "prose-editor", source.Skill)
	testutil.RequireWritten(t, root, target.Opencode, "code-architect", source.Agent)

	require.NoError(t, os.RemoveAll(filepath.Join(src, "skills")))

	require.NoError(t, Run(Options{Scope: source.ScopeSkills, Source: src, Target: target.Opencode}))
	testutil.RequireNotWritten(t, root, target.Opencode, "prose-editor", source.Skill)
	testutil.RequireWritten(t, root, target.Opencode, "code-architect", source.Agent)
}
