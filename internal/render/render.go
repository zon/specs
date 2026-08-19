package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/zon/specs/internal/frontmatter"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/targetdir"
)

// Definition returns the text a definition renders to for a target.
// Skills and docs return their file contents verbatim; agents are
// parsed and rendered for the target.
func Definition(d source.Definition, target string) (string, error) {
	if d.Kind == source.Skill || d.Kind == source.Doc {
		raw, err := os.ReadFile(d.Path)
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", d.Path, err)
		}
		return string(raw), nil
	}
	content, err := frontmatter.Read(d.Path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", d.Path, err)
	}
	if target == targetdir.Claude {
		return ClaudeAgent(content.Fields, content.Body), nil
	}
	return OpencodeAgent(content.Fields, content.Body)
}

// ClaudeAgent keeps the name, description, and tools, and uses the body as
// the prompt.
func ClaudeAgent(fields frontmatter.Fields, body string) string {
	var out strings.Builder
	out.WriteString("---\n")
	fmt.Fprintf(&out, "name: %s\n", fields.Name)
	fmt.Fprintf(&out, "description: %s\n", fields.Description)
	if len(fields.Tools) > 0 {
		out.WriteString("tools:\n")
		for _, tool := range fields.Tools {
			fmt.Fprintf(&out, "  - %s\n", tool)
		}
	}
	out.WriteString("---\n\n")
	out.WriteString(body)
	out.WriteString("\n")
	return out.String()
}

// OpencodeAgent writes the definition's mode, defaulting to subagent,
// drops the name, and denies every tool the definition does not list.
func OpencodeAgent(fields frontmatter.Fields, body string) (string, error) {
	mode := fields.Mode
	if mode == "" {
		mode = "subagent"
	}
	if !validModes[mode] {
		return "", fmt.Errorf("unknown mode %q", mode)
	}
	var out strings.Builder
	out.WriteString("---\n")
	fmt.Fprintf(&out, "mode: %s\n", mode)
	fmt.Fprintf(&out, "description: %s\n", fields.Description)
	if len(fields.Tools) > 0 {
		denied := deniedTools(fields.Tools)
		out.WriteString("permission:\n")
		for _, tool := range denied {
			fmt.Fprintf(&out, "  %s: deny\n", tool)
		}
	}
	out.WriteString("---\n\n")
	out.WriteString(body)
	out.WriteString("\n")
	return out.String(), nil
}

// validModes are the mode values an opencode agent may take.
var validModes = map[string]bool{
	"primary":  true,
	"subagent": true,
	"all":      true,
}

// opencodeTools are the tools an opencode agent can use.
var opencodeTools = []string{
	"apply_patch",
	"bash",
	"edit",
	"glob",
	"grep",
	"lsp",
	"question",
	"read",
	"skill",
	"todowrite",
	"webfetch",
	"websearch",
	"write",
}

// deniedTools returns every opencode tool not in the allowed list, in order.
func deniedTools(allowed []string) []string {
	set := make(map[string]bool, len(allowed))
	for _, tool := range allowed {
		set[tool] = true
	}
	var denied []string
	for _, tool := range opencodeTools {
		if !set[tool] {
			denied = append(denied, tool)
		}
	}
	return denied
}
