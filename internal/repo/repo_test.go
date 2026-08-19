package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// gitRoot returns a temp dir that looks like a git repository root.
func gitRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))
	return dir
}

func TestRootAtRepositoryRoot(t *testing.T) {
	root := gitRoot(t)
	got, err := Root(root)
	require.NoError(t, err)
	require.Equal(t, root, got)
}

func TestRootFindsRepositoryFromSubdirectory(t *testing.T) {
	root := gitRoot(t)
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
