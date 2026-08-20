package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/report"
)

func TestGitRepoCreatesRepoWithFiles(t *testing.T) {
	dir := GitRepo(t, map[string]string{
		filepath.Join("skills", "seed", "SKILL.md"): "content\n",
	})

	require.DirExists(t, dir)

	content, err := os.ReadFile(filepath.Join(dir, "skills", "seed", "SKILL.md"))
	require.NoError(t, err)
	require.Equal(t, "content\n", string(content))

	runGit(t, dir, "rev-parse", "HEAD")
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
	runGit(t, dir, "clone", url, filepath.Join(dir, "clone"))

	content, err := os.ReadFile(filepath.Join(dir, "clone", "seed"))
	require.NoError(t, err)
	require.Equal(t, "content\n", string(content))
}

func TestWriteFileCreatesFileAndDirectories(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, filepath.Join("a", "b", "seed"), "content\n")

	content, err := os.ReadFile(filepath.Join(dir, "a", "b", "seed"))
	require.NoError(t, err)
	require.Equal(t, "content\n", string(content))
}

func TestSkillSourceCreatesSkill(t *testing.T) {
	dir := SkillSource(t, "seed")

	content, err := os.ReadFile(filepath.Join(dir, "skills", "seed", "SKILL.md"))
	require.NoError(t, err)
	require.Equal(t, "# seed\n", string(content))
}

func TestRequireFilePassesForExistingFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "seed", "content\n")

	RequireFile(t, dir, "seed")
}

func TestCaptureReportCaptures(t *testing.T) {
	captured := CaptureReport(t)

	fmt.Fprint(report.Out, "hello\n")

	require.Equal(t, "hello\n", captured())
}

func runGitErr(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	_, err := cmd.CombinedOutput()
	return err
}
