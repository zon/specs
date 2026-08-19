package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zon/specs/internal/testutil"
)

func TestRootAtRepositoryRoot(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	got, err := Root(root)
	require.NoError(t, err)
	require.Equal(t, root, got)
}

func TestRootFindsRepositoryFromSubdirectory(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	sub := filepath.Join(root, "docs", "specs")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	got, err := Root(sub)
	require.NoError(t, err)
	require.Equal(t, root, got)
}

func TestRootErrorsOutsideRepository(t *testing.T) {
	dir := t.TempDir()
	_, err := Root(dir)
	require.Error(t, err)
}
