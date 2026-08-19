package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/zon/specs/internal/frontmatter"
	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/target"
)

// Definition returns the text a definition renders to for a target.
// Skills and docs return their file contents verbatim; agents are
// parsed and rendered for the target.
func Definition(d source.Definition, targetName string) (string, error) {
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
	if targetName == target.Claude {
		return ClaudeAgent(content.Fields, content.Body), nil
	}
	return OpencodeAgent(content.Fields, content.Body)
}

// ClaudeAgent keeps the name, description, and tools, and uses the body as
// the prompt.
func ClaudeAgent(fields frontmatter.Fields, body string) string {
	var lines []string
	lines = append(lines, "name: "+fields.Name, "description: "+fields.Description)
	if len(fields.Tools) > 0 {
		lines = append(lines, "tools:")
		for _, tool := range fields.Tools {
			lines = append(lines, "  - "+tool)
		}
	}
	return frame(lines, body)
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
	var lines []string
	lines = append(lines, "mode: "+mode, "description: "+fields.Description)
	if len(fields.Tools) > 0 {
		lines = append(lines, "permission:")
		for _, tool := range deniedTools(fields.Tools) {
			lines = append(lines, "  "+tool+": deny")
		}
	}
	return frame(lines, body), nil
}

// frame wraps the frontmatter lines and body in an agent file: the `---`
// delimiters, the lines, a blank line, and a trailing newline.
func frame(lines []string, body string) string {
	var out strings.Builder
	out.WriteString("---\n")
	for _, line := range lines {
		out.WriteString(line)
		out.WriteString("\n")
	}
	out.WriteString("---\n\n")
	out.WriteString(body)
	out.WriteString("\n")
	return out.String()
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
