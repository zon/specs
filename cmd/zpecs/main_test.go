package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/spec"
)

func buildBinary(t *testing.T, dir string) string {
	t.Helper()
	binary := filepath.Join(dir, "zpecs")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, ".")
	out, err := build.CombinedOutput()
	require.NoError(t, err, "go build failed\n%s", out)
	return binary
}

func TestBuildProducesRunnableCLIBinary(t *testing.T) {
	binary := buildBinary(t, t.TempDir())

	info, err := os.Stat(binary)
	require.NoError(t, err)
	require.False(t, runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0, "binary is not executable: %v", info.Mode())

	cmd := exec.Command(binary)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "binary failed to run\n%s", out)
}

func TestBinaryPrintsVersion(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	out, err := exec.Command(binary, "--version").CombinedOutput()
	require.NoError(t, err, "zpecs --version failed\n%s", out)
	require.Equal(t, version, strings.TrimSpace(string(out)))
}

func TestUnmarshalScope(t *testing.T) {
	cases := []struct {
		name    string
		s       string
		want    scope
		wantErr bool
	}{
		{name: "all", s: "all", want: scopeAll},
		{name: "skills", s: "skills", want: scopeSkills},
		{name: "agents", s: "agents", want: scopeAgents},
		{name: "docs", s: "docs", want: scopeDocs},
		{name: "unknown scope", s: "vscode", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got scope
			err := got.UnmarshalText([]byte(tc.s))
			require.Equal(t, tc.wantErr, err != nil)
			if err == nil {
				require.Equal(t, tc.want, got)
			}
		})
	}
}

// parseUpdateArgs parses `update <args>` with the same kong grammar
// `run` uses. It returns the selected options.
func parseUpdateArgs(t *testing.T, args ...string) (options, error) {
	t.Helper()
	var c cli
	parser, err := kong.New(&c, cliVars)
	require.NoError(t, err)
	if _, err := parser.Parse(append([]string{"update"}, args...)); err != nil {
		return options{}, err
	}
	return options{scope: c.Update.Scope, source: c.Update.Source, target: c.Update.Target}, nil
}

func TestParseUpdate(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    options
		wantErr bool
	}{
		{name: "defaults", args: nil, want: options{scope: scopeAll, target: targetOpencode, source: defaultSourceURL}},
		{name: "all scope", args: []string{"all"}, want: options{scope: scopeAll, target: targetOpencode, source: defaultSourceURL}},
		{name: "scope", args: []string{"skills"}, want: options{scope: scopeSkills, target: targetOpencode, source: defaultSourceURL}},
		{name: "docs scope", args: []string{"docs"}, want: options{scope: scopeDocs, target: targetOpencode, source: defaultSourceURL}},
		{name: "claude target", args: []string{"--target", "claude"}, want: options{scope: scopeAll, target: targetClaude, source: defaultSourceURL}},
		{name: "claude target equals", args: []string{"--target=claude"}, want: options{scope: scopeAll, target: targetClaude, source: defaultSourceURL}},
		{name: "source", args: []string{"--source", "/tmp/src"}, want: options{scope: scopeAll, target: targetOpencode, source: "/tmp/src"}},
		{name: "source equals", args: []string{"--source=/tmp/src"}, want: options{scope: scopeAll, target: targetOpencode, source: "/tmp/src"}},
		{name: "all together", args: []string{"agents", "--source", "/tmp/src", "--target", "claude"}, want: options{scope: scopeAgents, target: targetClaude, source: "/tmp/src"}},
		{name: "unknown scope", args: []string{"vscode"}, wantErr: true},
		{name: "unknown target", args: []string{"--target", "vscode"}, wantErr: true},
		{name: "missing target value", args: []string{"--target"}, wantErr: true},
		{name: "missing source value", args: []string{"--source"}, wantErr: true},
		{name: "unknown flag", args: []string{"--force"}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseUpdateArgs(t, tc.args...)
			require.Equal(t, tc.wantErr, err != nil)
			if err == nil {
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestParseUpdateEnvOverridesDefault(t *testing.T) {
	t.Setenv("ZPECS_SOURCE", "/env/src")
	got, err := parseUpdateArgs(t)
	require.NoError(t, err)
	want := options{scope: scopeAll, target: targetOpencode, source: "/env/src"}
	require.Equal(t, want, got)
}

func TestRunRecognizesCommands(t *testing.T) {
	t.Chdir(gitRepo(t))
	t.Setenv("ZPECS_SOURCE", gitCloneSource(t))
	sourceDir := t.TempDir()
	cases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: nil},
		{name: "update", args: []string{"update"}},
		{name: "update skills", args: []string{"update", "skills"}},
		{name: "update agents", args: []string{"update", "agents"}},
		{name: "update with target", args: []string{"update", "--target", "claude"}},
		{name: "update with source", args: []string{"update", "skills", "--source", sourceDir}},
		{name: "unknown command", args: []string{"install"}, wantErr: true},
		{name: "update docs", args: []string{"update", "docs"}},
		{name: "unknown scope", args: []string{"update", "vscode"}, wantErr: true},
		{name: "too many arguments", args: []string{"update", "skills", "agents"}, wantErr: true},
		{name: "invalid target", args: []string{"update", "--target", "vscode"}, wantErr: true},
		{name: "unknown flag", args: []string{"update", "--force"}, wantErr: true},
		{name: "missing source", args: []string{"update", "--source", "/tmp/nope/does-not-exist"}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := run(tc.args)
			require.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestPrintErrorShowsUsageForParseErrors(t *testing.T) {
	var c cli
	parser, err := kong.New(&c, cliVars)
	require.NoError(t, err)
	_, parseErr := parser.Parse([]string{"update", "--target", "vscode"})
	require.Error(t, parseErr)

	stderr := captureStderr(t)
	printError(parseErr)
	out := string(stderr())
	require.Contains(t, out, "Usage:")
}

func TestPrintErrorOmitsUsageForRuntimeErrors(t *testing.T) {
	stderr := captureStderr(t)
	printError(errors.New("update outside a repository"))
	out := string(stderr())
	require.NotContains(t, out, "Usage:")
}

func TestBinaryRunsEachUpdateCommand(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	repoDir := gitRepo(t)
	sourceDir := t.TempDir()
	src := gitCloneSource(t)
	cases := []struct {
		args []string
		want string
	}{
		{args: []string{"update"}, want: "updating skills and agents for opencode"},
		{args: []string{"update", "skills"}, want: "updating skills for opencode"},
		{args: []string{"update", "agents"}, want: "updating agents for opencode"},
		{args: []string{"update", "docs"}, want: "updating docs"},
		{args: []string{"update", "--target", "claude"}, want: "updating skills and agents for claude"},
		{args: []string{"update", "--source", sourceDir, "skills"}, want: "updating skills for opencode from " + sourceDir},
	}

	for _, tc := range cases {
		cmd := exec.Command(binary, tc.args...)
		cmd.Dir = repoDir
		cmd.Env = append(os.Environ(), "ZPECS_SOURCE="+src)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "zpecs %v failed", tc.args)
		require.Contains(t, string(out), tc.want)
	}
}

func TestBinaryRejectsUnknownCommand(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	cmd := exec.Command(binary, "install")
	require.Error(t, cmd.Run())
}

func writeSourceFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

// gitRepo returns a temp dir that looks like a git repository root.
func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))
	return dir
}

// gitCloneSource returns a temp dir that is a real git repository with
// one commit, so clone can copy it.
func gitCloneSource(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "test")
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	writeSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-qm", "seed")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v\n%s", args, out)
}

func TestUpdateReadsFromDefaultSource(t *testing.T) {
	t.Chdir(gitRepo(t))
	t.Setenv("ZPECS_SOURCE", gitCloneSource(t))

	require.NoError(t, run([]string{"update"}))
	_, err := os.Stat(filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(".opencode", "agents", "code-architect.md"))
	require.NoError(t, err)
}

func TestUpdateReadsFromLocalSourceOverDefault(t *testing.T) {
	t.Chdir(gitRepo(t))
	t.Setenv("ZPECS_SOURCE", gitCloneSource(t))
	local := t.TempDir()
	writeSourceFile(t, local, filepath.Join("skills", "local-only", "SKILL.md"), "# local-only\n")

	require.NoError(t, run([]string{"update", "--source", local}))
	_, err := os.Stat(filepath.Join(".opencode", "skills", "local-only", "SKILL.md"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md"))
	require.Error(t, err)
}

func TestBinaryReadsFromDefaultSource(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	work := gitRepo(t)
	cmd := exec.Command(binary, "update")
	cmd.Dir = work
	cmd.Env = append(os.Environ(), "ZPECS_SOURCE="+gitCloneSource(t))
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zpecs update failed\n%s", out)
	_, err = os.Stat(filepath.Join(work, ".opencode", "skills", "prose-editor", "SKILL.md"))
	require.NoError(t, err)
}

func TestUpdateReadsSameFrontmatterForBothTargets(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), `---
name: prose-editor
description: Reviews prose.
tools:
  - read
  - edit
---

Review prose.
`)

	cases := []targetName{targetClaude, targetOpencode}
	for _, trgt := range cases {
		t.Run(string(trgt), func(t *testing.T) {
			err := run([]string{"update", "--source", dir, "--target", string(trgt)})
			require.NoError(t, err)
		})
	}
}

func TestUpdateWritesAgentUnderSourceNameForBothTargets(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), `---
name: renamed
description: Reviews prose.
---
Review prose.
`)

	cases := []targetName{targetClaude, targetOpencode}
	for _, trgt := range cases {
		t.Run(string(trgt), func(t *testing.T) {
			err := run([]string{"update", "--source", dir, "--target", string(trgt)})
			require.NoError(t, err)
			path := filepath.Join("."+string(trgt), "agents", "prose-editor.md")
			content, err := os.ReadFile(path)
			require.NoError(t, err)
			if trgt == targetClaude {
				require.Contains(t, string(content), "name: renamed")
			} else {
				require.NotContains(t, string(content), "name:")
			}
		})
	}
}

func TestUpdateWritesSkillAndAgentToClaude(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), `---
name: prose-editor
description: Reviews prose.
---
Review prose.
`)

	err := run([]string{"update", "--source", dir, "--target", "claude"})
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(".claude", "skills", "prose-editor", "SKILL.md"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(".claude", "agents", "prose-editor.md"))
	require.NoError(t, err)
}

func TestUpdateWritesSkillAndAgentToOpencode(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), `---
name: prose-editor
description: Reviews prose.
---
Review prose.
`)

	err := run([]string{"update", "--source", dir, "--target", "opencode"})
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(".opencode", "agents", "prose-editor.md"))
	require.NoError(t, err)
}

func TestBinaryWritesToClaudeTarget(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	work := gitRepo(t)
	cmd := exec.Command(binary, "update", "--source", dir, "--target", "claude")
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zpecs update failed\n%s", out)

	_, err = os.Stat(filepath.Join(work, ".claude", "skills", "prose-editor", "SKILL.md"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(work, ".claude", "agents", "prose-editor.md"))
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
		{name: "update renders skills, agents, and docs", wantSkill: true, wantAgent: true, wantDoc: true},
		{name: "update skills renders skills only", scope: "skills", wantSkill: true},
		{name: "update agents renders agents only", scope: "agents", wantAgent: true},
		{name: "update docs renders docs only", scope: "docs", wantDoc: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(gitRepo(t))
			dir := t.TempDir()
			writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
			writeSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")
			writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

			args := []string{"update"}
			if tc.scope != "" {
				args = append(args, tc.scope)
			}
			args = append(args, "--source", dir)
			require.NoError(t, run(args))

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

func TestBinaryScopedUpdateWritesOnlyWhatItNames(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	writeSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")

	for _, tc := range []struct {
		scope string
	}{
		{scope: "skills"},
		{scope: "agents"},
	} {
		t.Run(tc.scope, func(t *testing.T) {
			work := gitRepo(t)
			cmd := exec.Command(binary, "update", tc.scope, "--source", dir)
			cmd.Dir = work
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "zpecs update %s failed\n%s", tc.scope, out)

			skillPath := filepath.Join(work, ".opencode", "skills", "prose-editor", "SKILL.md")
			agentPath := filepath.Join(work, ".opencode", "agents", "code-architect.md")
			if tc.scope == "skills" {
				_, err := os.Stat(skillPath)
				require.NoError(t, err)
				_, err = os.Stat(agentPath)
				require.Error(t, err)
			} else {
				_, err := os.Stat(agentPath)
				require.NoError(t, err)
				_, err = os.Stat(skillPath)
				require.Error(t, err)
			}
		})
	}
}

func TestUpdateWritesToRepositoryRoot(t *testing.T) {
	root := gitRepo(t)
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	work := filepath.Join(root, "nested", "deep")
	require.NoError(t, os.MkdirAll(work, 0o755))
	t.Chdir(work)

	require.NoError(t, run([]string{"update", "--source", dir}))

	skillPath := filepath.Join(root, ".opencode", "skills", "prose-editor", "SKILL.md")
	_, err := os.Stat(skillPath)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(work, ".opencode", "skills", "prose-editor", "SKILL.md"))
	require.Error(t, err)
}

func TestUpdateErrorsOutsideRepository(t *testing.T) {
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	t.Chdir(t.TempDir())
	err := run([]string{"update", "--source", dir})
	require.Error(t, err)
	_, statErr := os.Stat(filepath.Join(".opencode", "agents", "prose-editor.md"))
	require.Error(t, statErr)
}

func TestBinaryWritesToRepositoryRoot(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	root := gitRepo(t)
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	work := filepath.Join(root, "nested")
	require.NoError(t, os.MkdirAll(work, 0o755))
	cmd := exec.Command(binary, "update", "--source", dir, "--target", "claude")
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zpecs update failed\n%s", out)

	agentPath := filepath.Join(root, ".claude", "agents", "prose-editor.md")
	_, err = os.Stat(agentPath)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(work, ".claude", "agents", "prose-editor.md"))
	require.Error(t, err)
}

func TestBinaryErrorsOutsideRepository(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	work := t.TempDir()
	cmd := exec.Command(binary, "update", "--source", dir)
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "expected failure outside a repository, got success\n%s", out)
	_, err = os.Stat(filepath.Join(work, ".opencode"))
	require.Error(t, err)
}

func TestUpdateCreatesMissingDirectories(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	require.NoError(t, run([]string{"update", "skills", "--source", dir, "--target", "claude"}))

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
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	path := filepath.Join(".claude", "agents", "prose-editor.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("manual content\n"), 0o644))

	require.NoError(t, run([]string{"update", "--source", dir, "--target", "claude"}))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "manual content\n", string(content))
}

func TestUpdateReplacesOwnedFiles(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nFirst.\n")

	require.NoError(t, run([]string{"update", "--source", dir}))

	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nSecond.\n")
	require.NoError(t, run([]string{"update", "--source", dir}))

	content, err := os.ReadFile(filepath.Join(".opencode", "agents", "prose-editor.md"))
	require.NoError(t, err)
	require.Contains(t, string(content), "Second.")
}

func TestBinaryLeavesForeignFileAlone(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	work := gitRepo(t)
	path := filepath.Join(work, ".claude", "agents", "prose-editor.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("manual content\n"), 0o644))

	cmd := exec.Command(binary, "update", "--source", dir, "--target", "claude")
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zpecs update failed\n%s", out)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "manual content\n", string(content))
}

func TestUpdateRemovesStaleSkill(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	require.NoError(t, run([]string{"update", "--source", dir}))
	path := filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")
	_, err := os.Stat(path)
	require.NoError(t, err)

	require.NoError(t, os.RemoveAll(filepath.Join(dir, "skills")))

	require.NoError(t, run([]string{"update", "--source", dir}))
	_, err = os.Stat(path)
	require.Error(t, err)
}

func TestUpdateRemovesStaleAgent(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	require.NoError(t, run([]string{"update", "--source", dir, "--target", "claude"}))
	path := filepath.Join(".claude", "agents", "prose-editor.md")
	_, err := os.Stat(path)
	require.NoError(t, err)

	require.NoError(t, os.RemoveAll(filepath.Join(dir, "agents")))

	require.NoError(t, run([]string{"update", "--source", dir, "--target", "claude"}))
	_, err = os.Stat(path)
	require.Error(t, err)
}

func TestUpdateAllWritesDocs(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	writeSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	require.NoError(t, run([]string{"update", "--source", dir}))

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
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	require.NoError(t, run([]string{"update", "--source", dir, "--target", "claude"}))

	_, err := os.Stat(filepath.Join(".claude", "agents", "code-architect.md"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join("docs", "zpecs", "prose.md"))
	require.NoError(t, err)
}

func TestUpdateDocsWritesFromSource(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	require.NoError(t, run([]string{"update", "docs", "--source", dir}))

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
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	require.NoError(t, run([]string{"update", "docs", "--source", dir, "--target", "claude"}))

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
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	path := filepath.Join("docs", "zpecs", "prose.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("manual content\n"), 0o644))

	require.NoError(t, run([]string{"update", "docs", "--source", dir}))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "manual content\n", string(content))
}

func TestUpdateDocsReplacesOwned(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	require.NoError(t, run([]string{"update", "docs", "--source", dir}))

	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture, second\n")
	require.NoError(t, run([]string{"update", "docs", "--source", dir}))

	content, err := os.ReadFile(filepath.Join("docs", "zpecs", "architecture.md"))
	require.NoError(t, err)
	require.Contains(t, string(content), "second")
}

func TestUpdateDocsRemovesStale(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	require.NoError(t, run([]string{"update", "docs", "--source", dir}))
	path := filepath.Join("docs", "zpecs", "architecture.md")
	_, err := os.Stat(path)
	require.NoError(t, err)

	require.NoError(t, os.RemoveAll(filepath.Join(dir, "docs")))

	require.NoError(t, run([]string{"update", "docs", "--source", dir}))
	_, err = os.Stat(path)
	require.Error(t, err)
}

func TestUpdateDocsWritesToRepositoryRoot(t *testing.T) {
	root := gitRepo(t)
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	work := filepath.Join(root, "nested", "deep")
	require.NoError(t, os.MkdirAll(work, 0o755))
	t.Chdir(work)

	require.NoError(t, run([]string{"update", "docs", "--source", dir}))

	path := filepath.Join(root, "docs", "zpecs", "architecture.md")
	_, err := os.Stat(path)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(work, "docs", "zpecs", "architecture.md"))
	require.Error(t, err)
}

func TestUpdateDocsErrorsOutsideRepository(t *testing.T) {
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	t.Chdir(t.TempDir())
	err := run([]string{"update", "docs", "--source", dir})
	require.Error(t, err)
	_, statErr := os.Stat(filepath.Join("docs", "zpecs", "architecture.md"))
	require.Error(t, statErr)
}

func TestUpdateScopedRemovalLeavesOtherKinds(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	writeSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")

	require.NoError(t, run([]string{"update", "--source", dir}))
	skillPath := filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")
	agentPath := filepath.Join(".opencode", "agents", "code-architect.md")
	_, err := os.Stat(skillPath)
	require.NoError(t, err)
	_, err = os.Stat(agentPath)
	require.NoError(t, err)

	require.NoError(t, os.RemoveAll(filepath.Join(dir, "skills")))

	require.NoError(t, run([]string{"update", "skills", "--source", dir}))
	_, err = os.Stat(skillPath)
	require.Error(t, err)
	_, err = os.Stat(agentPath)
	require.NoError(t, err)
}

func TestBinaryRemovesStaleSkill(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	work := gitRepo(t)
	cmd := exec.Command(binary, "update", "--source", dir)
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "first update failed\n%s", out)
	path := filepath.Join(work, ".opencode", "skills", "prose-editor", "SKILL.md")
	_, err = os.Stat(path)
	require.NoError(t, err)

	require.NoError(t, os.RemoveAll(filepath.Join(dir, "skills")))

	cmd = exec.Command(binary, "update", "--source", dir)
	cmd.Dir = work
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "second update failed\n%s", out)
	_, err = os.Stat(path)
	require.Error(t, err)
}

func TestConvertPrintsTheSpecAsJSON(t *testing.T) {
	path := filepath.Join("..", "..", "internal", "spec", "testdata", "convert.md")

	stdout := captureStdout(t)
	require.NoError(t, run([]string{"convert", path}))
	out := stdout()

	var doc spec.Document
	require.NoError(t, json.Unmarshal(out, &doc), "convert output is not one JSON object\n%s", out)
	require.NotEmpty(t, doc.Title)
	require.NotEmpty(t, doc.Purpose)
	require.NotEmpty(t, doc.Requirements)
}

func TestConvertErrorsOnMissingFile(t *testing.T) {
	stdout := captureStdout(t)
	err := run([]string{"convert", filepath.Join(t.TempDir(), "missing.md")})
	require.Error(t, err)
	out := stdout()
	require.Empty(t, out)
}

func TestConvertErrorsOnFileWithoutTopLevelHeading(t *testing.T) {
	path := filepath.Join("..", "..", "internal", "spec", "testdata", "no-title.md")

	require.Error(t, run([]string{"convert", path}))
}

func TestConvertErrorsOnRequirementWithoutName(t *testing.T) {
	path := filepath.Join("..", "..", "internal", "spec", "testdata", "requirement-without-name.md")

	require.Error(t, run([]string{"convert", path}))
}

func TestBinaryConvertPrintsTheSpecAsJSON(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	path, err := filepath.Abs(filepath.Join("..", "..", "internal", "spec", "testdata", "convert.md"))
	require.NoError(t, err)

	cmd := exec.Command(binary, "convert", path)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zpecs convert failed\n%s", out)

	var doc spec.Document
	require.NoError(t, json.Unmarshal(out, &doc), "convert output is not one JSON object\n%s", out)
	require.NotEmpty(t, doc.Title)
	require.NotEmpty(t, doc.Purpose)
	require.NotEmpty(t, doc.Requirements)
}

// captureStdout redirects os.Stdout to a pipe. The returned func reads
// everything written during the capture. The test restores os.Stdout
// when it finishes.
func captureStdout(t *testing.T) func() []byte {
	t.Helper()
	return capture(t, os.Stdout, func(f *os.File) { os.Stdout = f })
}

// captureStderr redirects os.Stderr to a pipe, mirroring captureStdout.
func captureStderr(t *testing.T) func() []byte {
	t.Helper()
	return capture(t, os.Stderr, func(f *os.File) { os.Stderr = f })
}

// capture redirects a process stream to a pipe. The returned func reads
// everything written during the capture. The test restores the stream
// when it finishes.
func capture(t *testing.T, old *os.File, set func(*os.File)) func() []byte {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	set(w)
	t.Cleanup(func() {
		set(old)
		_ = w.Close()
		_ = r.Close()
	})
	return func() []byte {
		t.Helper()
		_ = w.Close()
		out, err := io.ReadAll(r)
		require.NoError(t, err)
		return out
	}
}
