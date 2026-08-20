package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/zon/specs/internal/source"
	"github.com/zon/specs/internal/target"
)

// definition returns a definition's text for a target. Skills and docs
// return their contents verbatim. It parses and renders agents.
func definition(d source.Definition, targetName string) (string, error) {
	if d.Kind == source.Skill || d.Kind == source.Doc {
		raw, err := os.ReadFile(d.Path)
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", d.Path, err)
		}
		return string(raw), nil
	}
	content, err := read(d.Path)
	if err != nil {
		return "", err
	}
	if targetName == target.Claude {
		return claudeAgent(content.fields, content.body), nil
	}
	return opencodeAgent(content.fields, content.body)
}

// ForTarget returns the function that renders each definition for a target.
func ForTarget(targetName string) func(source.Definition) (string, error) {
	return func(d source.Definition) (string, error) {
		return definition(d, targetName)
	}
}

// claudeAgent keeps the name, description, and tools. It uses the body
// as the prompt.
func claudeAgent(fields fields, body string) string {
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

// opencodeAgent writes the definition's mode, defaulting to subagent.
// It drops the name and denies every unlisted tool.
func opencodeAgent(fields fields, body string) (string, error) {
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

// frame wraps the frontmatter lines and body between `---` delimiters.
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
