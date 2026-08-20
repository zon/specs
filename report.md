Simplify the delimiter search in frontmatter.parse

parse now tracks the closing delimiter's index and slices the block and
body around it. The existing frontmatter tests cover the behavior, so no
new tests were added. Removed the completed plan from refactoring.md.
Bumped VERSION to 0.1.11.

Ralph item 1 completed
