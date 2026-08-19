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
	"github.com/zon/specs/internal/spec"
)

func buildBinary(t *testing.T, dir string) string {
	t.Helper()
	binary := filepath.Join(dir, "zpecs")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	return binary
}

func TestBuildProducesRunnableCLIBinary(t *testing.T) {
	binary := buildBinary(t, t.TempDir())

	info, err := os.Stat(binary)
	if err != nil {
		t.Fatalf("binary not produced: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("binary is not executable: %v", info.Mode())
	}

	cmd := exec.Command(binary)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("binary failed to run: %v\n%s", err, out)
	}
}

func TestBinaryPrintsVersion(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	out, err := exec.Command(binary, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("zpecs --version failed: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); got != version {
		t.Fatalf("zpecs --version = %q, want %q", got, version)
	}
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
			if (err != nil) != tc.wantErr {
				t.Fatalf("UnmarshalText(%q) error = %v, wantErr %v", tc.s, err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Fatalf("UnmarshalText(%q) = %v, want %v", tc.s, got, tc.want)
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
	if err != nil {
		t.Fatalf("kong.New: %v", err)
	}
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
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseUpdateArgs(%v) error = %v, wantErr %v", tc.args, err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Fatalf("parseUpdateArgs(%v) = %+v, want %+v", tc.args, got, tc.want)
			}
		})
	}
}

func TestParseUpdateEnvOverridesDefault(t *testing.T) {
	t.Setenv("ZPECS_SOURCE", "/env/src")
	got, err := parseUpdateArgs(t)
	if err != nil {
		t.Fatalf("parseUpdateArgs: %v", err)
	}
	want := options{scope: scopeAll, target: targetOpencode, source: "/env/src"}
	if got != want {
		t.Fatalf("parseUpdateArgs() = %+v, want %+v", got, want)
	}
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
			if (err != nil) != tc.wantErr {
				t.Fatalf("run(%v) error = %v, wantErr %v", tc.args, err, tc.wantErr)
			}
		})
	}
}

func TestPrintErrorShowsUsageForParseErrors(t *testing.T) {
	var c cli
	parser, err := kong.New(&c, cliVars)
	if err != nil {
		t.Fatalf("kong.New: %v", err)
	}
	_, parseErr := parser.Parse([]string{"update", "--target", "vscode"})
	if parseErr == nil {
		t.Fatal("expected a parse error")
	}

	stderr := captureStderr(t)
	printError(parseErr)
	if out := string(stderr()); !strings.Contains(out, "Usage:") {
		t.Fatalf("parse error did not print usage:\n%s", out)
	}
}

func TestPrintErrorOmitsUsageForRuntimeErrors(t *testing.T) {
	stderr := captureStderr(t)
	printError(errors.New("update outside a repository"))
	out := string(stderr())
	if strings.Contains(out, "Usage:") {
		t.Fatalf("runtime error printed usage:\n%s", out)
	}
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
		if err != nil {
			t.Fatalf("zpecs %v failed: %v", tc.args, err)
		}
		if !strings.Contains(string(out), tc.want) {
			t.Fatalf("zpecs %v output %q missing %q", tc.args, out, tc.want)
		}
	}
}

func TestBinaryRejectsUnknownCommand(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	cmd := exec.Command(binary, "install")
	if err := cmd.Run(); err == nil {
		t.Fatal("expected unknown command to fail")
	}
}

func writeSourceFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// gitRepo returns a temp dir that looks like a git repository root.
func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
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
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestUpdateReadsFromDefaultSource(t *testing.T) {
	t.Chdir(gitRepo(t))
	t.Setenv("ZPECS_SOURCE", gitCloneSource(t))

	if err := run([]string{"update"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")); err != nil {
		t.Fatalf("skill not written from the default source: %v", err)
	}
	if _, err := os.Stat(filepath.Join(".opencode", "agents", "code-architect.md")); err != nil {
		t.Fatalf("agent not written from the default source: %v", err)
	}
}

func TestUpdateReadsFromLocalSourceOverDefault(t *testing.T) {
	t.Chdir(gitRepo(t))
	t.Setenv("ZPECS_SOURCE", gitCloneSource(t))
	local := t.TempDir()
	writeSourceFile(t, local, filepath.Join("skills", "local-only", "SKILL.md"), "# local-only\n")

	if err := run([]string{"update", "--source", local}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(".opencode", "skills", "local-only", "SKILL.md")); err != nil {
		t.Fatalf("skill not read from the local source: %v", err)
	}
	if _, err := os.Stat(filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")); err == nil {
		t.Fatal("skill read from the default source despite a --source flag")
	}
}

func TestBinaryReadsFromDefaultSource(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	work := gitRepo(t)
	cmd := exec.Command(binary, "update")
	cmd.Dir = work
	cmd.Env = append(os.Environ(), "ZPECS_SOURCE="+gitCloneSource(t))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("zpecs update failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(work, ".opencode", "skills", "prose-editor", "SKILL.md")); err != nil {
		t.Fatalf("skill not written from the default source: %v", err)
	}
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

	cases := []target{targetClaude, targetOpencode}
	for _, trgt := range cases {
		t.Run(string(trgt), func(t *testing.T) {
			err := run([]string{"update", "--source", dir, "--target", string(trgt)})
			if err != nil {
				t.Fatalf("update for %s: %v", trgt, err)
			}
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

	cases := []target{targetClaude, targetOpencode}
	for _, trgt := range cases {
		t.Run(string(trgt), func(t *testing.T) {
			err := run([]string{"update", "--source", dir, "--target", string(trgt)})
			if err != nil {
				t.Fatalf("update for %s: %v", trgt, err)
			}
			path := filepath.Join("."+string(trgt), "agents", "prose-editor.md")
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%s wrote nothing at %s: %v", trgt, path, err)
			}
			if trgt == targetClaude {
				if !strings.Contains(string(content), "name: renamed") {
					t.Fatalf("%s wrote %q without the rendered name field", trgt, content)
				}
			} else if strings.Contains(string(content), "name:") {
				t.Fatalf("%s wrote %q with a name field", trgt, content)
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
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if _, err := os.Stat(filepath.Join(".claude", "skills", "prose-editor", "SKILL.md")); err != nil {
		t.Fatalf("claude skill not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(".claude", "agents", "prose-editor.md")); err != nil {
		t.Fatalf("claude agent not written: %v", err)
	}
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
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if _, err := os.Stat(filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")); err != nil {
		t.Fatalf("opencode skill not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(".opencode", "agents", "prose-editor.md")); err != nil {
		t.Fatalf("opencode agent not written: %v", err)
	}
}

func TestBinaryWritesToClaudeTarget(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	work := gitRepo(t)
	cmd := exec.Command(binary, "update", "--source", dir, "--target", "claude")
	cmd.Dir = work
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("zpecs update failed: %v\n%s", err, out)
	}

	if _, err := os.Stat(filepath.Join(work, ".claude", "skills", "prose-editor", "SKILL.md")); err != nil {
		t.Fatalf("claude skill not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(work, ".claude", "agents", "prose-editor.md")); err != nil {
		t.Fatalf("claude agent not written: %v", err)
	}
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
			if err := run(args); err != nil {
				t.Fatalf("run(%v): %v", args, err)
			}

			skillPath := filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")
			agentPath := filepath.Join(".opencode", "agents", "code-architect.md")
			docPath := filepath.Join("docs", "zpecs", "prose.md")
			if tc.wantSkill {
				if _, err := os.Stat(skillPath); err != nil {
					t.Fatalf("skill not rendered at %s: %v", skillPath, err)
				}
			} else if _, err := os.Stat(skillPath); err == nil {
				t.Fatalf("skill rendered at %s but the command does not name it", skillPath)
			}
			if tc.wantAgent {
				if _, err := os.Stat(agentPath); err != nil {
					t.Fatalf("agent not rendered at %s: %v", agentPath, err)
				}
			} else if _, err := os.Stat(agentPath); err == nil {
				t.Fatalf("agent rendered at %s but the command does not name it", agentPath)
			}
			if tc.wantDoc {
				if _, err := os.Stat(docPath); err != nil {
					t.Fatalf("doc not rendered at %s: %v", docPath, err)
				}
			} else if _, err := os.Stat(docPath); err == nil {
				t.Fatalf("doc rendered at %s but the command does not name it", docPath)
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
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("zpecs update %s failed: %v\n%s", tc.scope, err, out)
			}

			skillPath := filepath.Join(work, ".opencode", "skills", "prose-editor", "SKILL.md")
			agentPath := filepath.Join(work, ".opencode", "agents", "code-architect.md")
			if tc.scope == "skills" {
				if _, err := os.Stat(skillPath); err != nil {
					t.Fatalf("skill not written: %v", err)
				}
				if _, err := os.Stat(agentPath); err == nil {
					t.Fatal("agent written by update skills")
				}
			} else {
				if _, err := os.Stat(agentPath); err != nil {
					t.Fatalf("agent not written: %v", err)
				}
				if _, err := os.Stat(skillPath); err == nil {
					t.Fatal("skill written by update agents")
				}
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
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(work)

	if err := run([]string{"update", "--source", dir}); err != nil {
		t.Fatalf("update: %v", err)
	}

	skillPath := filepath.Join(root, ".opencode", "skills", "prose-editor", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("skill not written at the repository root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(work, ".opencode", "skills", "prose-editor", "SKILL.md")); err == nil {
		t.Fatal("skill written in the working subdirectory")
	}
}

func TestUpdateErrorsOutsideRepository(t *testing.T) {
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	t.Chdir(t.TempDir())
	err := run([]string{"update", "--source", dir})
	if err == nil {
		t.Fatal("expected update outside a repository to error")
	}
	if _, statErr := os.Stat(filepath.Join(".opencode", "agents", "prose-editor.md")); statErr == nil {
		t.Fatal("update outside a repository wrote a file")
	}
}

func TestBinaryWritesToRepositoryRoot(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	root := gitRepo(t)
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	work := filepath.Join(root, "nested")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary, "update", "--source", dir, "--target", "claude")
	cmd.Dir = work
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("zpecs update failed: %v\n%s", err, out)
	}

	agentPath := filepath.Join(root, ".claude", "agents", "prose-editor.md")
	if _, err := os.Stat(agentPath); err != nil {
		t.Fatalf("agent not written at the repository root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(work, ".claude", "agents", "prose-editor.md")); err == nil {
		t.Fatal("agent written in the working subdirectory")
	}
}

func TestBinaryErrorsOutsideRepository(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	work := t.TempDir()
	cmd := exec.Command(binary, "update", "--source", dir)
	cmd.Dir = work
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("expected failure outside a repository, got success\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(work, ".opencode")); err == nil {
		t.Fatal("binary wrote outside a repository")
	}
}

func TestUpdateCreatesMissingDirectories(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	if err := run([]string{"update", "skills", "--source", dir, "--target", "claude"}); err != nil {
		t.Fatalf("update: %v", err)
	}

	for _, path := range []string{
		filepath.Join(".claude"),
		filepath.Join(".claude", "skills"),
		filepath.Join(".claude", "skills", "prose-editor"),
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("directory %s not created: %v", path, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", path)
		}
	}
}

func TestUpdateLeavesForeignFileAlone(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	path := filepath.Join(".claude", "agents", "prose-editor.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("manual content\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{"update", "--source", dir, "--target", "claude"}); err != nil {
		t.Fatalf("update: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("foreign file: %v", err)
	}
	if string(content) != "manual content\n" {
		t.Fatalf("foreign file changed to %q", content)
	}
}

func TestUpdateReplacesOwnedFiles(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nFirst.\n")

	if err := run([]string{"update", "--source", dir}); err != nil {
		t.Fatalf("first update: %v", err)
	}

	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nSecond.\n")
	if err := run([]string{"update", "--source", dir}); err != nil {
		t.Fatalf("second update: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(".opencode", "agents", "prose-editor.md"))
	if err != nil {
		t.Fatalf("owned file: %v", err)
	}
	if !strings.Contains(string(content), "Second.") {
		t.Fatalf("owned file not replaced: %q", content)
	}
}

func TestBinaryLeavesForeignFileAlone(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	work := gitRepo(t)
	path := filepath.Join(work, ".claude", "agents", "prose-editor.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("manual content\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binary, "update", "--source", dir, "--target", "claude")
	cmd.Dir = work
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("zpecs update failed: %v\n%s", err, out)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("foreign file: %v", err)
	}
	if string(content) != "manual content\n" {
		t.Fatalf("foreign file changed to %q", content)
	}
}

func TestUpdateRemovesStaleSkill(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	if err := run([]string{"update", "--source", dir}); err != nil {
		t.Fatalf("first update: %v", err)
	}
	path := filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("skill not written: %v", err)
	}

	if err := os.RemoveAll(filepath.Join(dir, "skills")); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{"update", "--source", dir}); err != nil {
		t.Fatalf("second update: %v", err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("stale skill still present after the source stopped listing it")
	}
}

func TestUpdateRemovesStaleAgent(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	if err := run([]string{"update", "--source", dir, "--target", "claude"}); err != nil {
		t.Fatalf("first update: %v", err)
	}
	path := filepath.Join(".claude", "agents", "prose-editor.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("agent not written: %v", err)
	}

	if err := os.RemoveAll(filepath.Join(dir, "agents")); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{"update", "--source", dir, "--target", "claude"}); err != nil {
		t.Fatalf("second update: %v", err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("stale agent still present after the source stopped listing it")
	}
}

func TestUpdateAllWritesDocs(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	writeSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	if err := run([]string{"update", "--source", dir}); err != nil {
		t.Fatalf("update: %v", err)
	}

	for _, path := range []string{
		filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md"),
		filepath.Join(".opencode", "agents", "code-architect.md"),
		filepath.Join("docs", "zpecs", "prose.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s not written: %v", path, err)
		}
	}
}

func TestUpdateAllWritesDocsToTheTargetItNames(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	if err := run([]string{"update", "--source", dir, "--target", "claude"}); err != nil {
		t.Fatalf("update: %v", err)
	}

	if _, err := os.Stat(filepath.Join(".claude", "agents", "code-architect.md")); err != nil {
		t.Fatalf("claude agent not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join("docs", "zpecs", "prose.md")); err != nil {
		t.Fatalf("doc not written: %v", err)
	}
}

func TestUpdateDocsWritesFromSource(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	if err := run([]string{"update", "docs", "--source", dir}); err != nil {
		t.Fatalf("update: %v", err)
	}

	path := filepath.Join("docs", "zpecs", "architecture.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("doc not written: %v", err)
	}
	if string(content) != "# Architecture\n" {
		t.Fatalf("doc content = %q, want %q", content, "# Architecture\n")
	}
	if _, err := os.Stat(filepath.Join(".opencode")); err == nil {
		t.Fatal("update docs touched the opencode target")
	}
	if _, err := os.Stat(filepath.Join(".claude")); err == nil {
		t.Fatal("update docs touched the claude target")
	}
}

func TestUpdateDocsIgnoresTarget(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	if err := run([]string{"update", "docs", "--source", dir, "--target", "claude"}); err != nil {
		t.Fatalf("update: %v", err)
	}

	path := filepath.Join("docs", "zpecs", "prose.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("doc not written: %v", err)
	}
	if string(content) != "# Prose guidelines\n" {
		t.Fatalf("doc content = %q, want %q", content, "# Prose guidelines\n")
	}
	if _, err := os.Stat(filepath.Join(".claude")); err == nil {
		t.Fatal("update docs wrote to the claude target")
	}
	if _, err := os.Stat(filepath.Join(".opencode")); err == nil {
		t.Fatal("update docs wrote to the opencode target")
	}
}

func TestUpdateDocsLeavesForeignFileAlone(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	path := filepath.Join("docs", "zpecs", "prose.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("manual content\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{"update", "docs", "--source", dir}); err != nil {
		t.Fatalf("update: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("foreign file: %v", err)
	}
	if string(content) != "manual content\n" {
		t.Fatalf("foreign file changed to %q", content)
	}
}

func TestUpdateDocsReplacesOwned(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	if err := run([]string{"update", "docs", "--source", dir}); err != nil {
		t.Fatalf("first update: %v", err)
	}

	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture, second\n")
	if err := run([]string{"update", "docs", "--source", dir}); err != nil {
		t.Fatalf("second update: %v", err)
	}

	content, err := os.ReadFile(filepath.Join("docs", "zpecs", "architecture.md"))
	if err != nil {
		t.Fatalf("owned file: %v", err)
	}
	if !strings.Contains(string(content), "second") {
		t.Fatalf("owned file not replaced: %q", content)
	}
}

func TestUpdateDocsRemovesStale(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	if err := run([]string{"update", "docs", "--source", dir}); err != nil {
		t.Fatalf("first update: %v", err)
	}
	path := filepath.Join("docs", "zpecs", "architecture.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("doc not written: %v", err)
	}

	if err := os.RemoveAll(filepath.Join(dir, "docs")); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{"update", "docs", "--source", dir}); err != nil {
		t.Fatalf("second update: %v", err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("stale doc still present after the source stopped listing it")
	}
}

func TestUpdateDocsWritesToRepositoryRoot(t *testing.T) {
	root := gitRepo(t)
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	work := filepath.Join(root, "nested", "deep")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(work)

	if err := run([]string{"update", "docs", "--source", dir}); err != nil {
		t.Fatalf("update: %v", err)
	}

	path := filepath.Join(root, "docs", "zpecs", "architecture.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("doc not written at the repository root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(work, "docs", "zpecs", "architecture.md")); err == nil {
		t.Fatal("doc written in the working subdirectory")
	}
}

func TestUpdateDocsErrorsOutsideRepository(t *testing.T) {
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("docs", "zpecs", "architecture.md"), "# Architecture\n")

	t.Chdir(t.TempDir())
	err := run([]string{"update", "docs", "--source", dir})
	if err == nil {
		t.Fatal("expected update outside a repository to error")
	}
	if _, statErr := os.Stat(filepath.Join("docs", "zpecs", "architecture.md")); statErr == nil {
		t.Fatal("update outside a repository wrote a file")
	}
}

func TestUpdateScopedRemovalLeavesOtherKinds(t *testing.T) {
	t.Chdir(gitRepo(t))
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	writeSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")

	if err := run([]string{"update", "--source", dir}); err != nil {
		t.Fatalf("first update: %v", err)
	}
	skillPath := filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md")
	agentPath := filepath.Join(".opencode", "agents", "code-architect.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("skill not written: %v", err)
	}
	if _, err := os.Stat(agentPath); err != nil {
		t.Fatalf("agent not written: %v", err)
	}

	if err := os.RemoveAll(filepath.Join(dir, "skills")); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{"update", "skills", "--source", dir}); err != nil {
		t.Fatalf("update skills: %v", err)
	}
	if _, err := os.Stat(skillPath); err == nil {
		t.Fatal("stale skill still present")
	}
	if _, err := os.Stat(agentPath); err != nil {
		t.Fatalf("agent removed by update skills: %v", err)
	}
}

func TestBinaryRemovesStaleSkill(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	work := gitRepo(t)
	cmd := exec.Command(binary, "update", "--source", dir)
	cmd.Dir = work
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("first update failed: %v\n%s", err, out)
	}
	path := filepath.Join(work, ".opencode", "skills", "prose-editor", "SKILL.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("skill not written: %v", err)
	}

	if err := os.RemoveAll(filepath.Join(dir, "skills")); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command(binary, "update", "--source", dir)
	cmd.Dir = work
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("second update failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("stale skill still present after the binary ran")
	}
}

func TestConvertPrintsTheSpecAsJSON(t *testing.T) {
	path := filepath.Join("..", "..", "internal", "spec", "testdata", "convert.md")

	stdout := captureStdout(t)
	if err := run([]string{"convert", path}); err != nil {
		t.Fatalf("convert: %v", err)
	}
	out := stdout()

	var doc spec.Document
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("convert output is not one JSON object: %v\n%s", err, out)
	}
	if doc.Title == "" {
		t.Fatal("title is empty")
	}
	if doc.Purpose == "" {
		t.Fatal("purpose is empty")
	}
	if len(doc.Requirements) == 0 {
		t.Fatal("no requirements")
	}
}

func TestConvertErrorsOnMissingFile(t *testing.T) {
	stdout := captureStdout(t)
	err := run([]string{"convert", filepath.Join(t.TempDir(), "missing.md")})
	if err == nil {
		t.Fatal("convert on a missing file should error")
	}
	if out := stdout(); len(out) != 0 {
		t.Fatalf("convert printed %q on a missing file", out)
	}
}

func TestConvertErrorsOnFileWithoutTopLevelHeading(t *testing.T) {
	path := filepath.Join("..", "..", "internal", "spec", "testdata", "no-title.md")

	if err := run([]string{"convert", path}); err == nil {
		t.Fatal("convert on a file without a top-level heading should error")
	}
}

func TestConvertErrorsOnRequirementWithoutName(t *testing.T) {
	path := filepath.Join("..", "..", "internal", "spec", "testdata", "requirement-without-name.md")

	if err := run([]string{"convert", path}); err == nil {
		t.Fatal("convert on a requirement without a name should error")
	}
}

func TestBinaryConvertPrintsTheSpecAsJSON(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	path, err := filepath.Abs(filepath.Join("..", "..", "internal", "spec", "testdata", "convert.md"))
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}

	cmd := exec.Command(binary, "convert", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("zpecs convert failed: %v\n%s", err, out)
	}

	var doc spec.Document
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("convert output is not one JSON object: %v\n%s", err, out)
	}
	if doc.Title == "" {
		t.Fatal("title is empty")
	}
	if doc.Purpose == "" {
		t.Fatal("purpose is empty")
	}
	if len(doc.Requirements) == 0 {
		t.Fatal("no requirements")
	}
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
	if err != nil {
		t.Fatal(err)
	}
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
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
}
