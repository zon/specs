package review

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/opencode"
	"github.com/zon/specs/internal/testutil"
)

func TestRunRunsOpenCodeAtTheRepositoryRoot(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	read := testutil.FakeOpenCode(t)

	work := filepath.Join(root, "nested", "deep")
	require.NoError(t, os.MkdirAll(work, 0o755))
	t.Chdir(work)

	require.NoError(t, Run(Options{Scope: opencode.ScopeCode}))

	require.Contains(t, read(), testutil.RanIn(root))
	require.NotContains(t, read(), testutil.RanIn(work))
}

func TestRunErrorsOutsideRepository(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	read := testutil.FakeOpenCode(t)

	err := Run(Options{Scope: opencode.ScopeCode})
	require.Error(t, err)
	require.Empty(t, read())
}
