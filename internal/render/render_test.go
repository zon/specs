package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zon/specs/internal/source"
)

func TestEnabledMapsOrchestrationFeature(t *testing.T) {
	cases := []struct {
		name          string
		orchestration bool
		want          Features
	}{
		{name: "off", orchestration: false, want: Features{}},
		{name: "on", orchestration: true, want: Features{Orchestration: true}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, Enabled(tc.orchestration))
		})
	}
}

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

	got, err := definition(source.Definition{Kind: source.Skill, Name: "prose-editor", Path: path}, source.Claude, Features{})
	require.NoError(t, err)
	want := "# prose-editor\n\nReview prose.\n"
	require.Equal(t, want, got)
}

func TestDefinitionReturnsDocVerbatim(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	got, err := definition(source.Definition{Kind: source.Doc, Name: "prose", Path: path}, source.Opencode, Features{})
	require.NoError(t, err)
	want := "# Prose guidelines\n"
	require.Equal(t, want, got)
}

func TestDefinitionRendersAgentForClaude(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n")

	got, err := definition(source.Definition{Kind: source.Agent, Name: "prose-editor", Path: path}, source.Claude, Features{})
	require.NoError(t, err)
	want := "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"
	require.Equal(t, want, got)
}

func TestDefinitionRendersAgentForOpencode(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n")

	got, err := definition(source.Definition{Kind: source.Agent, Name: "prose-editor", Path: path}, source.Opencode, Features{})
	require.NoError(t, err)
	want := "---\nmode: subagent\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"
	require.Equal(t, want, got)
}

func TestDefinitionReportsUnreadableAgentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agents", "missing.md")

	_, err := definition(source.Definition{Kind: source.Agent, Name: "missing", Path: path}, source.Opencode, Features{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "reading")
}

func TestDefinitionReportsUnreadableSkillFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "skills", "missing", "SKILL.md")

	_, err := definition(source.Definition{Kind: source.Skill, Name: "missing", Path: path}, source.Opencode, Features{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "reading")
}

func TestDefinitionReportsInvalidAgentFrontmatter(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("agents", "prose-editor.md"), "---\ntools: [read, edit\n---\n")

	_, err := definition(source.Definition{Kind: source.Agent, Name: "prose-editor", Path: path}, source.Claude, Features{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "parsing")
}

func TestForTargetRendersAgentForOpencode(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\ndescription: Reviews prose with style.\n---\n\nReview prose with style.\n")
	d := source.Definition{Kind: source.Agent, Name: "prose-editor", Path: path}

	got, err := ForTarget(source.Opencode, Features{})(d)
	require.NoError(t, err)
	want := "---\nmode: subagent\ndescription: Reviews prose with style.\n---\n\nReview prose with style.\n"

	require.Equal(t, want, got)
}

// TestForTargetRendersEveryKindWithEnabledFeatures checks that enabling a
// feature does not break rendering of a skill, a doc, or an agent. None of
// these definitions carry template actions, so the enabled feature does not
// change their output.
func TestForTargetRendersEveryKindWithEnabledFeatures(t *testing.T) {
	features := Features{Orchestration: true}
	cases := []struct {
		name    string
		kind    source.Kind
		defName string
		target  string
		rel     string
		content string
		want    string
	}{
		{
			name:    "skill",
			kind:    source.Skill,
			defName: "prose-editor",
			target:  source.Claude,
			rel:     filepath.Join("skills", "prose-editor", "SKILL.md"),
			content: "# prose-editor\n\nReview prose.\n",
			want:    "# prose-editor\n\nReview prose.\n",
		},
		{
			name:    "doc",
			kind:    source.Doc,
			defName: "prose",
			target:  source.Opencode,
			rel:     filepath.Join("docs", "zpecs", "prose.md"),
			content: "# Prose guidelines\n",
			want:    "# Prose guidelines\n",
		},
		{
			name:    "agent",
			kind:    source.Agent,
			defName: "prose-editor",
			target:  source.Claude,
			rel:     filepath.Join("agents", "prose-editor.md"),
			content: "---\nname: prose-editor\ndescription: Reviews prose.\n---\n\nReview prose.\n",
			want:    "---\nname: prose-editor\ndescription: Reviews prose.\n---\n\nReview prose.\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeDefinitionFile(t, tc.rel, tc.content)
			d := source.Definition{Kind: tc.kind, Name: tc.defName, Path: path}

			got, err := ForTarget(tc.target, features)(d)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

// TestDefinitionGatesContent checks that a definition gates optional
// content per kind with text/template. Each case asserts the exact output
// with the flag on and off. The skill cases use the plain
// {{if .orchestration}} form, and the doc and agent cases use trim markers.
func TestDefinitionGatesContent(t *testing.T) {
	cases := []struct {
		name     string
		kind     source.Kind
		defName  string
		target   string
		rel      string
		content  string
		features Features
		want     string
	}{
		{
			name:     "skill off",
			kind:     source.Skill,
			defName:  "prose-editor",
			target:   source.Claude,
			rel:      filepath.Join("skills", "prose-editor", "SKILL.md"),
			content:  "# prose-editor\n\n- read prose{{if .orchestration}}\n- orchestrate prose{{end}}\n- write prose\n",
			features: Features{},
			want:     "# prose-editor\n\n- read prose\n- write prose\n",
		},
		{
			name:     "skill on",
			kind:     source.Skill,
			defName:  "prose-editor",
			target:   source.Claude,
			rel:      filepath.Join("skills", "prose-editor", "SKILL.md"),
			content:  "# prose-editor\n\n- read prose{{if .orchestration}}\n- orchestrate prose{{end}}\n- write prose\n",
			features: Features{Orchestration: true},
			want:     "# prose-editor\n\n- read prose\n- orchestrate prose\n- write prose\n",
		},
		{
			name:     "doc off",
			kind:     source.Doc,
			defName:  "prose",
			target:   source.Opencode,
			rel:      filepath.Join("docs", "zpecs", "prose.md"),
			content:  "# Prose guidelines\n\n- Be minimal\n{{- if .orchestration}}\n- Orchestrate\n{{- end}}\n- Link content\n",
			features: Features{},
			want:     "# Prose guidelines\n\n- Be minimal\n- Link content\n",
		},
		{
			name:     "doc on",
			kind:     source.Doc,
			defName:  "prose",
			target:   source.Opencode,
			rel:      filepath.Join("docs", "zpecs", "prose.md"),
			content:  "# Prose guidelines\n\n- Be minimal\n{{- if .orchestration}}\n- Orchestrate\n{{- end}}\n- Link content\n",
			features: Features{Orchestration: true},
			want:     "# Prose guidelines\n\n- Be minimal\n- Orchestrate\n- Link content\n",
		},
		{
			name:     "agent off",
			kind:     source.Agent,
			defName:  "prose-editor",
			target:   source.Claude,
			rel:      filepath.Join("agents", "prose-editor.md"),
			content:  "---\nname: prose-editor\ndescription: Reviews prose.\n---\n\n- Review prose\n{{- if .orchestration}}\n- Orchestrate prose\n{{- end}}\n- Fix prose\n",
			features: Features{},
			want:     "---\nname: prose-editor\ndescription: Reviews prose.\n---\n\n- Review prose\n- Fix prose\n",
		},
		{
			name:     "agent on",
			kind:     source.Agent,
			defName:  "prose-editor",
			target:   source.Claude,
			rel:      filepath.Join("agents", "prose-editor.md"),
			content:  "---\nname: prose-editor\ndescription: Reviews prose.\n---\n\n- Review prose\n{{- if .orchestration}}\n- Orchestrate prose\n{{- end}}\n- Fix prose\n",
			features: Features{Orchestration: true},
			want:     "---\nname: prose-editor\ndescription: Reviews prose.\n---\n\n- Review prose\n- Orchestrate prose\n- Fix prose\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeDefinitionFile(t, tc.rel, tc.content)
			d := source.Definition{Kind: tc.kind, Name: tc.defName, Path: path}

			got, err := definition(d, tc.target, tc.features)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

// TestDefinitionGatedBlockLeavesNoBlankLineOrDanglingPunctuation checks the
// authoring convention for a disabled gated list: the markers sit without
// their own line, or they trim, so the output holds no blank line between
// the items and no empty list marker.
func TestDefinitionGatedBlockLeavesNoBlankLineOrDanglingPunctuation(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{
			name:    "plain markers",
			content: "# prose-editor\n\n- read prose{{if .orchestration}}\n- orchestrate prose{{end}}\n- write prose\n",
		},
		{
			name:    "trim markers",
			content: "# prose-editor\n\n- read prose\n{{- if .orchestration}}\n- orchestrate prose\n{{- end}}\n- write prose\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeDefinitionFile(t, filepath.Join("skills", "prose-editor", "SKILL.md"), tc.content)

			got, err := definition(source.Definition{Kind: source.Skill, Name: "prose-editor", Path: path}, source.Claude, Features{})
			require.NoError(t, err)

			require.Equal(t, "# prose-editor\n\n- read prose\n- write prose\n", got)
			require.NotContains(t, got, "- read prose\n\n")
			require.NotContains(t, got, "\n- \n")
			require.NotContains(t, got, "\n-\n")
		})
	}
}

// TestDefinitionRendersEscapedBrace checks that a definition writes a
// literal {{ with the {{"{{"}} escape.
func TestDefinitionRendersEscapedBrace(t *testing.T) {
	content := `# prose-editor

A literal {{"{{"}} opens a template action.
`
	path := writeDefinitionFile(t, filepath.Join("skills", "prose-editor", "SKILL.md"), content)

	got, err := definition(source.Definition{Kind: source.Skill, Name: "prose-editor", Path: path}, source.Claude, Features{})
	require.NoError(t, err)
	require.Equal(t, "# prose-editor\n\nA literal {{ opens a template action.\n", got)
}

func TestDefinitionGatedLinkDisappearsAndReturns(t *testing.T) {
	content := "See the docs.\n{{- if .orchestration}}\n[Orchestration](orchestration.md)\n{{- end}}\n"
	link := "[Orchestration](orchestration.md)"

	offPath := writeDefinitionFile(t, filepath.Join("skills", "prose-editor", "SKILL.md"), content)
	off, err := definition(source.Definition{Kind: source.Skill, Name: "prose-editor", Path: offPath}, source.Claude, Features{})
	require.NoError(t, err)
	require.NotContains(t, off, link)
	require.Equal(t, "See the docs.\n", off)

	onPath := writeDefinitionFile(t, filepath.Join("skills", "prose-editor", "SKILL.md"), content)
	on, err := definition(source.Definition{Kind: source.Skill, Name: "prose-editor", Path: onPath}, source.Claude, Features{Orchestration: true})
	require.NoError(t, err)
	require.Contains(t, on, link)
	require.Equal(t, "See the docs.\n"+link+"\n", on)
}

func TestDefinitionReportsInvalidTemplate(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("skills", "prose-editor", "SKILL.md"), "{{if .orchestration}}unterminated\n")

	_, err := definition(source.Definition{Kind: source.Skill, Name: "prose-editor", Path: path}, source.Claude, Features{})

	require.ErrorContains(t, err, "rendering")
	require.Contains(t, err.Error(), path)
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

func TestParseReadsNameDescriptionAndTools(t *testing.T) {
	content := `---
name: prose-editor
description: "Reviews prose: checks it against the guidelines, and fixes it."
tools:
  - read
  - edit
---

Body text.
`
	want := fields{
		Name:        "prose-editor",
		Description: "Reviews prose: checks it against the guidelines, and fixes it.",
		Tools:       []string{"read", "edit"},
	}

	got, err := parse(content)
	require.NoError(t, err)
	require.Equal(t, want, got.fields)
}

func TestParseReadsInlineTools(t *testing.T) {
	got, err := parse("---\nname: prose-editor\ntools: [read, edit]\n---\n")
	require.NoError(t, err)
	require.Equal(t, []string{"read", "edit"}, got.fields.Tools)
}

func TestParseReadsMode(t *testing.T) {
	got, err := parse("---\nname: code-architect\nmode: primary\n---\n\nPlan the work.\n")
	require.NoError(t, err)
	require.Equal(t, "primary", got.fields.Mode)
}

func TestParseYieldsZeroFieldsWithoutFrontmatter(t *testing.T) {
	got, err := parse("# Just a body\n")
	require.NoError(t, err)
	require.Equal(t, fields{}, got.fields)
}

func TestParseErrorsOnUnterminatedFrontmatter(t *testing.T) {
	_, err := parse("---\nname: prose-editor\n")
	require.ErrorContains(t, err, "unterminated frontmatter")
}

func TestParseErrorsOnInvalidFrontmatter(t *testing.T) {
	_, err := parse("---\ntools: [read, edit\n---\n")
	require.Error(t, err)
}

func TestParseReadsBody(t *testing.T) {
	content := `---
name: prose-editor
description: Reviews prose.
---

You are a prose editor. Review the prose against the guidelines.
`
	got, err := parse(content)
	require.NoError(t, err)
	want := "You are a prose editor. Review the prose against the guidelines."
	require.Equal(t, want, got.body)
}

func TestParseKeepsBodyLines(t *testing.T) {
	content := `---
name: prose-editor
---

You are a prose editor.

Give numbered instructions.
`
	got, err := parse(content)
	require.NoError(t, err)
	want := "You are a prose editor.\n\nGive numbered instructions."
	require.Equal(t, want, got.body)
}

func TestParseYieldsWholeContentAsBodyWithoutFrontmatter(t *testing.T) {
	got, err := parse("You are a prose editor.\n")
	require.NoError(t, err)
	require.Equal(t, "You are a prose editor.", got.body)
}

func TestParseYieldsEmptyBody(t *testing.T) {
	got, err := parse("---\nname: prose-editor\n---\n")
	require.NoError(t, err)
	require.Equal(t, "", got.body)
}

func TestParseYieldsZeroFieldsForEmptyFile(t *testing.T) {
	got, err := parse("")
	require.NoError(t, err)
	require.Equal(t, fields{}, got.fields)
	require.Equal(t, "", got.body)
}
