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

// writeFile writes content to rel under dir, creating parent directories.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
}

// WriteSourceFile writes a definition file at rel under dir, creating parent directories.
func WriteSourceFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	writeFile(t, dir, rel, content)
}

// WriteSkill writes one skill at skills/<name>/SKILL.md under dir.
func WriteSkill(t *testing.T, dir, name string) {
	t.Helper()
	WriteSourceFile(t, dir, filepath.Join("skills", name, "SKILL.md"), "# "+name+"\n")
}

// WriteAgent writes one agent at agents/<name>.md under dir.
func WriteAgent(t *testing.T, dir, name string) {
	t.Helper()
	WriteSourceFile(t, dir, filepath.Join("agents", name+".md"), "---\nname: "+name+"\n---\n\n"+name+".\n")
}

// WriteAgentBody writes one agent at agents/<name>.md with the given body.
func WriteAgentBody(t *testing.T, dir, name, body string) {
	t.Helper()
	WriteSourceFile(t, dir, filepath.Join("agents", name+".md"), "---\nname: "+name+"\n---\n\n"+body)
}

// WriteDoc writes one doc at docs/zpecs/<name>.md under dir.
func WriteDoc(t *testing.T, dir, name string) {
	t.Helper()
	WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", name+".md"), "# "+name+"\n")
}

// WriteDocBody writes one doc at docs/zpecs/<name>.md with the given content.
func WriteDocBody(t *testing.T, dir, name, content string) {
	t.Helper()
	WriteSourceFile(t, dir, filepath.Join("docs", "zpecs", name+".md"), content)
}

// SkillSource returns a temp dir with one skill at skills/<name>/SKILL.md.
func SkillSource(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	WriteSkill(t, dir, name)
	return dir
}

// AgentSource returns a temp dir with one agent at agents/<name>.md.
func AgentSource(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	WriteAgent(t, dir, name)
	return dir
}

// DocSource returns a temp dir with one doc at docs/zpecs/<name>.md.
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
