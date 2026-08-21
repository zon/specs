package opencode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/testutil"
)

func TestUnmarshalScope(t *testing.T) {
	cases := []struct {
		name    string
		s       string
		want    Scope
		wantErr bool
	}{
		{name: "code", s: "code", want: ScopeCode},
		{name: "architecture", s: "architecture", want: ScopeArchitecture},
		{name: "prose", s: "prose", want: ScopeProse},
		{name: "unknown scope", s: "vscode", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got Scope
			err := got.UnmarshalText([]byte(tc.s))
			require.Equal(t, tc.wantErr, err != nil)
			if err == nil {
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestGuideline(t *testing.T) {
	cases := []struct {
		name    string
		scope   Scope
		want    string
		wantErr bool
	}{
		{name: "architecture", scope: ScopeArchitecture, want: "docs/zpecs/architecture.md"},
		{name: "prose", scope: ScopeProse, want: "docs/zpecs/prose.md"},
		{name: "unknown scope", scope: Scope(99), wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := guideline(tc.scope)
			require.Equal(t, tc.wantErr, err != nil)
			if err == nil {
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestReviewRunsOpenCodeWithThePrompt(t *testing.T) {
	cases := []struct {
		name  string
		scope Scope
		doc   string
	}{
		{name: "architecture", scope: ScopeArchitecture, doc: "docs/zpecs/architecture.md"},
		{name: "prose", scope: ScopeProse, doc: "docs/zpecs/prose.md"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			read := testutil.FakeOpenCode(t)

			dir := t.TempDir()
			require.NoError(t, Review(dir, tc.scope, "", ""))
			record := read()
			require.Contains(t, record, testutil.RanAgainst(DefaultModel, tc.doc))
		})
	}
}

func TestCodePromptDirectsIssuesToRefactorProjects(t *testing.T) {
	msg, err := prompt(ScopeCode)
	require.NoError(t, err)
	require.Equal(t, CodeReviewPrompt, msg)
	require.Contains(t, msg, "Do not edit the code")
	require.Contains(t, msg, "refactor-<slug>.yaml")
	require.Contains(t, msg, "projects/")
	require.Contains(t, msg, "for its refactoring")
	require.Contains(t, msg, "Update a matching project when one exists")
	require.Contains(t, msg, "Write no project when you find no issues")
}

func TestRefactorInstructionStaysOutOfOtherScopes(t *testing.T) {
	cases := []struct {
		name  string
		scope Scope
	}{
		{name: "architecture", scope: ScopeArchitecture},
		{name: "prose", scope: ScopeProse},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := prompt(tc.scope)
			require.NoError(t, err)
			require.NotContains(t, msg, "refactor-<slug>.yaml")
		})
	}
}

func TestReviewErrorsOnUnknownScope(t *testing.T) {
	read := testutil.FakeOpenCode(t)

	err := Review(t.TempDir(), Scope(99), "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown scope")
	require.Empty(t, read())
}

func TestReviewRunsInTheGivenDirectory(t *testing.T) {
	read := testutil.FakeOpenCode(t)

	dir := t.TempDir()
	require.NoError(t, Review(dir, ScopeCode, "", ""))
	require.Contains(t, read(), testutil.RanIn(dir))
}

func TestReviewErrorsOnOpenCodeFailure(t *testing.T) {
	dir := t.TempDir()
	script := "#!/bin/sh\nexit 1\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "opencode"), []byte(script), 0o755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := Review(t.TempDir(), ScopeCode, "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "opencode")
}

func TestReviewRunsWithTheDefaultModel(t *testing.T) {
	read := testutil.FakeOpenCode(t)

	require.NoError(t, Review(t.TempDir(), ScopeCode, "", ""))

	require.Contains(t, read(), testutil.RanAgainstMessage(DefaultModel, CodeReviewPrompt))
}

func TestReviewRunsWithTheGivenModel(t *testing.T) {
	read := testutil.FakeOpenCode(t)

	require.NoError(t, Review(t.TempDir(), ScopeCode, "anthropic/claude-sonnet-4-5", ""))

	require.Contains(t, read(), testutil.RanAgainstMessage("anthropic/claude-sonnet-4-5", CodeReviewPrompt))
}

func TestReviewRunsWithTheGivenVariant(t *testing.T) {
	read := testutil.FakeOpenCode(t)

	require.NoError(t, Review(t.TempDir(), ScopeCode, "", "minimal"))

	require.Contains(t, read(), testutil.RanAgainstMessageVariant(DefaultModel, "minimal", CodeReviewPrompt))
}

func TestReviewRunsWithTheModelAndVariant(t *testing.T) {
	read := testutil.FakeOpenCode(t)

	dir := t.TempDir()
	require.NoError(t, Review(dir, ScopeCode, "anthropic/claude-sonnet-4-5", "minimal"))

	record := read()
	require.Contains(t, record, testutil.RanAgainstMessageVariant("anthropic/claude-sonnet-4-5", "minimal", CodeReviewPrompt))
	require.Contains(t, record, testutil.RanIn(dir))
}

func TestReviewOmitsTheVariantFlagWhenEmpty(t *testing.T) {
	read := testutil.FakeOpenCode(t)

	require.NoError(t, Review(t.TempDir(), ScopeCode, "anthropic/claude-sonnet-4-5", ""))

	require.NotContains(t, read(), "--variant")
}
