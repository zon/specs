package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/frontmatter"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/target"
)

func TestClaudeAgentKeepsName(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor"}

	got := ClaudeAgent(fields, "Review prose.")

	require.Contains(t, got, "name: prose-editor")
}

func TestClaudeAgentRendersNameDescriptionAndBody(t *testing.T) {
	fields := frontmatter.Fields{
		Name:        "prose-editor",
		Description: "Reviews prose against the guidelines.",
	}

	got := ClaudeAgent(fields, "Review prose against the guidelines.")
	want := "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"

	require.Equal(t, want, got)
}

func TestClaudeAgentListsTools(t *testing.T) {
	fields := frontmatter.Fields{Tools: []string{"read", "edit"}}

	got := ClaudeAgent(fields, "Review prose.")

	require.Contains(t, got, "tools:\n  - read\n  - edit")
}

func TestClaudeAgentRendersToolsAfterDescription(t *testing.T) {
	fields := frontmatter.Fields{
		Name:        "prose-editor",
		Description: "Reviews prose.",
		Tools:       []string{"read", "edit"},
	}

	got := ClaudeAgent(fields, "Review prose.")
	want := "---\nname: prose-editor\ndescription: Reviews prose.\ntools:\n  - read\n  - edit\n---\n\nReview prose.\n"

	require.Equal(t, want, got)
}

func TestClaudeAgentOmitsToolsWhenEmpty(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor"}

	got := ClaudeAgent(fields, "Review prose.")

	require.NotContains(t, got, "tools:")
}

func TestOpencodeAgentDefaultsToSubagentMode(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor"}

	got, err := OpencodeAgent(fields, "Review prose.")
	require.NoError(t, err)

	require.Contains(t, got, "mode: subagent")
}

func TestOpencodeAgentUsesDefinitionMode(t *testing.T) {
	fields := frontmatter.Fields{Name: "code-architect", Mode: "primary"}

	got, err := OpencodeAgent(fields, "Plan the work.")
	require.NoError(t, err)

	require.Contains(t, got, "mode: primary")
}

func TestOpencodeAgentRejectsUnknownMode(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor", Mode: "banana"}

	_, err := OpencodeAgent(fields, "Review prose.")
	require.Error(t, err)
}

func TestOpencodeAgentDropsName(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor"}

	got, err := OpencodeAgent(fields, "Review prose.")
	require.NoError(t, err)

	require.NotContains(t, got, "name:")
}

func TestOpencodeAgentRendersModeDescriptionAndBody(t *testing.T) {
	fields := frontmatter.Fields{Description: "Reviews prose against the guidelines."}

	got, err := OpencodeAgent(fields, "Review prose against the guidelines.")
	require.NoError(t, err)
	want := "---\nmode: subagent\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"

	require.Equal(t, want, got)
}

func TestOpencodeAgentDeniesEveryOtherTool(t *testing.T) {
	fields := frontmatter.Fields{Tools: []string{"read", "edit"}}

	got, err := OpencodeAgent(fields, "Review prose.")
	require.NoError(t, err)

	for _, denied := range []string{"bash", "write", "grep", "glob"} {
		require.Contains(t, got, denied+": deny")
	}
}

func TestOpencodeAgentDoesNotDenyListedTools(t *testing.T) {
	fields := frontmatter.Fields{Tools: []string{"read", "edit"}}

	got, err := OpencodeAgent(fields, "Review prose.")
	require.NoError(t, err)

	require.NotContains(t, got, "read: deny")
	require.NotContains(t, got, "edit: deny")
}

func TestOpencodeAgentRendersDenyRulesAfterDescription(t *testing.T) {
	fields := frontmatter.Fields{
		Description: "Reviews prose.",
		Tools:       []string{"read", "edit"},
	}

	got, err := OpencodeAgent(fields, "Review prose.")
	require.NoError(t, err)
	want := "---\nmode: subagent\ndescription: Reviews prose.\npermission:\n" +
		"  apply_patch: deny\n  bash: deny\n  glob: deny\n  grep: deny\n" +
		"  lsp: deny\n  question: deny\n  skill: deny\n  todowrite: deny\n" +
		"  webfetch: deny\n  websearch: deny\n  write: deny\n" +
		"---\n\nReview prose.\n"

	require.Equal(t, want, got)
}

func TestOpencodeAgentOmitsDenyRulesWhenToolsEmpty(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor"}

	got, err := OpencodeAgent(fields, "Review prose.")
	require.NoError(t, err)

	require.NotContains(t, got, "permission:")
}

func TestDefinitionReturnsSkillVerbatim(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n\nReview prose.\n")

	got, err := Definition(source.Definition{Kind: source.Skill, Name: "prose-editor", Path: path}, target.Claude)
	require.NoError(t, err)
	want := "# prose-editor\n\nReview prose.\n"
	require.Equal(t, want, got)
}

func TestDefinitionReturnsDocVerbatim(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	got, err := Definition(source.Definition{Kind: source.Doc, Name: "prose", Path: path}, target.Opencode)
	require.NoError(t, err)
	want := "# Prose guidelines\n"
	require.Equal(t, want, got)
}

func TestDefinitionRendersAgentForClaude(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n")

	got, err := Definition(source.Definition{Kind: source.Agent, Name: "prose-editor", Path: path}, target.Claude)
	require.NoError(t, err)
	want := "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"
	require.Equal(t, want, got)
}

func TestDefinitionRendersAgentForOpencode(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n")

	got, err := Definition(source.Definition{Kind: source.Agent, Name: "prose-editor", Path: path}, target.Opencode)
	require.NoError(t, err)
	want := "---\nmode: subagent\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"
	require.Equal(t, want, got)
}

func TestDefinitionReportsUnreadableAgentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agents", "missing.md")

	_, err := Definition(source.Definition{Kind: source.Agent, Name: "missing", Path: path}, target.Opencode)
	require.Error(t, err)
	require.Contains(t, err.Error(), "reading")
}

// writeDefinitionFile writes content at rel under a fresh temp dir and
// returns the path.
func writeDefinitionFile(t *testing.T, rel, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}
