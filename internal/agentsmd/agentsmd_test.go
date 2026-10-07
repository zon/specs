package agentsmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceAppendsSectionWhenAbsent(t *testing.T) {
	content := "# Title\n\nIntro.\n"
	section := "## Zpecs\n\nNew section.\n"

	got := Replace(content, section)

	require.Equal(t, "# Title\n\nIntro.\n\n## Zpecs\n\nNew section.\n", got)
}

func TestReplaceWritesSectionIntoEmptyContent(t *testing.T) {
	section := "## Zpecs\n\nNew section.\n"

	got := Replace("", section)

	require.Equal(t, "## Zpecs\n\nNew section.\n", got)
}

func TestReplaceReplacesSectionInPlace(t *testing.T) {
	content := "# Title\n\nIntro.\n\n## Zpecs\n\nOld section.\n\n## Other\n\nBody.\n"
	section := "## Zpecs\n\nNew section.\n"

	got := Replace(content, section)

	want := "# Title\n\nIntro.\n\n## Zpecs\n\nNew section.\n\n## Other\n\nBody.\n"
	require.Equal(t, want, got)
}

func TestReplaceKeepsSectionAtEnd(t *testing.T) {
	content := "# Title\n\n## Zpecs\n\nOld section.\n"
	section := "## Zpecs\n\nNew section.\n"

	got := Replace(content, section)

	require.Equal(t, "# Title\n\n## Zpecs\n\nNew section.\n", got)
}

func TestReplaceKeepsOnlyOneSection(t *testing.T) {
	content := "## Zpecs\n\nOld section.\n"
	section := "## Zpecs\n\nNew section.\n"

	got := Replace(content, section)

	require.Equal(t, "## Zpecs\n\nNew section.\n", got)
}

func TestUpdateCreatesAgentsFile(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, Update(root, "## Zpecs\n\nNew section.\n"))

	require.FileExists(t, filepath.Join(root, FileName))
	content, err := os.ReadFile(filepath.Join(root, FileName))
	require.NoError(t, err)
	require.Equal(t, "## Zpecs\n\nNew section.\n", string(content))
}

func TestUpdateKeepsOtherContent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, FileName)
	require.NoError(t, os.WriteFile(path, []byte("# Title\n\n## Zpecs\n\nOld.\n\n## Other\n\nBody.\n"), 0o644))

	require.NoError(t, Update(root, "## Zpecs\n\nNew.\n"))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "# Title\n\n## Zpecs\n\nNew.\n\n## Other\n\nBody.\n", string(content))
}
