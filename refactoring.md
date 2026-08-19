# Refactoring Plans

## Remove the dead guard in frontmatter.parse

`internal/frontmatter/frontmatter.go:40` checks `len(lines) == 0`, but `strings.Split` always returns one element.
