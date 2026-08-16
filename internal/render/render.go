package render

import (
	"fmt"

	"github.com/zon/specs/internal/frontmatter"
)

// ClaudeAgent renders an agent definition for the claude target, keeping
// the name and description with the body as prompt.
func ClaudeAgent(fields frontmatter.Fields, body string) string {
	return fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n%s\n", fields.Name, fields.Description, body)
}
