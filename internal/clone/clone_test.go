package clone

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zon/specs/internal/testutil"
)

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
