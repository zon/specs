package render

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/zon/specs/internal/source"
)

// Features names the optional feature areas a run enables.
type Features map[string]bool

// Orchestration is the feature name for orchestration-only content.
const Orchestration = "orchestration"

// Enabled returns the feature names a run enables. Every feature
// defaults to off.
func Enabled(orchestration bool) Features {
	if !orchestration {
		return Features{}
	}
	return Features{Orchestration: true}
}

// definition returns a definition's text for a target. It reads the
// definition file and renders it through the template with the enabled
// features. Skills and docs return the rendered text. It parses and frames
// agents.
func definition(d source.Definition, targetName string, features Features) (string, error) {
	raw, err := os.ReadFile(d.Path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", d.Path, err)
	}
	text, err := renderTemplate(string(raw), d.Path, features)
	if err != nil {
		return "", err
	}
	if d.Kind == source.Skill || d.Kind == source.Doc {
		return text, nil
	}
	content, err := parse(text)
	if err != nil {
		return "", fmt.Errorf("parsing %s: %w", d.Path, err)
	}
	if targetName == source.Claude {
		return claudeAgent(content.fields, content.body), nil
	}
	return opencodeAgent(content.fields, content.body)
}

// renderTemplate executes raw as a text/template with features as data. A
// missing feature is false. A definition gates optional content with
// {{if .orchestration}} ... {{end}}. The markers sit without their own
// line, or trim with {{- and -}}, so a disabled block leaves no blank
// line or dangling list punctuation. A definition writes a literal {{ with
// the {{"{{"}} escape. A template without actions renders verbatim.
func renderTemplate(raw, path string, features Features) (string, error) {
	tmpl, err := template.New("definition").Parse(raw)
	if err != nil {
		return "", fmt.Errorf("rendering %s: %w", path, err)
	}
	var out strings.Builder
	if err := tmpl.Execute(&out, features); err != nil {
		return "", fmt.Errorf("rendering %s: %w", path, err)
	}
	return out.String(), nil
}

// ForTarget returns the function that renders each definition for a target.
func ForTarget(targetName string, features Features) func(source.Definition) (string, error) {
	return func(d source.Definition) (string, error) {
		return definition(d, targetName, features)
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
