package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zon/specs/internal/frontmatter"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/targetdir"
)

func TestClaudeAgentKeepsName(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor"}

	got := ClaudeAgent(fields, "Review prose.")

	if !strings.Contains(got, "name: prose-editor") {
		t.Fatalf("ClaudeAgent rendered %q without the definition's name", got)
	}
}

func TestClaudeAgentRendersNameDescriptionAndBody(t *testing.T) {
	fields := frontmatter.Fields{
		Name:        "prose-editor",
		Description: "Reviews prose against the guidelines.",
	}

	got := ClaudeAgent(fields, "Review prose against the guidelines.")
	want := "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"

	if got != want {
		t.Fatalf("ClaudeAgent = %q, want %q", got, want)
	}
}

func TestClaudeAgentListsTools(t *testing.T) {
	fields := frontmatter.Fields{Tools: []string{"read", "edit"}}

	got := ClaudeAgent(fields, "Review prose.")

	if !strings.Contains(got, "tools:\n  - read\n  - edit") {
		t.Fatalf("ClaudeAgent rendered %q without the tools list", got)
	}
}

func TestClaudeAgentRendersToolsAfterDescription(t *testing.T) {
	fields := frontmatter.Fields{
		Name:        "prose-editor",
		Description: "Reviews prose.",
		Tools:       []string{"read", "edit"},
	}

	got := ClaudeAgent(fields, "Review prose.")
	want := "---\nname: prose-editor\ndescription: Reviews prose.\ntools:\n  - read\n  - edit\n---\n\nReview prose.\n"

	if got != want {
		t.Fatalf("ClaudeAgent = %q, want %q", got, want)
	}
}

func TestClaudeAgentOmitsToolsWhenEmpty(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor"}

	got := ClaudeAgent(fields, "Review prose.")

	if strings.Contains(got, "tools:") {
		t.Fatalf("ClaudeAgent rendered %q with a tools field for an empty list", got)
	}
}

func TestOpencodeAgentDefaultsToSubagentMode(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor"}

	got, err := OpencodeAgent(fields, "Review prose.")
	if err != nil {
		t.Fatalf("OpencodeAgent: %v", err)
	}

	if !strings.Contains(got, "mode: subagent") {
		t.Fatalf("OpencodeAgent rendered %q without mode: subagent", got)
	}
}

func TestOpencodeAgentUsesDefinitionMode(t *testing.T) {
	fields := frontmatter.Fields{Name: "code-architect", Mode: "primary"}

	got, err := OpencodeAgent(fields, "Plan the work.")
	if err != nil {
		t.Fatalf("OpencodeAgent: %v", err)
	}

	if !strings.Contains(got, "mode: primary") {
		t.Fatalf("OpencodeAgent rendered %q without mode: primary", got)
	}
}

func TestOpencodeAgentRejectsUnknownMode(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor", Mode: "banana"}

	if _, err := OpencodeAgent(fields, "Review prose."); err == nil {
		t.Fatal("OpencodeAgent accepted an unknown mode")
	}
}

func TestOpencodeAgentDropsName(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor"}

	got, err := OpencodeAgent(fields, "Review prose.")
	if err != nil {
		t.Fatalf("OpencodeAgent: %v", err)
	}

	if strings.Contains(got, "name:") {
		t.Fatalf("OpencodeAgent rendered %q with a name field", got)
	}
}

func TestOpencodeAgentRendersModeDescriptionAndBody(t *testing.T) {
	fields := frontmatter.Fields{Description: "Reviews prose against the guidelines."}

	got, err := OpencodeAgent(fields, "Review prose against the guidelines.")
	if err != nil {
		t.Fatalf("OpencodeAgent: %v", err)
	}
	want := "---\nmode: subagent\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"

	if got != want {
		t.Fatalf("OpencodeAgent = %q, want %q", got, want)
	}
}

func TestOpencodeAgentDeniesEveryOtherTool(t *testing.T) {
	fields := frontmatter.Fields{Tools: []string{"read", "edit"}}

	got, err := OpencodeAgent(fields, "Review prose.")
	if err != nil {
		t.Fatalf("OpencodeAgent: %v", err)
	}

	for _, denied := range []string{"bash", "write", "grep", "glob"} {
		if !strings.Contains(got, denied+": deny") {
			t.Fatalf("OpencodeAgent rendered %q without denying %s", got, denied)
		}
	}
}

func TestOpencodeAgentDoesNotDenyListedTools(t *testing.T) {
	fields := frontmatter.Fields{Tools: []string{"read", "edit"}}

	got, err := OpencodeAgent(fields, "Review prose.")
	if err != nil {
		t.Fatalf("OpencodeAgent: %v", err)
	}

	if strings.Contains(got, "read: deny") || strings.Contains(got, "edit: deny") {
		t.Fatalf("OpencodeAgent rendered %q denying a listed tool", got)
	}
}

func TestOpencodeAgentRendersDenyRulesAfterDescription(t *testing.T) {
	fields := frontmatter.Fields{
		Description: "Reviews prose.",
		Tools:       []string{"read", "edit"},
	}

	got, err := OpencodeAgent(fields, "Review prose.")
	if err != nil {
		t.Fatalf("OpencodeAgent: %v", err)
	}
	want := "---\nmode: subagent\ndescription: Reviews prose.\npermission:\n" +
		"  apply_patch: deny\n  bash: deny\n  glob: deny\n  grep: deny\n" +
		"  lsp: deny\n  question: deny\n  skill: deny\n  todowrite: deny\n" +
		"  webfetch: deny\n  websearch: deny\n  write: deny\n" +
		"---\n\nReview prose.\n"

	if got != want {
		t.Fatalf("OpencodeAgent = %q, want %q", got, want)
	}
}

func TestOpencodeAgentOmitsDenyRulesWhenToolsEmpty(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor"}

	got, err := OpencodeAgent(fields, "Review prose.")
	if err != nil {
		t.Fatalf("OpencodeAgent: %v", err)
	}

	if strings.Contains(got, "permission:") {
		t.Fatalf("OpencodeAgent rendered %q with deny rules for an empty tools list", got)
	}
}

func TestDefinitionReturnsSkillVerbatim(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("skills", "prose-editor", "SKILL.md"), "# prose-editor\n\nReview prose.\n")

	got, err := Definition(source.Definition{Kind: source.Skill, Name: "prose-editor", Path: path}, targetdir.Claude)
	if err != nil {
		t.Fatalf("Definition: %v", err)
	}
	want := "# prose-editor\n\nReview prose.\n"
	if got != want {
		t.Fatalf("Definition = %q, want %q", got, want)
	}
}

func TestDefinitionReturnsDocVerbatim(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("docs", "zpecs", "prose.md"), "# Prose guidelines\n")

	got, err := Definition(source.Definition{Kind: source.Doc, Name: "prose", Path: path}, targetdir.Opencode)
	if err != nil {
		t.Fatalf("Definition: %v", err)
	}
	want := "# Prose guidelines\n"
	if got != want {
		t.Fatalf("Definition = %q, want %q", got, want)
	}
}

func TestDefinitionRendersAgentForClaude(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n")

	got, err := Definition(source.Definition{Kind: source.Agent, Name: "prose-editor", Path: path}, targetdir.Claude)
	if err != nil {
		t.Fatalf("Definition: %v", err)
	}
	want := "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"
	if got != want {
		t.Fatalf("Definition = %q, want %q", got, want)
	}
}

func TestDefinitionRendersAgentForOpencode(t *testing.T) {
	path := writeDefinitionFile(t, filepath.Join("agents", "prose-editor.md"), "---\nname: prose-editor\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n")

	got, err := Definition(source.Definition{Kind: source.Agent, Name: "prose-editor", Path: path}, targetdir.Opencode)
	if err != nil {
		t.Fatalf("Definition: %v", err)
	}
	want := "---\nmode: subagent\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"
	if got != want {
		t.Fatalf("Definition = %q, want %q", got, want)
	}
}

func TestDefinitionReportsUnreadableAgentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agents", "missing.md")

	_, err := Definition(source.Definition{Kind: source.Agent, Name: "missing", Path: path}, targetdir.Opencode)
	if err == nil {
		t.Fatal("Definition accepted a missing agent file")
	}
	if !strings.Contains(err.Error(), "reading") {
		t.Fatalf("Definition error %q does not name the file it could not read", err)
	}
}

// writeDefinitionFile writes content at rel under a fresh temp dir and
// returns the path.
func writeDefinitionFile(t *testing.T, rel, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
