package render

import (
	"fmt"
	"strings"

	"github.com/zon/specs/internal/frontmatter"
)

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

// OpencodeAgent uses subagent mode and drops the name.
func OpencodeAgent(fields frontmatter.Fields, body string) string {
	return fmt.Sprintf("---\nmode: subagent\ndescription: %s\n---\n\n%s\n", fields.Description, body)
}
