package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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

func TestScopeFromArgs(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    scope
		wantErr bool
	}{
		{name: "no scope", args: nil, want: scopeAll},
		{name: "empty scope", args: []string{}, want: scopeAll},
		{name: "skills", args: []string{"skills"}, want: scopeSkills},
		{name: "agents", args: []string{"agents"}, want: scopeAgents},
		{name: "unknown scope", args: []string{"docs"}, wantErr: true},
		{name: "too many scopes", args: []string{"skills", "agents"}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := scopeFromArgs(tc.args)
			if (err != nil) != tc.wantErr {
				t.Fatalf("scopeFromArgs(%v) error = %v, wantErr %v", tc.args, err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Fatalf("scopeFromArgs(%v) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

func TestParseTarget(t *testing.T) {
	cases := []struct {
		name    string
		s       string
		want    target
		wantErr bool
	}{
		{name: "claude", s: "claude", want: targetClaude},
		{name: "opencode", s: "opencode", want: targetOpencode},
		{name: "unknown target", s: "vscode", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseTarget(tc.s)
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseTarget(%q) error = %v, wantErr %v", tc.s, err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Fatalf("parseTarget(%q) = %v, want %v", tc.s, got, tc.want)
			}
		})
	}
}

func TestParseOptions(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    options
		wantErr bool
	}{
		{name: "defaults", args: nil, want: options{scope: scopeAll, target: targetOpencode}},
		{name: "scope", args: []string{"skills"}, want: options{scope: scopeSkills, target: targetOpencode}},
		{name: "claude target", args: []string{"--target", "claude"}, want: options{scope: scopeAll, target: targetClaude}},
		{name: "claude target equals", args: []string{"--target=claude"}, want: options{scope: scopeAll, target: targetClaude}},
		{name: "source", args: []string{"--source", "/tmp/src"}, want: options{scope: scopeAll, target: targetOpencode, source: "/tmp/src"}},
		{name: "source equals", args: []string{"--source=/tmp/src"}, want: options{scope: scopeAll, target: targetOpencode, source: "/tmp/src"}},
		{name: "all together", args: []string{"agents", "--source", "/tmp/src", "--target", "claude"}, want: options{scope: scopeAgents, target: targetClaude, source: "/tmp/src"}},
		{name: "unknown target", args: []string{"--target", "vscode"}, wantErr: true},
		{name: "missing target value", args: []string{"--target"}, wantErr: true},
		{name: "missing source value", args: []string{"--source"}, wantErr: true},
		{name: "unknown flag", args: []string{"--force"}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseOptions(tc.args)
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseOptions(%v) error = %v, wantErr %v", tc.args, err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Fatalf("parseOptions(%v) = %+v, want %+v", tc.args, got, tc.want)
			}
		})
	}
}

func TestRunRecognizesCommands(t *testing.T) {
	t.Chdir(gitRepo(t))
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
		{name: "unknown scope", args: []string{"update", "docs"}, wantErr: true},
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

func TestBinaryRunsEachUpdateCommand(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	repoDir := gitRepo(t)
	sourceDir := t.TempDir()
	cases := []struct {
		args []string
		want string
	}{
		{args: []string{"update"}, want: "updating skills and agents for opencode"},
		{args: []string{"update", "skills"}, want: "updating skills for opencode"},
		{args: []string{"update", "agents"}, want: "updating agents for opencode"},
		{args: []string{"update", "--target", "claude"}, want: "updating skills and agents for claude"},
		{args: []string{"update", "--source", sourceDir, "skills"}, want: "updating skills for opencode from " + sourceDir},
	}

	for _, tc := range cases {
		cmd := exec.Command(binary, tc.args...)
		cmd.Dir = repoDir
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
	}{
		{name: "update renders skills and agents", wantSkill: true, wantAgent: true},
		{name: "update skills renders skills only", scope: "skills", wantSkill: true},
		{name: "update agents renders agents only", scope: "agents", wantAgent: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(gitRepo(t))
			dir := t.TempDir()
			writeSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
			writeSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")

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
