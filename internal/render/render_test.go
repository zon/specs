package render

import (
	"strings"
	"testing"

	"github.com/zon/specs/internal/frontmatter"
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

func TestOpencodeAgentUsesSubagentMode(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor"}

	got := OpencodeAgent(fields, "Review prose.")

	if !strings.Contains(got, "mode: subagent") {
		t.Fatalf("OpencodeAgent rendered %q without mode: subagent", got)
	}
}

func TestOpencodeAgentDropsName(t *testing.T) {
	fields := frontmatter.Fields{Name: "prose-editor"}

	got := OpencodeAgent(fields, "Review prose.")

	if strings.Contains(got, "name:") {
		t.Fatalf("OpencodeAgent rendered %q with a name field", got)
	}
}

func TestOpencodeAgentRendersModeDescriptionAndBody(t *testing.T) {
	fields := frontmatter.Fields{Description: "Reviews prose against the guidelines."}

	got := OpencodeAgent(fields, "Review prose against the guidelines.")
	want := "---\nmode: subagent\ndescription: Reviews prose against the guidelines.\n---\n\nReview prose against the guidelines.\n"

	if got != want {
		t.Fatalf("OpencodeAgent = %q, want %q", got, want)
	}
}
