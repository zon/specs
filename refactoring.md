# Refactoring Plans

## Inline the checkDir helper

`internal/source/source.go:93` `checkDir` wraps `os.Stat` in a one-line helper called once. Call `os.Stat` in `ReadKinds` and delete the helper.

## Simplify the delimiter search in frontmatter.parse

`internal/frontmatter/frontmatter.go:43` finds the closing `---`. It derives the body start as `i + 2` and slices `lines[1:bodyAt-1]`. Track the closing delimiter's index and slice the block and body around it.
