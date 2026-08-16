Render an opencode agent in subagent mode without its name

The render module now maps an agent definition to the opencode target.
An opencode agent keeps its description and body, and uses
mode: subagent instead of a name field. It reads the same frontmatter
fields as the claude render, dropping only the name.

Added tests cover the rendered agent using mode: subagent, the rendered
agent having no name field, and the description and body in order.

Ralph item 6 completed
