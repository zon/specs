Drop the dead guard in frontmatter.parse

strings.Split always returns at least one line, so the len(lines) == 0
check in parse can never fire. Remove it. Add a test that an empty
file parses to zero fields and an empty body. Remove the completed
plan from refactoring.md. Delete the file, since no plans remain.

Ralph item 4 completed
