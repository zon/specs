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
