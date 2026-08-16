package render

import (
	"fmt"

	"github.com/zon/specs/internal/frontmatter"
)

// ClaudeAgent keeps the name and description, and prompts with the body.
func ClaudeAgent(fields frontmatter.Fields, body string) string {
	return fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n%s\n", fields.Name, fields.Description, body)
}

// OpencodeAgent uses subagent mode and drops the name.
func OpencodeAgent(fields frontmatter.Fields, body string) string {
	return fmt.Sprintf("---\nmode: subagent\ndescription: %s\n---\n\n%s\n", fields.Description, body)
}
