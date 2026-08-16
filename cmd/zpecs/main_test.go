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
		out, err := exec.Command(binary, tc.args...).CombinedOutput()
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

func TestUpdateReadsSameFrontmatterForBothTargets(t *testing.T) {
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
