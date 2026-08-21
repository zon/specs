// Package testutil provides shared test fixtures.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/report"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/targetdir"
)

// runGit runs git with args in dir and fails the test on error.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v\n%s", args, out)
}

// GitRepo creates a temp git repository with the given files and returns its path.
func GitRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "test")
	for path, content := range files {
		writeFile(t, dir, path, content)
	}
	if len(files) > 0 {
		runGit(t, dir, "add", "-A")
		runGit(t, dir, "commit", "-qm", "seed")
	}
	return dir
}

// GitRepoURL returns the file:// URL of the repository GitRepo builds.
func GitRepoURL(t *testing.T, files map[string]string) string {
	t.Helper()
	return "file://" + GitRepo(t, files)
}

// ChdirInto creates parts as a nested directory under root, chdirs into
// it, and returns the path. Cleanup restores the working directory.
func ChdirInto(t *testing.T, root string, parts ...string) string {
	t.Helper()
	dir := filepath.Join(root, filepath.Join(parts...))
	require.NoError(t, os.MkdirAll(dir, 0o755))
	t.Chdir(dir)
	return dir
}

// writeFile writes content to rel under dir, creating parent directories.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
}

// writeSourceFile writes a definition file at rel under dir.
func writeSourceFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	writeFile(t, dir, rel, content)
}

// WriteSkill writes one skill at its layout path under dir.
func WriteSkill(t *testing.T, dir, name string) {
	t.Helper()
	writeSourceFile(t, dir, source.RelPath(source.Definition{Kind: source.Skill, Name: name}), "# "+name+"\n")
}

// WriteAgent writes one agent at its layout path under dir.
func WriteAgent(t *testing.T, dir, name string) {
	t.Helper()
	writeSourceFile(t, dir, source.RelPath(source.Definition{Kind: source.Agent, Name: name}), "---\nname: "+name+"\n---\n\n"+name+".\n")
}

// WriteAgentBody writes one agent at its layout path with the given body.
func WriteAgentBody(t *testing.T, dir, name, body string) {
	t.Helper()
	writeSourceFile(t, dir, source.RelPath(source.Definition{Kind: source.Agent, Name: name}), "---\nname: "+name+"\n---\n\n"+body)
}

// WriteDoc writes one doc at its layout path under dir.
func WriteDoc(t *testing.T, dir, name string) {
	t.Helper()
	writeSourceFile(t, dir, source.RelPath(source.Definition{Kind: source.Doc, Name: name}), "# "+name+"\n")
}

// WriteDocBody writes one doc at its layout path with the given content.
func WriteDocBody(t *testing.T, dir, name, content string) {
	t.Helper()
	writeSourceFile(t, dir, source.RelPath(source.Definition{Kind: source.Doc, Name: name}), content)
}

// SkillSource returns a temp dir with one skill.
func SkillSource(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	WriteSkill(t, dir, name)
	return dir
}

// AgentSource returns a temp dir with one agent.
func AgentSource(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	WriteAgent(t, dir, name)
	return dir
}

// DocSource returns a temp dir with one doc.
func DocSource(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	WriteDoc(t, dir, name)
	return dir
}

// RequireFile asserts a file exists at rel under dir.
func RequireFile(t *testing.T, dir, rel string) {
	t.Helper()
	require.FileExists(t, filepath.Join(dir, rel))
}

// RequireWritten asserts the definition has a written file under root for the target.
func RequireWritten(t *testing.T, root, targetName, name string, kind source.Kind) {
	t.Helper()
	rel := targetdir.RelPath(targetName, source.Definition{Kind: kind, Name: name})
	require.FileExists(t, filepath.Join(root, rel))
}

// RequireNotWritten asserts the definition has no written file under root for the target.
func RequireNotWritten(t *testing.T, root, targetName, name string, kind source.Kind) {
	t.Helper()
	rel := targetdir.RelPath(targetName, source.Definition{Kind: kind, Name: name})
	require.NoFileExists(t, filepath.Join(root, rel))
}

// WrittenContent returns the definition's written text under root for the target.
func WrittenContent(t *testing.T, root, targetName, name string, kind source.Kind) string {
	t.Helper()
	rel := targetdir.RelPath(targetName, source.Definition{Kind: kind, Name: name})
	content, err := os.ReadFile(filepath.Join(root, rel))
	require.NoError(t, err)
	return string(content)
}

// SeedForeignFile writes content by hand at the definition's path under root for the target, so the manifest does not record it.
func SeedForeignFile(t *testing.T, root, targetName, name string, kind source.Kind, content string) {
	t.Helper()
	rel := targetdir.RelPath(targetName, source.Definition{Kind: kind, Name: name})
	writeFile(t, root, rel, content)
}

// CaptureReport redirects report's output sink to a buffer. It returns
// everything written since the capture started. Captures must not run in
// parallel: report.Out is process-wide.
func CaptureReport(t *testing.T) func() string {
	t.Helper()
	var buf strings.Builder
	prev := report.Out
	report.Out = &buf
	t.Cleanup(func() { report.Out = prev })
	return buf.String
}

// FakeOpenCode installs a fake opencode executable on PATH that appends its
// working directory and arguments to a record file. The returned func
// returns the record, or "" when opencode has not run yet.
func FakeOpenCode(t *testing.T) func() string {
	t.Helper()
	dir := t.TempDir()
	record := filepath.Join(dir, "record")
	script := "#!/bin/sh\n" +
		"echo \"pwd: $(pwd)\" >> " + record + "\n" +
		"echo \"args: $*\" >> " + record + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "opencode"), []byte(script), 0o755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return func() string {
		content, err := os.ReadFile(record)
		if os.IsNotExist(err) {
			return ""
		}
		require.NoError(t, err)
		return string(content)
	}
}

// RanIn returns the record line the fake opencode writes for a run in dir.
func RanIn(dir string) string {
	return "pwd: " + dir + "\n"
}

// RanAgainst returns the args record line the fake opencode writes for a
// review with the model against the guidelines doc. It matches the prompt
// internal/opencode builds for the prose scope.
func RanAgainst(model, guidelines string) string {
	return "args: run --model " + model + " Review the repository against the guidelines in " + guidelines + ".\n"
}

// RanAgainstMessage returns the args record line the fake opencode writes
// for a review with the model and the given prompt message.
func RanAgainstMessage(model, message string) string {
	return "args: run --model " + model + " " + message + "\n"
}

// RanAgainstMessageVariant returns the args record line the fake opencode
// writes for a review with the model and variant and the given prompt
// message.
func RanAgainstMessageVariant(model, variant, message string) string {
	return "args: run --model " + model + " --variant " + variant + " " + message + "\n"
}
