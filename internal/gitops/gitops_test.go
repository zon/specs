package gitops

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zon/specs/internal/testutil"
)

func TestRootResolvesRepositoryRootFromWorkingDirectory(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	t.Chdir(root)
	got, err := Root()
	require.NoError(t, err)
	require.Equal(t, root, got)
}

func TestRootFindsRepositoryFromSubdirectory(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	sub := filepath.Join(root, "docs", "specs")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	t.Chdir(sub)
	got, err := Root()
	require.NoError(t, err)
	require.Equal(t, root, got)
}

func TestRootErrorsOutsideRepository(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	_, err := Root()
	require.Error(t, err)
}

func TestRootReturnsStatErrorForBrokenDotGit(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	sub := filepath.Join(root, "sub")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	// A self-referencing symlink makes os.Stat fail for a reason other
	// than absence. Root must return that error instead of walking up.
	if err := os.Symlink(".git", filepath.Join(sub, ".git")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	t.Chdir(sub)
	_, err := Root()
	require.Error(t, err)
	require.NotErrorIs(t, err, fs.ErrNotExist)
}

func TestCloneCopiesRepository(t *testing.T) {
	src := testutil.GitRepo(t, map[string]string{filepath.Join("skills", "seed", "SKILL.md"): "content\n"})

	dir, cleanup, err := Clone(src)
	require.NoError(t, err)
	defer cleanup()

	require.FileExists(t, filepath.Join(dir, "skills", "seed", "SKILL.md"))
}

func TestCloneCleanupRemovesDirectory(t *testing.T) {
	src := testutil.GitRepo(t, map[string]string{filepath.Join("skills", "seed", "SKILL.md"): "content\n"})

	dir, cleanup, err := Clone(src)
	require.NoError(t, err)
	cleanup()

	require.NoFileExists(t, dir)
}

func TestCloneErrorWrapsCauseAndTrimsOutput(t *testing.T) {
	_, _, err := Clone(filepath.Join(t.TempDir(), "missing"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "cloning ")
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.False(t, strings.HasSuffix(err.Error(), "\n"))
}

func TestIsRemote(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{name: "https URL", source: "https://github.com/zon/specs", want: true},
		{name: "file URL", source: "file:///tmp/repo", want: true},
		{name: "local path", source: "/tmp/src", want: false},
		{name: "scp style", source: "git@github.com:zon/specs.git", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsRemote(tt.source))
		})
	}
}
