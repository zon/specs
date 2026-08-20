package update

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/testutil"
)

func TestResolveSourceClonesRemote(t *testing.T) {
	src := testutil.GitRepoURL(t, map[string]string{"seed": "content\n"})

	dir, label, cleanup, err := resolveSource(src)
	require.NoError(t, err)
	defer cleanup()
	require.Equal(t, src, label)
	require.DirExists(t, dir)
	require.FileExists(t, filepath.Join(dir, "seed"))
	cleanup()
	require.NoFileExists(t, dir)
}

func TestResolveSourceReadsLocalInPlace(t *testing.T) {
	dir := t.TempDir()

	gotDir, label, cleanup, err := resolveSource(dir)
	require.NoError(t, err)
	require.Equal(t, dir, gotDir)
	require.Equal(t, dir, label)
	cleanup()
	require.DirExists(t, dir)
}
