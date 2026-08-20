// Package testutil builds temp git repositories and source trees for tests.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// RunGit runs git with args in dir and fails the test on error.
func RunGit(t *testing.T, dir string, args ...string) {
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
	RunGit(t, dir, "init", "-q")
	RunGit(t, dir, "config", "user.email", "test@example.com")
	RunGit(t, dir, "config", "user.name", "test")
	for path, content := range files {
		WriteSourceFile(t, dir, path, content)
	}
	if len(files) > 0 {
		RunGit(t, dir, "add", "-A")
		RunGit(t, dir, "commit", "-qm", "seed")
	}
	return dir
}

// GitRepoURL creates a temp git repository with the given files and
// returns its file:// URL.
func GitRepoURL(t *testing.T, files map[string]string) string {
	t.Helper()
	return "file://" + GitRepo(t, files)
}

// WriteSourceFile writes content to dir/rel, creating parent
// directories, and fails the test on error.
func WriteSourceFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}
