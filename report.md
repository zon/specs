Deny every tool an opencode agent does not list

The opencode render maps the definition's tools to a permission block
denying every other tool. An agent without tools omits the block.

Added tests cover the denied tools, the frontmatter order, and the no-tools
case.

Ralph item 8 completed
