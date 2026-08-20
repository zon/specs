package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestGitRepoURLReturnsCloneableURL(t *testing.T) {
	url := GitRepoURL(t, map[string]string{"seed": "content\n"})

	require.True(t, strings.HasPrefix(url, "file://"))

	dir := t.TempDir()
	RunGit(t, dir, "clone", url, filepath.Join(dir, "clone"))

	content, err := os.ReadFile(filepath.Join(dir, "clone", "seed"))
	require.NoError(t, err)
	require.Equal(t, "content\n", string(content))
}

func TestWriteSourceFileCreatesParentDirectories(t *testing.T) {
	dir := t.TempDir()
	rel := filepath.Join("skills", "prose-editor", "SKILL.md")

	WriteSourceFile(t, dir, rel, "# prose-editor\n")

	content, err := os.ReadFile(filepath.Join(dir, rel))
	require.NoError(t, err)
	require.Equal(t, "# prose-editor\n", string(content))
}

func runGitErr(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	_, err := cmd.CombinedOutput()
	return err
}
