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

func TestRunRecognizesCommands(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: nil},
		{name: "update", args: []string{"update"}},
		{name: "update skills", args: []string{"update", "skills"}},
		{name: "update agents", args: []string{"update", "agents"}},
		{name: "unknown command", args: []string{"install"}, wantErr: true},
		{name: "unknown scope", args: []string{"update", "docs"}, wantErr: true},
		{name: "too many arguments", args: []string{"update", "skills", "agents"}, wantErr: true},
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
	cases := []struct {
		args []string
		want string
	}{
		{args: []string{"update"}, want: "updating skills and agents"},
		{args: []string{"update", "skills"}, want: "updating skills"},
		{args: []string{"update", "agents"}, want: "updating agents"},
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
