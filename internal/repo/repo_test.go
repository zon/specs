package repo

import (
	"os"
	"path/filepath"
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
