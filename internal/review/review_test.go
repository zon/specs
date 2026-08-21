package review

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/opencode"
	"github.com/zon/specs/internal/testutil"
)

func TestRunRunsOpenCodeAtTheRepositoryRoot(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	read := testutil.FakeOpenCode(t)

	work := testutil.ChdirInto(t, root, "nested", "deep")

	require.NoError(t, Run(Options{Scope: opencode.ScopeCode}))

	require.Contains(t, read(), testutil.RanIn(root))
	require.NotContains(t, read(), testutil.RanIn(work))
}

func TestRunForwardsTheScope(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	read := testutil.FakeOpenCode(t)
	testutil.ChdirInto(t, root)

	require.NoError(t, Run(Options{Scope: opencode.ScopeArchitecture}))

	require.Contains(t, read(), testutil.RanAgainstMessage(opencode.DefaultModel, opencode.ArchitectureReviewPrompt))
}

func TestRunForwardsTheModel(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	read := testutil.FakeOpenCode(t)
	testutil.ChdirInto(t, root)

	require.NoError(t, Run(Options{Scope: opencode.ScopeCode, Model: "anthropic/claude-sonnet-4-5"}))

	require.Contains(t, read(), testutil.RanAgainstMessage("anthropic/claude-sonnet-4-5", opencode.CodeReviewPrompt))
}

func TestRunForwardsTheVariant(t *testing.T) {
	root := testutil.GitRepo(t, nil)
	read := testutil.FakeOpenCode(t)
	testutil.ChdirInto(t, root)

	require.NoError(t, Run(Options{Scope: opencode.ScopeCode, Model: opencode.DefaultModel, Variant: "minimal"}))

	require.Contains(t, read(), testutil.RanAgainstMessageVariant(opencode.DefaultModel, "minimal", opencode.CodeReviewPrompt))
}

func TestRunErrorsOutsideRepository(t *testing.T) {
	testutil.ChdirInto(t, t.TempDir())
	read := testutil.FakeOpenCode(t)

	err := Run(Options{Scope: opencode.ScopeCode})
	require.Error(t, err)
	require.Empty(t, read())
}
