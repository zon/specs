package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitRepoCreatesRepoWithFiles(t *testing.T) {
	dir := GitRepo(t, map[string]string{
		filepath.Join("skills", "seed", "SKILL.md"): "content\n",
	})

	require.DirExists(t, dir)

	content, err := os.ReadFile(filepath.Join(dir, "skills", "seed", "SKILL.md"))
	require.NoError(t, err)
	require.Equal(t, "content\n", string(content))

	RunGit(t, dir, "rev-parse", "HEAD")
}

func TestGitRepoWithoutFilesStillHasGitDir(t *testing.T) {
	dir := GitRepo(t, nil)

	require.DirExists(t, filepath.Join(dir, ".git"))
}

func TestGitRepoWithoutFilesHasNoCommit(t *testing.T) {
	dir := GitRepo(t, nil)

	err := runGitErr(dir, "rev-parse", "HEAD")
	require.Error(t, err)
}

func runGitErr(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	_, err := cmd.CombinedOutput()
	return err
}
