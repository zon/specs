package update

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
	require.FileExists(t, filepath.Join(dir, "seed"))
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

func TestUpdateReadsFromSource(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")

	require.NoError(t, Run(Options{Scope: "all", Source: dir, Target: "opencode"}))
	_, err := os.Stat(filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(".opencode", "agents", "code-architect.md"))
	require.NoError(t, err)
}

func TestUpdateReadsSameFrontmatterForBothTargets(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), `---
name: prose-editor
description: Reviews prose.
tools:
  - read
  - edit
---

Review prose.
`)

	cases := []string{target.Claude, target.Opencode}
	for _, trgt := range cases {
		t.Run(trgt, func(t *testing.T) {
			err := Run(Options{Scope: "all", Source: dir, Target: trgt})
			require.NoError(t, err)
		})
	}
}

func TestUpdateWritesAgentUnderSourceNameForBothTargets(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), `---
name: renamed
description: Reviews prose.
---
Review prose.
`)

	cases := []string{target.Claude, target.Opencode}
	for _, trgt := range cases {
		t.Run(trgt, func(t *testing.T) {
			err := Run(Options{Scope: "all", Source: dir, Target: trgt})
			require.NoError(t, err)
			path := filepath.Join("."+trgt, "agents", "prose-editor.md")
			content, err := os.ReadFile(path)
			require.NoError(t, err)
			if trgt == target.Claude {
				require.Contains(t, string(content), "name: renamed")
			} else {
				require.NotContains(t, string(content), "name:")
			}
		})
	}
}

func TestUpdateWritesSkillAndAgentToClaude(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), `---
name: prose-editor
description: Reviews prose.
---
Review prose.
`)

	err := Run(Options{Scope: "all", Source: dir, Target: target.Claude})
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(".claude", "skills", "prose-editor", "SKILL.md"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(".claude", "agents", "prose-editor.md"))
	require.NoError(t, err)
}

func TestUpdateWritesSkillAndAgentToOpencode(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), `---
name: prose-editor
description: Reviews prose.
---
Review prose.
`)

	err := Run(Options{Scope: "all", Source: dir, Target: target.Opencode})
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(".opencode", "agents", "prose-editor.md"))
	require.NoError(t, err)
}

func TestUpdateRendersWhatTheCommandNames(t *testing.T) {
	cases := []struct {
		name      string
		scope     string
		wantSkill bool
		wantAgent bool
		wantDoc   bool
	}{
		{name: "update renders skills, agents, and docs", scope: "all", wantSkill: true, wantAgent: true, wantDoc: true},
		{name: "update skills renders skills only", scope: "skills", wantSkill: true},
		{name: "update agents renders agents only", scope: "agents", wantAgent: true},
		{name: "update docs renders docs only", scope: "docs", wantDoc: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(testutil.GitRepo(t, nil))
			dir := t.TempDir()
			testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
			testutil.WriteSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")
			testutil.WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

			require.NoError(t, Run(Options{Scope: tc.scope, Source: dir, Target: "opencode"}))

			skillPath := filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")
			agentPath := filepath.Join(".opencode", "agents", "code-architect.md")
			docPath := filepath.Join("docs", "zpecs", "prose.md")
			if tc.wantSkill {
				_, err := os.Stat(skillPath)
				require.NoError(t, err)
			} else {
				_, err := os.Stat(skillPath)
				require.Error(t, err)
			}
			if tc.wantAgent {
				_, err := os.Stat(agentPath)
				require.NoError(t, err)
			} else {
				_, err := os.Stat(agentPath)
				require.Error(t, err)
			}
			if tc.wantDoc {
				_, err := os.Stat(docPath)
				require.NoError(t, err)
			} else {
				_, err := os.Stat(docPath)
				require.Error(t, err)
			}
		})
	}
}

func TestUpdateWritesToRepositoryRoot(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	work := filepath.Join(root, "nested", "deep")
	require.NoError(t, os.MkdirAll(work, 0o755))
	t.Chdir(work)

	require.NoError(t, Run(Options{Scope: "all", Source: dir, Target: "opencode"}))

	skillPath := filepath.Join(root, ".opencode", "skills", "prose-editor", "SKILL.md")
	_, err := os.Stat(skillPath)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(work, ".opencode", "skills", "prose-editor", "SKILL.md"))
	require.Error(t, err)
}

func TestUpdateErrorsOutsideRepository(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	t.Chdir(t.TempDir())
	err := Run(Options{Scope: "all", Source: dir, Target: "opencode"})
	require.Error(t, err)
	_, statErr := os.Stat(filepath.Join(".opencode", "agents", "prose-editor.md"))
	require.Error(t, statErr)
}

func TestUpdateCreatesMissingDirectories(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	require.NoError(t, Run(Options{Scope: "skills", Source: dir, Target: target.Claude}))

	for _, path := range []string{
		filepath.Join(".claude"),
		filepath.Join(".claude", "skills"),
		filepath.Join(".claude", "skills", "prose-editor"),
	} {
		info, err := os.Stat(path)
		require.NoError(t, err)
		require.True(t, info.IsDir())
	}
}

func TestUpdateLeavesForeignFileAlone(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	path := filepath.Join(".claude", "agents", "prose-editor.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("manual content\n"), 0o644))

	require.NoError(t, Run(Options{Scope: "all", Source: dir, Target: target.Claude}))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "manual content\n", string(content))
}

func TestUpdateReplacesOwnedFiles(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nFirst.\n")

	require.NoError(t, Run(Options{Scope: "all", Source: dir, Target: "opencode"}))

	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nSecond.\n")
	require.NoError(t, Run(Options{Scope: "all", Source: dir, Target: "opencode"}))

	content, err := os.ReadFile(filepath.Join(".opencode", "agents", "prose-editor.md"))
	require.NoError(t, err)
	require.Contains(t, string(content), "Second.")
}

func TestUpdateRemovesStaleSkill(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	require.NoError(t, Run(Options{Scope: "all", Source: dir, Target: "opencode"}))
	path := filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")
	_, err := os.Stat(path)
	require.NoError(t, err)

	require.NoError(t, os.RemoveAll(filepath.Join(dir, "skills")))

	require.NoError(t, Run(Options{Scope: "all", Source: dir, Target: "opencode"}))
	_, err = os.Stat(path)
	require.Error(t, err)
}

func TestUpdateRemovesStaleAgent(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	require.NoError(t, Run(Options{Scope: "all", Source: dir, Target: target.Claude}))
	path := filepath.Join(".claude", "agents", "prose-editor.md")
	_, err := os.Stat(path)
	require.NoError(t, err)

	require.NoError(t, os.RemoveAll(filepath.Join(dir, "agents")))

	require.NoError(t, Run(Options{Scope: "all", Source: dir, Target: target.Claude}))
	_, err = os.Stat(path)
	require.Error(t, err)
}

func TestUpdateAllWritesDocs(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")
	testutil.WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	require.NoError(t, Run(Options{Scope: "all", Source: dir, Target: "opencode"}))

	for _, path := range []string{
		filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md"),
		filepath.Join(".opencode", "agents", "code-architect.md"),
		filepath.Join("docs", "zpecs", "prose.md"),
	} {
		_, err := os.Stat(path)
		require.NoError(t, err)
	}
}

func TestUpdateAllWritesDocsToTheTargetItNames(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")
	testutil.WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	require.NoError(t, Run(Options{Scope: "all", Source: dir, Target: target.Claude}))

	_, err := os.Stat(filepath.Join(".claude", "agents", "code-architect.md"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join("docs", "zpecs", "prose.md"))
	require.NoError(t, err)
}

func TestUpdateDocsWritesFromSource(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	require.NoError(t, Run(Options{Scope: "docs", Source: dir, Target: "opencode"}))

	path := filepath.Join("docs", "zpecs", "architecture.md")
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "# Architecture\n", string(content))
	_, err = os.Stat(filepath.Join(".opencode"))
	require.Error(t, err)
	_, err = os.Stat(filepath.Join(".claude"))
	require.Error(t, err)
}

func TestUpdateDocsIgnoresTarget(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	require.NoError(t, Run(Options{Scope: "docs", Source: dir, Target: target.Claude}))

	path := filepath.Join("docs", "zpecs", "prose.md")
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "# Prose guidelines\n", string(content))
	_, err = os.Stat(filepath.Join(".claude"))
	require.Error(t, err)
	_, err = os.Stat(filepath.Join(".opencode"))
	require.Error(t, err)
}

func TestUpdateDocsLeavesForeignFileAlone(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	path := filepath.Join("docs", "zpecs", "prose.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("manual content\n"), 0o644))

	require.NoError(t, Run(Options{Scope: "docs", Source: dir, Target: "opencode"}))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "manual content\n", string(content))
}

func TestUpdateDocsReplacesOwned(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	require.NoError(t, Run(Options{Scope: "docs", Source: dir, Target: "opencode"}))

	testutil.WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture, second\n")
	require.NoError(t, Run(Options{Scope: "docs", Source: dir, Target: "opencode"}))

	content, err := os.ReadFile(filepath.Join("docs", "zpecs", "architecture.md"))
	require.NoError(t, err)
	require.Contains(t, string(content), "second")
}

func TestUpdateDocsRemovesStale(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	require.NoError(t, Run(Options{Scope: "docs", Source: dir, Target: "opencode"}))
	path := filepath.Join("docs", "zpecs", "architecture.md")
	_, err := os.Stat(path)
	require.NoError(t, err)

	require.NoError(t, os.RemoveAll(filepath.Join(dir, "docs")))

	require.NoError(t, Run(Options{Scope: "docs", Source: dir, Target: "opencode"}))
	_, err = os.Stat(path)
	require.Error(t, err)
}

func TestUpdateDocsWritesToRepositoryRoot(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	work := filepath.Join(root, "nested", "deep")
	require.NoError(t, os.MkdirAll(work, 0o755))
	t.Chdir(work)

	require.NoError(t, Run(Options{Scope: "docs", Source: dir, Target: "opencode"}))

	path := filepath.Join(root, "docs", "zpecs", "architecture.md")
	_, err := os.Stat(path)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(work, "docs", "zpecs", "architecture.md"))
	require.Error(t, err)
}

func TestUpdateDocsErrorsOutsideRepository(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	t.Chdir(t.TempDir())
	err := Run(Options{Scope: "docs", Source: dir, Target: "opencode"})
	require.Error(t, err)
	_, statErr := os.Stat(filepath.Join("docs", "zpecs", "architecture.md"))
	require.Error(t, statErr)
}

func TestUpdateScopedRemovalLeavesOtherKinds(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")

	require.NoError(t, Run(Options{Scope: "all", Source: dir, Target: "opencode"}))
	skillPath := filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")
	agentPath := filepath.Join(".opencode", "agents", "code-architect.md")
	_, err := os.Stat(skillPath)
	require.NoError(t, err)
	_, err = os.Stat(agentPath)
	require.NoError(t, err)

	require.NoError(t, os.RemoveAll(filepath.Join(dir, "skills")))

	require.NoError(t, Run(Options{Scope: "skills", Source: dir, Target: "opencode"}))
	_, err = os.Stat(skillPath)
	require.Error(t, err)
	_, err = os.Stat(agentPath)
	require.NoError(t, err)
}
