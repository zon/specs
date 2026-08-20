package cli

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/spec"
	"github.com/zon/specs/internal/target"
	"github.com/zon/specs/internal/testutil"
)

func buildBinary(t *testing.T, dir string) string {
	t.Helper()
	binary := filepath.Join(dir, "zpecs")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, "github.com/zon/specs/cmd/zpecs")
	out, err := build.CombinedOutput()
	require.NoError(t, err, "go build failed\n%s", out)
	return binary
}

func TestBuildProducesRunnableCLIBinary(t *testing.T) {
	binary := buildBinary(t, t.TempDir())

	info, err := os.Stat(binary)
	require.NoError(t, err)
	require.False(t, runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0, "binary is not executable: %v", info.Mode())

	cmd := exec.Command(binary)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "binary failed to run\n%s", out)
}

func TestBinaryPrintsVersion(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	out, err := exec.Command(binary, "--version").CombinedOutput()
	require.NoError(t, err, "zpecs --version failed\n%s", out)
	require.Equal(t, version, strings.TrimSpace(string(out)))
}

func TestUnmarshalScope(t *testing.T) {
	cases := []struct {
		name    string
		s       string
		want    scope
		wantErr bool
	}{
		{name: "all", s: "all", want: scopeAll},
		{name: "skills", s: "skills", want: scopeSkills},
		{name: "agents", s: "agents", want: scopeAgents},
		{name: "docs", s: "docs", want: scopeDocs},
		{name: "unknown scope", s: "vscode", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got scope
			err := got.UnmarshalText([]byte(tc.s))
			require.Equal(t, tc.wantErr, err != nil)
			if err == nil {
				require.Equal(t, tc.want, got)
			}
		})
	}
}

// options is the parsed input parseUpdateArgs returns.
type options struct {
	scope  scope
	source string
	target string
}

// parseUpdateArgs parses `update <args>` with the same kong grammar
// `run` uses. It returns the selected options.
func parseUpdateArgs(t *testing.T, args ...string) (options, error) {
	t.Helper()
	var c cli
	parser, err := kong.New(&c, cliVars)
	require.NoError(t, err)
	if _, err := parser.Parse(append([]string{"update"}, args...)); err != nil {
		return options{}, err
	}
	return options{scope: c.Update.Scope, source: c.Update.Source, target: c.Update.Target}, nil
}

func TestParseUpdate(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    options
		wantErr bool
	}{
		{name: "defaults", args: nil, want: options{scope: scopeAll, target: target.Opencode, source: defaultSourceURL}},
		{name: "all scope", args: []string{"all"}, want: options{scope: scopeAll, target: target.Opencode, source: defaultSourceURL}},
		{name: "scope", args: []string{"skills"}, want: options{scope: scopeSkills, target: target.Opencode, source: defaultSourceURL}},
		{name: "docs scope", args: []string{"docs"}, want: options{scope: scopeDocs, target: target.Opencode, source: defaultSourceURL}},
		{name: "claude target", args: []string{"--target", "claude"}, want: options{scope: scopeAll, target: target.Claude, source: defaultSourceURL}},
		{name: "claude target equals", args: []string{"--target=claude"}, want: options{scope: scopeAll, target: target.Claude, source: defaultSourceURL}},
		{name: "source", args: []string{"--source", "/tmp/src"}, want: options{scope: scopeAll, target: target.Opencode, source: "/tmp/src"}},
		{name: "source equals", args: []string{"--source=/tmp/src"}, want: options{scope: scopeAll, target: target.Opencode, source: "/tmp/src"}},
		{name: "all together", args: []string{"agents", "--source", "/tmp/src", "--target", "claude"}, want: options{scope: scopeAgents, target: target.Claude, source: "/tmp/src"}},
		{name: "unknown scope", args: []string{"vscode"}, wantErr: true},
		{name: "unknown target", args: []string{"--target", "vscode"}, wantErr: true},
		{name: "missing target value", args: []string{"--target"}, wantErr: true},
		{name: "missing source value", args: []string{"--source"}, wantErr: true},
		{name: "unknown flag", args: []string{"--force"}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseUpdateArgs(t, tc.args...)
			require.Equal(t, tc.wantErr, err != nil)
			if err == nil {
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestParseUpdateEnvOverridesDefault(t *testing.T) {
	t.Setenv("ZPECS_SOURCE", "/env/src")
	got, err := parseUpdateArgs(t)
	require.NoError(t, err)
	want := options{scope: scopeAll, target: target.Opencode, source: "/env/src"}
	require.Equal(t, want, got)
}

func TestRunRecognizesCommands(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	t.Setenv("ZPECS_SOURCE", gitCloneSource(t))
	sourceDir := t.TempDir()
	cases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: nil},
		{name: "update", args: []string{"update"}},
		{name: "update skills", args: []string{"update", "skills"}},
		{name: "update agents", args: []string{"update", "agents"}},
		{name: "update with target", args: []string{"update", "--target", "claude"}},
		{name: "update with source", args: []string{"update", "skills", "--source", sourceDir}},
		{name: "unknown command", args: []string{"install"}, wantErr: true},
		{name: "update docs", args: []string{"update", "docs"}},
		{name: "unknown scope", args: []string{"update", "vscode"}, wantErr: true},
		{name: "too many arguments", args: []string{"update", "skills", "agents"}, wantErr: true},
		{name: "invalid target", args: []string{"update", "--target", "vscode"}, wantErr: true},
		{name: "unknown flag", args: []string{"update", "--force"}, wantErr: true},
		{name: "missing source", args: []string{"update", "--source", "/tmp/nope/does-not-exist"}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := run(tc.args)
			require.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestPrintErrorShowsUsageForParseErrors(t *testing.T) {
	var c cli
	parser, err := kong.New(&c, cliVars)
	require.NoError(t, err)
	_, parseErr := parser.Parse([]string{"update", "--target", "vscode"})
	require.Error(t, parseErr)

	stderr := captureStderr(t)
	printError(parseErr)
	out := string(stderr())
	require.Contains(t, out, "Usage:")
}

func TestPrintErrorOmitsUsageForRuntimeErrors(t *testing.T) {
	stderr := captureStderr(t)
	printError(errors.New("update outside a repository"))
	out := string(stderr())
	require.NotContains(t, out, "Usage:")
}

func TestBinaryRunsEachUpdateCommand(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	repoDir := testutil.GitRepo(t, nil)
	sourceDir := t.TempDir()
	src := gitCloneSource(t)
	cases := []struct {
		args []string
		want string
	}{
		{args: []string{"update"}, want: "updating skills and agents for opencode"},
		{args: []string{"update", "skills"}, want: "updating skills for opencode"},
		{args: []string{"update", "agents"}, want: "updating agents for opencode"},
		{args: []string{"update", "docs"}, want: "updating docs"},
		{args: []string{"update", "--target", "claude"}, want: "updating skills and agents for claude"},
		{args: []string{"update", "--source", sourceDir, "skills"}, want: "updating skills for opencode from " + sourceDir},
	}

	for _, tc := range cases {
		cmd := exec.Command(binary, tc.args...)
		cmd.Dir = repoDir
		cmd.Env = append(os.Environ(), "ZPECS_SOURCE="+src)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "zpecs %v failed", tc.args)
		require.Contains(t, string(out), tc.want)
	}
}

func TestBinaryRejectsUnknownCommand(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	cmd := exec.Command(binary, "install")
	require.Error(t, cmd.Run())
}

// gitCloneSource returns a temp git repository seeded with source files.
func gitCloneSource(t *testing.T) string {
	t.Helper()
	return testutil.GitRepo(t, map[string]string{
		filepath.Join("skills", "prose-editor", "SKILL.md"): "# prose-editor\n",
		filepath.Join("agents", "code-architect.md"):        "---\nname: code-architect\n---\n\nArchitect code.\n",
		filepath.Join("docs", "zpecs", "prose.md"):          "# Prose guidelines\n",
	})
}

func TestUpdateReadsFromLocalSourceOverDefault(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	t.Setenv("ZPECS_SOURCE", gitCloneSource(t))
	local := t.TempDir()
	testutil.WriteSourceFile(t, local, filepath.Join("skills", "local-only", "SKILL.md"), "# local-only\n")

	require.NoError(t, run([]string{"update", "--source", local}))
	_, err := os.Stat(filepath.Join(".opencode", "skills", "local-only", "SKILL.md"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(".opencode", "skills", "prose-editor", "SKILL.md"))
	require.Error(t, err)
}

func TestBinaryReadsFromDefaultSource(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	work := testutil.GitRepo(t, nil)
	cmd := exec.Command(binary, "update")
	cmd.Dir = work
	cmd.Env = append(os.Environ(), "ZPECS_SOURCE="+gitCloneSource(t))
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zpecs update failed\n%s", out)
	_, err = os.Stat(filepath.Join(work, ".opencode", "skills", "prose-editor", "SKILL.md"))
	require.NoError(t, err)
}

func TestBinaryWritesToClaudeTarget(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	work := testutil.GitRepo(t, nil)
	cmd := exec.Command(binary, "update", "--source", dir, "--target", "claude")
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zpecs update failed\n%s", out)

	_, err = os.Stat(filepath.Join(work, ".claude", "skills", "prose-editor", "SKILL.md"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(work, ".claude", "agents", "prose-editor.md"))
	require.NoError(t, err)
}

func TestBinaryScopedUpdateWritesOnlyWhatItNames(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "code-architect.md"), "---\nname: code-architect\n---\n\nArchitect code.\n")

	for _, tc := range []struct {
		scope string
	}{
		{scope: "skills"},
		{scope: "agents"},
	} {
		t.Run(tc.scope, func(t *testing.T) {
			work := testutil.GitRepo(t, nil)
			cmd := exec.Command(binary, "update", tc.scope, "--source", dir)
			cmd.Dir = work
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "zpecs update %s failed\n%s", tc.scope, out)

			skillPath := filepath.Join(work, ".opencode", "skills", "prose-editor", "SKILL.md")
			agentPath := filepath.Join(work, ".opencode", "agents", "code-architect.md")
			if tc.scope == "skills" {
				_, err := os.Stat(skillPath)
				require.NoError(t, err)
				_, err = os.Stat(agentPath)
				require.Error(t, err)
			} else {
				_, err := os.Stat(agentPath)
				require.NoError(t, err)
				_, err = os.Stat(skillPath)
				require.Error(t, err)
			}
		})
	}
}

func TestBinaryWritesToRepositoryRoot(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	root := testutil.GitRepo(t, nil)
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	work := filepath.Join(root, "nested")
	require.NoError(t, os.MkdirAll(work, 0o755))
	cmd := exec.Command(binary, "update", "--source", dir, "--target", "claude")
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zpecs update failed\n%s", out)

	agentPath := filepath.Join(root, ".claude", "agents", "prose-editor.md")
	_, err = os.Stat(agentPath)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(work, ".claude", "agents", "prose-editor.md"))
	require.Error(t, err)
}

func TestBinaryErrorsOutsideRepository(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	work := t.TempDir()
	cmd := exec.Command(binary, "update", "--source", dir)
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "expected failure outside a repository, got success\n%s", out)
	_, err = os.Stat(filepath.Join(work, ".opencode"))
	require.Error(t, err)
}

func TestBinaryLeavesForeignFileAlone(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\n---\n\nReview prose.\n")

	work := testutil.GitRepo(t, nil)
	path := filepath.Join(work, ".claude", "agents", "prose-editor.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("manual content\n"), 0o644))

	cmd := exec.Command(binary, "update", "--source", dir, "--target", "claude")
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zpecs update failed\n%s", out)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "manual content\n", string(content))
}

func TestBinaryRemovesStaleSkill(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	dir := t.TempDir()
	testutil.WriteSourceFile(t, dir, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n")

	work := testutil.GitRepo(t, nil)
	cmd := exec.Command(binary, "update", "--source", dir)
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "first update failed\n%s", out)
	path := filepath.Join(work, ".opencode", "skills", "prose-editor", "SKILL.md")
	_, err = os.Stat(path)
	require.NoError(t, err)

	require.NoError(t, os.RemoveAll(filepath.Join(dir, "skills")))

	cmd = exec.Command(binary, "update", "--source", dir)
	cmd.Dir = work
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "second update failed\n%s", out)
	_, err = os.Stat(path)
	require.Error(t, err)
}

func TestConvertPrintsTheSpecAsJSON(t *testing.T) {
	path := filepath.Join("..", "spec", "testdata", "convert.md")

	stdout := captureStdout(t)
	require.NoError(t, run([]string{"convert", path}))
	out := stdout()

	var doc spec.Document
	require.NoError(t, json.Unmarshal(out, &doc), "convert output is not one JSON object\n%s", out)
	require.NotEmpty(t, doc.Title)
	require.NotEmpty(t, doc.Purpose)
	require.NotEmpty(t, doc.Requirements)
}

func TestConvertErrorsOnMissingFile(t *testing.T) {
	stdout := captureStdout(t)
	err := run([]string{"convert", filepath.Join(t.TempDir(), "missing.md")})
	require.Error(t, err)
	out := stdout()
	require.Empty(t, out)
}

func TestConvertErrorsOnFileWithoutTopLevelHeading(t *testing.T) {
	path := filepath.Join("..", "spec", "testdata", "no-title.md")

	require.Error(t, run([]string{"convert", path}))
}

func TestConvertErrorsOnRequirementWithoutName(t *testing.T) {
	path := filepath.Join("..", "spec", "testdata", "requirement-without-name.md")

	require.Error(t, run([]string{"convert", path}))
}

func TestBinaryConvertPrintsTheSpecAsJSON(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	path, err := filepath.Abs(filepath.Join("..", "spec", "testdata", "convert.md"))
	require.NoError(t, err)

	cmd := exec.Command(binary, "convert", path)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zpecs convert failed\n%s", out)

	var doc spec.Document
	require.NoError(t, json.Unmarshal(out, &doc), "convert output is not one JSON object\n%s", out)
	require.NotEmpty(t, doc.Title)
	require.NotEmpty(t, doc.Purpose)
	require.NotEmpty(t, doc.Requirements)
}

// captureStdout redirects os.Stdout to a pipe. The returned func reads
// everything written during the capture. The test restores os.Stdout
// when it finishes.
func captureStdout(t *testing.T) func() []byte {
	t.Helper()
	return capture(t, os.Stdout, func(f *os.File) { os.Stdout = f })
}

// captureStderr redirects os.Stderr to a pipe, mirroring captureStdout.
func captureStderr(t *testing.T) func() []byte {
	t.Helper()
	return capture(t, os.Stderr, func(f *os.File) { os.Stderr = f })
}

// capture redirects a process stream to a pipe. The returned func reads
// everything written during the capture. The test restores the stream
// when it finishes.
func capture(t *testing.T, old *os.File, set func(*os.File)) func() []byte {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	set(w)
	t.Cleanup(func() {
		set(old)
		_ = w.Close()
		_ = r.Close()
	})
	return func() []byte {
		t.Helper()
		_ = w.Close()
		out, err := io.ReadAll(r)
		require.NoError(t, err)
		return out
	}
}
