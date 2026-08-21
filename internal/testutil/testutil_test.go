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
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/targetdir"
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

func TestChdirIntoCreatesNestedDirAndEntersIt(t *testing.T) {
	root := t.TempDir()

	dir := ChdirInto(t, root, "nested", "deep")

	require.Equal(t, filepath.Join(root, "nested", "deep"), dir)
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, dir, wd)
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

func TestAgentSourceCreatesAgent(t *testing.T) {
	dir := AgentSource(t, "seed")

	content, err := os.ReadFile(filepath.Join(dir, "agents", "seed.md"))
	require.NoError(t, err)
	require.Equal(t, "---\nname: seed\n---\n\nseed.\n", string(content))
}

func TestWriteAgentBodyWritesTheBody(t *testing.T) {
	dir := t.TempDir()

	WriteAgentBody(t, dir, "seed", "body\n")

	content, err := os.ReadFile(filepath.Join(dir, "agents", "seed.md"))
	require.NoError(t, err)
	require.Contains(t, string(content), "body\n")
}

func TestDocSourceCreatesDoc(t *testing.T) {
	dir := DocSource(t, "seed")

	content, err := os.ReadFile(filepath.Join(dir, "docs", "zpecs", "seed.md"))
	require.NoError(t, err)
	require.Equal(t, "# seed\n", string(content))
}

func TestWriteDocBodyWritesTheContent(t *testing.T) {
	dir := t.TempDir()

	WriteDocBody(t, dir, "seed", "content\n")

	content, err := os.ReadFile(filepath.Join(dir, "docs", "zpecs", "seed.md"))
	require.NoError(t, err)
	require.Equal(t, "content\n", string(content))
}

func TestRequireWrittenPassesForWrittenFile(t *testing.T) {
	root := t.TempDir()
	rel := targetdir.RelPath(source.Opencode, source.Definition{Kind: source.Skill, Name: "seed"})
	writeFile(t, root, rel, "content\n")

	RequireWritten(t, root, source.Opencode, "seed", source.Skill)
}

func TestRequireNotWrittenPassesForMissingFile(t *testing.T) {
	root := t.TempDir()

	RequireNotWritten(t, root, source.Opencode, "seed", source.Skill)
}

func TestWrittenContentReturnsTheText(t *testing.T) {
	root := t.TempDir()
	rel := targetdir.RelPath(source.Opencode, source.Definition{Kind: source.Skill, Name: "seed"})
	writeFile(t, root, rel, "content\n")

	require.Equal(t, "content\n", WrittenContent(t, root, source.Opencode, "seed", source.Skill))
}

func TestSeedForeignFileWritesAtTargetPath(t *testing.T) {
	dir := t.TempDir()

	SeedForeignFile(t, dir, source.Claude, "seed", source.Agent, "manual content\n")

	rel := targetdir.RelPath(source.Claude, source.Definition{Kind: source.Agent, Name: "seed"})
	content, err := os.ReadFile(filepath.Join(dir, rel))
	require.NoError(t, err)
	require.Equal(t, "manual content\n", string(content))
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

func TestFakeOpenCodeReturnsEmptyBeforeInvocation(t *testing.T) {
	read := FakeOpenCode(t)

	require.Equal(t, "", read())
}

func TestFakeOpenCodeRecordsInvocation(t *testing.T) {
	read := FakeOpenCode(t)

	dir := t.TempDir()
	cmd := exec.Command("opencode", "run", "hello")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	record := read()
	require.Contains(t, record, "pwd: "+dir+"\n")
	require.Contains(t, record, "args: run hello\n")
}

func TestRanInFormatsTheRecordLine(t *testing.T) {
	dir := "/some/dir"

	require.Equal(t, "pwd: "+dir+"\n", RanIn(dir))
}

func TestRanAgainstFormatsTheRecordLine(t *testing.T) {
	require.Equal(t, "args: run --model deepseek/deepseek-v4-flash Review the repository against the guidelines in docs/zpecs/code.md.\n", RanAgainst("deepseek/deepseek-v4-flash", "docs/zpecs/code.md"))
}

func TestRanAgainstVariantFormatsTheRecordLine(t *testing.T) {
	require.Equal(t, "args: run --model deepseek/deepseek-v4-flash --variant minimal Review the repository against the guidelines in docs/zpecs/code.md.\n", RanAgainstVariant("deepseek/deepseek-v4-flash", "minimal", "docs/zpecs/code.md"))
}

func runGitErr(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	_, err := cmd.CombinedOutput()
	return err
}
