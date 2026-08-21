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
		{name: "code", scope: ScopeCode, want: "docs/zpecs/code.md"},
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
		{name: "code", scope: ScopeCode, doc: "docs/zpecs/code.md"},
		{name: "architecture", scope: ScopeArchitecture, doc: "docs/zpecs/architecture.md"},
		{name: "prose", scope: ScopeProse, doc: "docs/zpecs/prose.md"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			read := testutil.FakeOpenCode(t)

			dir := t.TempDir()
			require.NoError(t, Review(dir, tc.scope))
			record := read()
			require.Contains(t, record, "args: run ")
			require.Contains(t, record, tc.doc)
		})
	}
}

func TestReviewErrorsOnUnknownScope(t *testing.T) {
	read := testutil.FakeOpenCode(t)

	err := Review(t.TempDir(), Scope(99))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown scope")
	require.Empty(t, read())
}

func TestReviewRunsInTheGivenDirectory(t *testing.T) {
	read := testutil.FakeOpenCode(t)

	dir := t.TempDir()
	require.NoError(t, Review(dir, ScopeCode))
	require.Contains(t, read(), "pwd: "+dir+"\n")
}

func TestReviewErrorsOnOpenCodeFailure(t *testing.T) {
	dir := t.TempDir()
	script := "#!/bin/sh\nexit 1\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "opencode"), []byte(script), 0o755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := Review(t.TempDir(), ScopeCode)
	require.Error(t, err)
	require.Contains(t, err.Error(), "opencode")
}
