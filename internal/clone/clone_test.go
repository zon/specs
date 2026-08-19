package clone

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func gitRepo(t *testing.T, files ...string) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "test")
	for _, f := range files {
		path := filepath.Join(dir, f)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte("content\n"), 0o644))
	}
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

func TestCloneCopiesRepository(t *testing.T) {
	src := gitRepo(t, filepath.Join("skills", "seed", "SKILL.md"))

	dir, cleanup, err := Clone(src)
	require.NoError(t, err)
	defer cleanup()

	_, err = os.Stat(filepath.Join(dir, "skills", "seed", "SKILL.md"))
	require.NoError(t, err)
}

func TestCloneCleanupRemovesDirectory(t *testing.T) {
	src := gitRepo(t, filepath.Join("skills", "seed", "SKILL.md"))

	dir, cleanup, err := Clone(src)
	require.NoError(t, err)
	cleanup()

	_, err = os.Stat(dir)
	require.Error(t, err)
}
