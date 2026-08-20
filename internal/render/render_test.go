package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/target"
)

func TestClaudeAgentKeepsName(t *testing.T) {
	fields := fields{Name: "prose-editor"}

	got := claudeAgent(fields, "Review prose.")

	require.Contains(t, got, "name: prose-editor")
}

func TestClaudeAgentRendersNameDescriptionAndBody(t *testing.T) {
	fields := fields{
		Name:        "prose-editor",
		Description: "Reviews prose against the guidelines.",
	}

	got := claudeAgent(fields, "Review prose against the guidelines.")
	want := "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"

	require.Equal(t, want, got)
}

func TestClaudeAgentListsTools(t *testing.T) {
	fields := fields{Tools: []string{"read", "edit"}}

	got := claudeAgent(fields, "Review prose.")

	require.Contains(t, got, "tools:\n  - read\n  - edit")
}

func TestClaudeAgentRendersToolsAfterDescription(t *testing.T) {
	fields := fields{
		Name:        "prose-editor",
		Description: "Reviews prose.",
		Tools:       []string{"read", "edit"},
	}

	got := claudeAgent(fields, "Review prose.")
	want := "---\nname: prose-editor\ndescription: Reviews prose.\ntools:\n  - read\n  - edit\n---\n\nReview prose.\n"

	require.Equal(t, want, got)
}

func TestClaudeAgentOmitsToolsWhenEmpty(t *testing.T) {
	fields := fields{Name: "prose-editor"}

	got := claudeAgent(fields, "Review prose.")

	require.NotContains(t, got, "tools:")
}

func TestOpencodeAgentDefaultsToSubagentMode(t *testing.T) {
	fields := fields{Name: "prose-editor"}

	got, err := opencodeAgent(fields, "Review prose.")
	require.NoError(t, err)

	require.Contains(t, got, "mode: subagent")
}

func TestOpencodeAgentUsesDefinitionMode(t *testing.T) {
	fields := fields{Name: "code-architect", Mode: "primary"}

	got, err := opencodeAgent(fields, "Plan the work.")
	require.NoError(t, err)

	require.Contains(t, got, "mode: primary")
}

func TestOpencodeAgentRejectsUnknownMode(t *testing.T) {
	fields := fields{Name: "prose-editor", Mode: "banana"}

	_, err := opencodeAgent(fields, "Review prose.")
	require.Error(t, err)
}

func TestOpencodeAgentDropsName(t *testing.T) {
	fields := fields{Name: "prose-editor"}

	got, err := opencodeAgent(fields, "Review prose.")
	require.NoError(t, err)

	require.NotContains(t, got, "name:")
}

func TestOpencodeAgentRendersModeDescriptionAndBody(t *testing.T) {
	fields := fields{Description: "Reviews prose against the guidelines."}

	got, err := opencodeAgent(fields, "Review prose against the guidelines.")
	require.NoError(t, err)
	want := "---\nmode: subagent\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"

	require.Equal(t, want, got)
}

func TestOpencodeAgentDeniesUnlistedTools(t *testing.T) {
	fields := fields{Tools: []string{"read", "edit"}}

	got, err := opencodeAgent(fields, "Review prose.")
	require.NoError(t, err)

	for _, denied := range []string{"bash", "write", "grep", "glob"} {
		require.Contains(t, got, denied+": deny")
	}
}

func TestOpencodeAgentAllowsListedTools(t *testing.T) {
	fields := fields{Tools: []string{"read", "edit"}}

	got, err := opencodeAgent(fields, "Review prose.")
	require.NoError(t, err)

	require.NotContains(t, got, "read: deny")
	require.NotContains(t, got, "edit: deny")
}

func TestOpencodeAgentRendersDenyRulesAfterDescription(t *testing.T) {
	fields := fields{
		Description: "Reviews prose.",
		Tools:       []string{"read", "edit"},
	}

	got, err := opencodeAgent(fields, "Review prose.")
	require.NoError(t, err)
	want := "---\nmode: subagent\ndescription: Reviews prose.\npermission:\n" +
		"  apply_patch: deny\n  bash: deny\n  glob: deny\n  grep: deny\n" +
		"  lsp: deny\n  question: deny\n  skill: deny\n  todowrite: deny\n" +
		"  webfetch: deny\n  websearch: deny\n  write: deny\n" +
		"---\n\nReview prose.\n"

	require.Equal(t, want, got)
}

func TestOpencodeAgentOmitsDenyRulesWhenToolsEmpty(t *testing.T) {
	fields := fields{Name: "prose-editor"}

	got, err := opencodeAgent(fields, "Review prose.")
	require.NoError(t, err)

	require.NotContains(t, got, "permission:")
}

func TestDefinitionReturnsSkillVerbatim(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n\nReview prose.\n")

	got, err := definition(source.Definition{Kind: source.Skill, Name: "prose-editor", Path: path}, target.Claude)
	require.NoError(t, err)
	want := "# prose-editor\n\nReview prose.\n"
	require.Equal(t, want, got)
}

func TestDefinitionReturnsDocVerbatim(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	got, err := definition(source.Definition{Kind: source.Doc, Name: "prose", Path: path}, target.Opencode)
	require.NoError(t, err)
	want := "# Prose guidelines\n"
	require.Equal(t, want, got)
}

func TestDefinitionRendersAgentForClaude(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n")

	got, err := definition(source.Definition{Kind: source.Agent, Name: "prose-editor", Path: path}, target.Claude)
	require.NoError(t, err)
	want := "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"
	require.Equal(t, want, got)
}

func TestDefinitionRendersAgentForOpencode(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n")

	got, err := definition(source.Definition{Kind: source.Agent, Name: "prose-editor", Path: path}, target.Opencode)
	require.NoError(t, err)
	want := "---\nmode: subagent\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"
	require.Equal(t, want, got)
}

func TestDefinitionReportsUnreadableAgentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agents", "missing.md")

	_, err := definition(source.Definition{Kind: source.Agent, Name: "missing", Path: path}, target.Opencode)
	require.Error(t, err)
	require.Contains(t, err.Error(), "reading")
}

func TestDefinitionReportsUnreadableSkillFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "skills", "missing", "SKILL.md")

	_, err := definition(source.Definition{Kind: source.Skill, Name: "missing", Path: path}, target.Opencode)
	require.Error(t, err)
	require.Contains(t, err.Error(), "reading")
}

func TestDefinitionReportsInvalidAgentFrontmatter(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("agents", "prose-editor.md"), "---\ntools: [read, edit\n---\n")

	_, err := definition(source.Definition{Kind: source.Agent, Name: "prose-editor", Path: path}, target.Claude)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parsing")
}

func TestForTargetRendersAgentForOpencode(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\ndescription: Reviews prose with style.\n---\n\nReview prose with style.\n")
	d := source.Definition{Kind: source.Agent, Name: "prose-editor", Path: path}

	got, err := ForTarget(target.Opencode)(d)
	require.NoError(t, err)
	want := "---\nmode: subagent\ndescription: Reviews prose with style.\n---\n\nReview prose with style.\n"

	require.Equal(t, want, got)
}

// writeDefinitionFile writes content to rel in a fresh temp dir and
// returns the path.
func writeDefinitionFile(t *testing.T, rel, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestReadReadsNameDescriptionAndTools(t *testing.T) {
	path := writeDefinitionFile(t, "definition.md", `---
name: prose-editor
description: "Reviews prose: checks it against the guidelines, and fixes it."
tools:
  - read
  - edit
---

Body text.
`)
	want := fields{
		Name:        "prose-editor",
		Description: "Reviews prose: checks it against the guidelines, and fixes it.",
		Tools:       []string{"read", "edit"},
	}

	got, err := read(path)
	require.NoError(t, err)
	require.Equal(t, want, got.fields)
}

func TestReadReadsInlineTools(t *testing.T) {
	path := writeDefinitionFile(t, "definition.md", "---\nname: prose-editor\ntools: [read, edit]\n---\n")

	got, err := read(path)
	require.NoError(t, err)
	require.Equal(t, []string{"read", "edit"}, got.fields.Tools)
}

func TestReadReadsMode(t *testing.T) {
	path := writeDefinitionFile(t, "definition.md", "---\nname: code-architect\nmode: primary\n---\n\nPlan the work.\n")

	got, err := read(path)
	require.NoError(t, err)
	require.Equal(t, "primary", got.fields.Mode)
}

func TestReadYieldsZeroFieldsWithoutFrontmatter(t *testing.T) {
	path := writeDefinitionFile(t, "definition.md", "# Just a body\n")

	got, err := read(path)
	require.NoError(t, err)
	require.Equal(t, fields{}, got.fields)
}

func TestReadErrorsOnUnterminatedFrontmatter(t *testing.T) {
	path := writeDefinitionFile(t, "definition.md", "---\nname: prose-editor\n")

	_, err := read(path)
	require.ErrorContains(t, err, "unterminated frontmatter")
}

func TestReadErrorsOnInvalidFrontmatter(t *testing.T) {
	path := writeDefinitionFile(t, "definition.md", "---\ntools: [read, edit\n---\n")

	_, err := read(path)
	require.ErrorContains(t, err, "parsing")
}

func TestReadErrorsOnMissingFile(t *testing.T) {
	_, err := read(filepath.Join(t.TempDir(), "missing.md"))
	require.ErrorContains(t, err, "reading")
}

func TestReadReadsBody(t *testing.T) {
	path := writeDefinitionFile(t, "definition.md", `---
name: prose-editor
description: Reviews prose.
---

You are a prose editor. Review the prose against the guidelines.
`)

	got, err := read(path)
	require.NoError(t, err)
	want := "You are a prose editor. Review the prose against the guidelines."
	require.Equal(t, want, got.body)
}

func TestReadKeepsBodyLines(t *testing.T) {
	path := writeDefinitionFile(t, "definition.md", `---
name: prose-editor
---

You are a prose editor.

Give numbered instructions.
`)

	got, err := read(path)
	require.NoError(t, err)
	want := "You are a prose editor.\n\nGive numbered instructions."
	require.Equal(t, want, got.body)
}

func TestReadYieldsWholeContentAsBodyWithoutFrontmatter(t *testing.T) {
	path := writeDefinitionFile(t, "definition.md", "You are a prose editor.\n")

	got, err := read(path)
	require.NoError(t, err)
	require.Equal(t, "You are a prose editor.", got.body)
}

func TestReadYieldsEmptyBody(t *testing.T) {
	path := writeDefinitionFile(t, "definition.md", "---\nname: prose-editor\n---\n")

	got, err := read(path)
	require.NoError(t, err)
	require.Equal(t, "", got.body)
}

func TestReadYieldsZeroFieldsForEmptyFile(t *testing.T) {
	path := writeDefinitionFile(t, "definition.md", "")

	got, err := read(path)
	require.NoError(t, err)
	require.Equal(t, fields{}, got.fields)
	require.Equal(t, "", got.body)
}
