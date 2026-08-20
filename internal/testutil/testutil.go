// Package testutil provides shared test fixtures: temp git repositories,
// skill sources, and captured report output.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/report"
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

// SkillSource returns a temp dir with one skill at skills/<name>/SKILL.md.
func SkillSource(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("skills", name, "SKILL.md"), "# "+name+"\n")
	return dir
}

// RequireFile asserts a file exists at rel under dir.
func RequireFile(t *testing.T, dir, rel string) {
	t.Helper()
	require.FileExists(t, filepath.Join(dir, rel))
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
