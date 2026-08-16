Make the agent's body the rendered system prompt

The frontmatter reader returns the body after the closing frontmatter
line along with the fields. The claude and opencode renders write that
body as the prompt.

Added tests cover the body after the frontmatter, intact body lines, a
whole-content body without frontmatter, an empty body, and the fields
reading with the new return.

Ralph item 9 completed
