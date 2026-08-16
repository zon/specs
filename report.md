Render a claude agent keeping its name

A render module maps definitions to each target's format. For claude, an
agent keeps its name and description in the frontmatter, with its body as
the prompt. It reads the same frontmatter fields every target reads, so the
name carries over from the definition unchanged.

Added tests cover the claude agent keeping its name and the rendered agent
having the name, description, and body in order.

Ralph item 5 completed
