package main

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
	"github.com/zon/specs/internal/opencode"
	"github.com/zon/specs/internal/review"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/spec"
	"github.com/zon/specs/internal/testutil"
	"github.com/zon/specs/internal/update"
)

// buildBinary builds the zpecs binary into dir and returns its path.
func buildBinary(t *testing.T, dir string) string {
	t.Helper()
	binary := filepath.Join(dir, "zpecs")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, ".")
	out, err := build.CombinedOutput()
	require.NoError(t, err, "go build failed\n%s", out)
	return binary
}

func TestBuildProducesRunnableCLIBinary(t *testing.T) {
	binary := buildBinary(t, t.TempDir())

	info, err := os.Stat(binary)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		require.True(t, info.Mode().Perm()&0o111 != 0, "binary is not executable: %v", info.Mode())
	}

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

func TestVersionFileHoldsSemver(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "VERSION"))
	require.NoError(t, err)
	require.Regexp(t, `^\d+\.\d+\.\d+$`, strings.TrimSpace(string(data)))
}

// parseUpdateArgs parses `update <args>` with the same kong grammar
// `run` uses. It returns the selected options.
func parseUpdateArgs(t *testing.T, args ...string) (update.Options, error) {
	t.Helper()
	var c cli
	parser, err := kong.New(&c, cliVars)
	require.NoError(t, err)
	if _, err := parser.Parse(append([]string{"update"}, args...)); err != nil {
		return update.Options{}, err
	}
	return update.Options{Scope: c.Update.Scope, Source: c.Update.Source, Target: c.Update.Target}, nil
}

func TestParseUpdate(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    update.Options
		wantErr bool
	}{
		{name: "defaults", args: nil, want: update.Options{Scope: source.ScopeAll, Target: source.Opencode, Source: defaultSourceURL}},
		{name: "all scope", args: []string{"all"}, want: update.Options{Scope: source.ScopeAll, Target: source.Opencode, Source: defaultSourceURL}},
		{name: "scope", args: []string{"skills"}, want: update.Options{Scope: source.ScopeSkills, Target: source.Opencode, Source: defaultSourceURL}},
		{name: "docs scope", args: []string{"docs"}, want: update.Options{Scope: source.ScopeDocs, Target: source.Opencode, Source: defaultSourceURL}},
		{name: "claude target", args: []string{"--target", "claude"}, want: update.Options{Scope: source.ScopeAll, Target: source.Claude, Source: defaultSourceURL}},
		{name: "claude target equals", args: []string{"--target=claude"}, want: update.Options{Scope: source.ScopeAll, Target: source.Claude, Source: defaultSourceURL}},
		{name: "source", args: []string{"--source", "/tmp/src"}, want: update.Options{Scope: source.ScopeAll, Target: source.Opencode, Source: "/tmp/src"}},
		{name: "source equals", args: []string{"--source=/tmp/src"}, want: update.Options{Scope: source.ScopeAll, Target: source.Opencode, Source: "/tmp/src"}},
		{name: "all together", args: []string{"agents", "--source", "/tmp/src", "--target", "claude"}, want: update.Options{Scope: source.ScopeAgents, Target: source.Claude, Source: "/tmp/src"}},
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

// parseReviewArgs parses `review <args>` with the same kong grammar
// `run` uses. It returns the selected options.
func parseReviewArgs(t *testing.T, args ...string) (review.Options, error) {
	t.Helper()
	var c cli
	parser, err := kong.New(&c, cliVars)
	require.NoError(t, err)
	if _, err := parser.Parse(append([]string{"review"}, args...)); err != nil {
		return review.Options{}, err
	}
	return review.Options{Scope: c.Review.Scope, Model: c.Review.Model}, nil
}

func TestParseReview(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    review.Options
		wantErr bool
	}{
		{name: "code", args: []string{"code"}, want: review.Options{Scope: opencode.ScopeCode, Model: opencode.DefaultModel}},
		{name: "architecture", args: []string{"architecture"}, want: review.Options{Scope: opencode.ScopeArchitecture, Model: opencode.DefaultModel}},
		{name: "prose", args: []string{"prose"}, want: review.Options{Scope: opencode.ScopeProse, Model: opencode.DefaultModel}},
		{name: "model", args: []string{"code", "--model", "anthropic/claude-sonnet-4-5"}, want: review.Options{Scope: opencode.ScopeCode, Model: "anthropic/claude-sonnet-4-5"}},
		{name: "model equals", args: []string{"code", "--model=anthropic/claude-sonnet-4-5"}, want: review.Options{Scope: opencode.ScopeCode, Model: "anthropic/claude-sonnet-4-5"}},
		{name: "missing model value", args: []string{"code", "--model"}, wantErr: true},
		{name: "unknown scope", args: []string{"vscode"}, wantErr: true},
		{name: "missing scope", args: nil, wantErr: true},
		{name: "unknown flag", args: []string{"--force"}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseReviewArgs(t, tc.args...)
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
	want := update.Options{Scope: source.ScopeAll, Target: source.Opencode, Source: "/env/src"}
	require.Equal(t, want, got)
}

func TestRunRecognizesCommands(t *testing.T) {
	t.Chdir(testutil.GitRepo(t, nil))
	testutil.FakeOpenCode(t)
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
		{name: "unknown review scope", args: []string{"review", "vscode"}, wantErr: true},
		{name: "review missing scope", args: []string{"review"}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := run(tc.args)
			require.Equal(t, tc.wantErr, err != nil)
		})
	}

	// Each review case must delegate to opencode, so install a fresh fake
	// and assert opencode ran. A no-op review.Run would pass an err == nil
	// check on its own.
	reviewCases := []struct {
		name string
		args []string
	}{
		{name: "review code", args: []string{"review", "code"}},
		{name: "review architecture", args: []string{"review", "architecture"}},
		{name: "review prose", args: []string{"review", "prose"}},
		{name: "review with model", args: []string{"review", "code", "--model", "anthropic/claude-sonnet-4-5"}},
	}

	for _, tc := range reviewCases {
		t.Run(tc.name, func(t *testing.T) {
			read := testutil.FakeOpenCode(t)
			require.NoError(t, run(tc.args))
			require.NotEmpty(t, read(), "review did not delegate to opencode")
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
	}{
		{args: []string{"update"}},
		{args: []string{"update", "skills"}},
		{args: []string{"update", "agents"}},
		{args: []string{"update", "docs"}},
		{args: []string{"update", "--target", "claude"}},
		{args: []string{"update", "--source", sourceDir, "skills"}},
	}

	for _, tc := range cases {
		cmd := exec.Command(binary, tc.args...)
		cmd.Dir = repoDir
		cmd.Env = append(os.Environ(), "ZPECS_SOURCE="+src)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "zpecs %v failed\n%s", tc.args, out)
	}
}

func TestBinaryRejectsUnknownCommand(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	cmd := exec.Command(binary, "install")
	require.Error(t, cmd.Run())
}

func TestBinaryRunsReviewCommand(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	repoDir := testutil.GitRepo(t, nil)
	cases := []struct {
		name  string
		args  []string
		doc   string
		model string
	}{
		{name: "code", args: []string{"review", "code"}, doc: "docs/zpecs/code.md", model: opencode.DefaultModel},
		{name: "architecture", args: []string{"review", "architecture"}, doc: "docs/zpecs/architecture.md", model: opencode.DefaultModel},
		{name: "prose", args: []string{"review", "prose"}, doc: "docs/zpecs/prose.md", model: opencode.DefaultModel},
		{name: "custom model", args: []string{"review", "code", "--model", "anthropic/claude-sonnet-4-5"}, doc: "docs/zpecs/code.md", model: "anthropic/claude-sonnet-4-5"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			read := testutil.FakeOpenCode(t)
			cmd := exec.Command(binary, tc.args...)
			cmd.Dir = repoDir
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "zpecs %v failed\n%s", tc.args, out)
			record := read()
			require.Contains(t, record, "args: run ")
			require.Contains(t, record, "--model "+tc.model)
			require.Contains(t, record, tc.doc)
			require.Contains(t, record, "pwd: "+repoDir)
		})
	}
}

func TestBinaryRejectsUnknownReviewScope(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	repoDir := testutil.GitRepo(t, nil)
	read := testutil.FakeOpenCode(t)
	cmd := exec.Command(binary, "review", "vscode")
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "zpecs review vscode unexpectedly succeeded\n%s", out)
	require.Empty(t, read(), "zpecs review vscode delegated to opencode")
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

func TestConvertPrintsTheSpecAsJSON(t *testing.T) {
	path := filepath.Join("..", "..", "internal", "spec", "testdata", "convert.md")

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
	path := filepath.Join("..", "..", "internal", "spec", "testdata", "no-title.md")

	require.Error(t, run([]string{"convert", path}))
}

func TestConvertErrorsOnRequirementWithoutName(t *testing.T) {
	path := filepath.Join("..", "..", "internal", "spec", "testdata", "requirement-without-name.md")

	require.Error(t, run([]string{"convert", path}))
}

func TestBinaryConvertPrintsTheSpecAsJSON(t *testing.T) {
	binary := buildBinary(t, t.TempDir())
	path, err := filepath.Abs(filepath.Join("..", "..", "internal", "spec", "testdata", "convert.md"))
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

// captureStdout redirects os.Stdout the way capture does.
func captureStdout(t *testing.T) func() []byte {
	t.Helper()
	return capture(t, os.Stdout, func(f *os.File) { os.Stdout = f })
}

// captureStderr redirects os.Stderr the same way captureStdout does.
func captureStderr(t *testing.T) func() []byte {
	t.Helper()
	return capture(t, os.Stderr, func(f *os.File) { os.Stderr = f })
}

// capture redirects a process stream to a pipe. The returned func reads
// everything written during the capture.
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
