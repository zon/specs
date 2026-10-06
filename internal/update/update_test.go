package update

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/render"
	"github.com/zon/specs/internal/source"
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

	err := updatePair(root, src, src, pair{target: source.Opencode, kinds: []source.Kind{source.Skill}}, render.Features{})
	require.NoError(t, err)
	require.Contains(t, reported(), src)
}

func TestPairsSelectsRunsPerScope(t *testing.T) {
	const targetName = source.Opencode
	cases := []struct {
		name   string
		scope  source.Scope
		agents bool
		want   []pair
	}{
		{name: "skills", scope: source.ScopeSkills, want: []pair{{target: targetName, kinds: []source.Kind{source.Skill}}}},
		{name: "agents", scope: source.ScopeAgents, want: []pair{{target: targetName, kinds: []source.Kind{source.Agent}}}},
		{name: "docs", scope: source.ScopeDocs, want: []pair{{target: source.Docs, kinds: []source.Kind{source.Doc}}}},
		{name: "all", scope: source.ScopeAll, want: []pair{
			{target: targetName, kinds: []source.Kind{source.Skill}},
			{target: source.Docs, kinds: []source.Kind{source.Doc}},
		}},
		{name: "all with agents", scope: source.ScopeAll, agents: true, want: []pair{
			{target: targetName, kinds: []source.Kind{source.Skill, source.Agent}},
			{target: source.Docs, kinds: []source.Kind{source.Doc}},
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, pairs(tc.scope, targetName, tc.agents))
		})
	}
}

func TestRunRendersSkillsAndAgents(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "code-architect")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: source.Opencode}))
	testutil.RequireWritten(t, root, source.Opencode, "prose-editor", source.Skill)
	testutil.RequireWritten(t, root, source.Opencode, "code-architect", source.Agent)
}

func TestRunSkipsAgentsWithoutTheFlag(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "code-architect")
	testutil.WriteDoc(t, src, "prose")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: source.Opencode}))
	testutil.RequireWritten(t, root, source.Opencode, "prose-editor", source.Skill)
	testutil.RequireNotWritten(t, root, source.Opencode, "code-architect", source.Agent)
	testutil.RequireWritten(t, root, source.Docs, "prose", source.Doc)
}

func TestRunAgentsScopeRendersAgentsWithoutTheFlag(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteAgent(t, src, "code-architect")

	require.NoError(t, Run(Options{Scope: source.ScopeAgents, Source: src, Target: source.Opencode}))
	testutil.RequireWritten(t, root, source.Opencode, "code-architect", source.Agent)
}

func TestRunRendersOnlyTheGivenSource(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.SkillSource(t, "local-only")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: source.Opencode}))
	testutil.RequireWritten(t, root, source.Opencode, "local-only", source.Skill)
	testutil.RequireNotWritten(t, root, source.Opencode, "prose-editor", source.Skill)
}

func TestUpdateWritesAgentUnderSourceNameForBothTargets(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.AgentSource(t, "prose-editor")

	cases := []string{source.Claude, source.Opencode}
	for _, trgt := range cases {
		t.Run(trgt, func(t *testing.T) {
			require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: trgt}))
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

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: source.Claude}))
	testutil.RequireWritten(t, root, source.Claude, "prose-editor", source.Skill)
	testutil.RequireWritten(t, root, source.Claude, "prose-editor", source.Agent)
}

func TestUpdateWritesSkillAndAgentToOpencode(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "prose-editor")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: source.Opencode}))
	testutil.RequireWritten(t, root, source.Opencode, "prose-editor", source.Skill)
	testutil.RequireWritten(t, root, source.Opencode, "prose-editor", source.Agent)
}

func TestUpdateRendersWhatTheCommandNames(t *testing.T) {
	cases := []struct {
		name      string
		scope     source.Scope
		agents    bool
		wantSkill bool
		wantAgent bool
		wantDoc   bool
	}{
		{name: "update renders skills, agents, and docs", scope: source.ScopeAll, agents: true, wantSkill: true, wantAgent: true, wantDoc: true},
		{name: "update renders skills and docs without the agents flag", scope: source.ScopeAll, wantSkill: true, wantDoc: true},
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

			require.NoError(t, Run(Options{Scope: tc.scope, Agents: tc.agents, Source: src, Target: source.Opencode}))

			if tc.wantSkill {
				testutil.RequireWritten(t, root, source.Opencode, "prose-editor", source.Skill)
			} else {
				testutil.RequireNotWritten(t, root, source.Opencode, "prose-editor", source.Skill)
			}
			if tc.wantAgent {
				testutil.RequireWritten(t, root, source.Opencode, "code-architect", source.Agent)
			} else {
				testutil.RequireNotWritten(t, root, source.Opencode, "code-architect", source.Agent)
			}
			if tc.wantDoc {
				testutil.RequireWritten(t, root, source.Docs, "prose", source.Doc)
			} else {
				testutil.RequireNotWritten(t, root, source.Docs, "prose", source.Doc)
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

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: source.Opencode}))

	testutil.RequireWritten(t, root, source.Opencode, "prose-editor", source.Skill)
	testutil.RequireNotWritten(t, work, source.Opencode, "prose-editor", source.Skill)
}

func TestUpdateErrorsOutsideRepository(t *testing.T) {
	root := t.TempDir()
	src := testutil.AgentSource(t, "prose-editor")

	t.Chdir(root)
	err := Run(Options{Scope: source.ScopeAll, Source: src, Target: source.Opencode})
	require.Error(t, err)
	testutil.RequireNotWritten(t, root, source.Opencode, "prose-editor", source.Agent)
}

func TestUpdateCreatesMissingDirectories(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")

	require.NoError(t, Run(Options{Scope: source.ScopeSkills, Source: src, Target: source.Claude}))
	testutil.RequireWritten(t, root, source.Claude, "prose-editor", source.Skill)
}

func TestUpdateLeavesForeignFileAlone(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.AgentSource(t, "prose-editor")

	testutil.SeedForeignFile(t, root, source.Claude, "prose-editor", source.Agent, "manual content\n")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: source.Claude}))

	require.Equal(t, "manual content\n", testutil.WrittenContent(t, root, source.Claude, "prose-editor", source.Agent))
}

func TestUpdateReplacesOwnedFiles(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteAgentBody(t, src, "prose-editor", "First.\n")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: source.Opencode}))

	testutil.WriteAgentBody(t, src, "prose-editor", "Second.\n")
	require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: source.Opencode}))

	require.Contains(t, testutil.WrittenContent(t, root, source.Opencode, "prose-editor", source.Agent), "Second.")
}

func TestUpdateRemovesStaleSkill(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.SkillSource(t, "prose-editor")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: source.Opencode}))
	testutil.RequireWritten(t, root, source.Opencode, "prose-editor", source.Skill)

	require.NoError(t, os.RemoveAll(filepath.Join(src, "skills")))

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Source: src, Target: source.Opencode}))
	testutil.RequireNotWritten(t, root, source.Opencode, "prose-editor", source.Skill)
}

func TestUpdateRemovesStaleAgent(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.AgentSource(t, "prose-editor")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: source.Claude}))
	testutil.RequireWritten(t, root, source.Claude, "prose-editor", source.Agent)

	require.NoError(t, os.RemoveAll(filepath.Join(src, "agents")))

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: source.Claude}))
	testutil.RequireNotWritten(t, root, source.Claude, "prose-editor", source.Agent)
}

func TestUpdateAllWritesDocs(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "code-architect")
	testutil.WriteDoc(t, src, "prose")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: source.Opencode}))
	testutil.RequireWritten(t, root, source.Opencode, "prose-editor", source.Skill)
	testutil.RequireWritten(t, root, source.Opencode, "code-architect", source.Agent)
	testutil.RequireWritten(t, root, source.Docs, "prose", source.Doc)
}

func TestUpdateAllWritesDocsToTheTargetItNames(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteAgent(t, src, "code-architect")
	testutil.WriteDoc(t, src, "prose")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: source.Claude}))
	testutil.RequireWritten(t, root, source.Claude, "code-architect", source.Agent)
	testutil.RequireWritten(t, root, source.Docs, "prose", source.Doc)
}

func TestUpdateDocsWritesFromSource(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "code-architect")
	testutil.WriteDoc(t, src, "architecture")

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Opencode}))

	testutil.RequireWritten(t, root, source.Docs, "architecture", source.Doc)
	testutil.RequireNotWritten(t, root, source.Opencode, "prose-editor", source.Skill)
	testutil.RequireNotWritten(t, root, source.Opencode, "code-architect", source.Agent)
	testutil.RequireNotWritten(t, root, source.Claude, "prose-editor", source.Skill)
	testutil.RequireNotWritten(t, root, source.Claude, "code-architect", source.Agent)
}

func TestUpdateDocsIgnoresTarget(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "code-architect")
	testutil.WriteDoc(t, src, "prose")

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Claude}))

	testutil.RequireWritten(t, root, source.Docs, "prose", source.Doc)
	testutil.RequireNotWritten(t, root, source.Claude, "prose-editor", source.Skill)
	testutil.RequireNotWritten(t, root, source.Claude, "code-architect", source.Agent)
	testutil.RequireNotWritten(t, root, source.Opencode, "prose-editor", source.Skill)
	testutil.RequireNotWritten(t, root, source.Opencode, "code-architect", source.Agent)
}

func TestUpdateDocsLeavesForeignFileAlone(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.DocSource(t, "architecture")

	testutil.SeedForeignFile(t, root, source.Docs, "prose", source.Doc, "manual content\n")

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Opencode}))

	require.Equal(t, "manual content\n", testutil.WrittenContent(t, root, source.Docs, "prose", source.Doc))
}

func TestUpdateDocsReplacesOwned(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.DocSource(t, "architecture")

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Opencode}))

	testutil.WriteDocBody(t, src, "architecture", "# Architecture, second\n")
	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Opencode}))

	require.Contains(t, testutil.WrittenContent(t, root, source.Docs, "architecture", source.Doc), "second")
}

func TestUpdateDocsRemovesStale(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := testutil.DocSource(t, "architecture")

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Opencode}))
	testutil.RequireWritten(t, root, source.Docs, "architecture", source.Doc)

	require.NoError(t, os.RemoveAll(filepath.Join(src, "docs")))

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Opencode}))
	testutil.RequireNotWritten(t, root, source.Docs, "architecture", source.Doc)
}

func TestUpdateDocsWritesToRepositoryRoot(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	src := testutil.DocSource(t, "architecture")

	work := filepath.Join(root, "nested", "deep")
	require.NoError(t, os.MkdirAll(work, 0o755))
	t.Chdir(work)

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Opencode}))

	testutil.RequireWritten(t, root, source.Docs, "architecture", source.Doc)
	testutil.RequireNotWritten(t, work, source.Docs, "architecture", source.Doc)
}

func TestUpdateDocsErrorsOutsideRepository(t *testing.T) {
	root := t.TempDir()
	src := testutil.DocSource(t, "architecture")

	t.Chdir(root)
	err := Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Opencode})
	require.Error(t, err)
	testutil.RequireNotWritten(t, root, source.Docs, "architecture", source.Doc)
}

func TestUpdateSkipsDefinitionThatRendersToNothing(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteDocBody(t, src, "orchestration", "{{- if .orchestration}}\n# Orchestration\n{{- end}}\n")

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Opencode}))
	testutil.RequireNotWritten(t, root, source.Docs, "orchestration", source.Doc)
	_, err := os.Stat(filepath.Join(root, "docs", "zpecs", ".zpecs"))
	require.True(t, os.IsNotExist(err), "manifest should be absent when nothing is written")

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Orchestration: true, Source: src, Target: source.Opencode}))
	testutil.RequireWritten(t, root, source.Docs, "orchestration", source.Doc)

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Opencode}))
	testutil.RequireNotWritten(t, root, source.Docs, "orchestration", source.Doc)
}

func TestUpdateSummaryCountsOnlyWrittenDefinitions(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteDoc(t, src, "prose")
	testutil.WriteDocBody(t, src, "orchestration", "{{- if .orchestration}}\n# Orchestration\n{{- end}}\n")
	reported := testutil.CaptureReport(t)

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Opencode}))
	require.Contains(t, reported(), "(1 files)")
	testutil.RequireWritten(t, root, source.Docs, "prose", source.Doc)
	testutil.RequireNotWritten(t, root, source.Docs, "orchestration", source.Doc)

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Orchestration: true, Source: src, Target: source.Opencode}))
	require.Contains(t, reported(), "(2 files)")
}

// TestUpdateGatesOptionalContentForEachKind checks the orchestration flag
// for each kind at the update level. Optional content appears in the
// written definition only when the flag is on.
func TestUpdateGatesOptionalContentForEachKind(t *testing.T) {
	content := "Kept.\n{{- if .orchestration}}\nOrchestrate.\n{{- end}}\n"
	cases := []struct {
		name    string
		kind    source.Kind
		scope   source.Scope
		target  string
		defName string
		write   func(*testing.T, string, string, string)
	}{
		{name: "skill", kind: source.Skill, scope: source.ScopeSkills, target: source.Opencode, defName: "prose-editor", write: testutil.WriteSkillBody},
		{name: "agent", kind: source.Agent, scope: source.ScopeAgents, target: source.Opencode, defName: "code-architect", write: testutil.WriteAgentBody},
		{name: "doc", kind: source.Doc, scope: source.ScopeDocs, target: source.Docs, defName: "architecture", write: testutil.WriteDocBody},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := testutil.GitRepo(t, nil)
			t.Chdir(root)
			src := t.TempDir()
			tc.write(t, src, tc.defName, content)

			require.NoError(t, Run(Options{Scope: tc.scope, Source: src, Target: tc.target}))
			off := testutil.WrittenContent(t, root, tc.target, tc.defName, tc.kind)
			require.Contains(t, off, "Kept.")
			require.NotContains(t, off, "Orchestrate.")

			require.NoError(t, Run(Options{Scope: tc.scope, Orchestration: true, Source: src, Target: tc.target}))
			on := testutil.WrittenContent(t, root, tc.target, tc.defName, tc.kind)
			require.Contains(t, on, "Kept.")
			require.Contains(t, on, "Orchestrate.")
		})
	}
}

// TestUpdateGatedLinkDisappearsAndReturns checks that a link to gated
// content is absent from the written definition when the flag is off and
// present when it is on.
func TestUpdateGatedLinkDisappearsAndReturns(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	link := "[Orchestration](orchestration.md)"
	body := "# Architecture\n\nSee the docs.\n{{- if .orchestration}}\n" + link + "\n{{- end}}\n"
	testutil.WriteDocBody(t, src, "architecture", body)

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Source: src, Target: source.Opencode}))
	off := testutil.WrittenContent(t, root, source.Docs, "architecture", source.Doc)
	require.NotContains(t, off, link)

	require.NoError(t, Run(Options{Scope: source.ScopeDocs, Orchestration: true, Source: src, Target: source.Opencode}))
	on := testutil.WrittenContent(t, root, source.Docs, "architecture", source.Doc)
	require.Contains(t, on, link)
}

func TestUpdateScopedRemovalLeavesOtherKinds(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	src := t.TempDir()
	testutil.WriteSkill(t, src, "prose-editor")
	testutil.WriteAgent(t, src, "code-architect")

	require.NoError(t, Run(Options{Scope: source.ScopeAll, Agents: true, Source: src, Target: source.Opencode}))
	testutil.RequireWritten(t, root, source.Opencode, "prose-editor", source.Skill)
	testutil.RequireWritten(t, root, source.Opencode, "code-architect", source.Agent)

	require.NoError(t, os.RemoveAll(filepath.Join(src, "skills")))

	require.NoError(t, Run(Options{Scope: source.ScopeSkills, Source: src, Target: source.Opencode}))
	testutil.RequireNotWritten(t, root, source.Opencode, "prose-editor", source.Skill)
	testutil.RequireWritten(t, root, source.Opencode, "code-architect", source.Agent)
}
