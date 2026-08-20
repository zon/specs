# Refactoring Plans

## Simplify the delimiter search in frontmatter.parse

`internal/frontmatter/frontmatter.go:43` finds the closing `---`. It derives the body start as `i + 2` and slices `lines[1:bodyAt-1]`. Track the closing delimiter's index and slice the block and body around it.

## Extract format details from the update orchestration tests

The update module's test file embeds output paths and rendered-content literals in test bodies. `orchestration.md:21` says to extract these into named test helpers. `orchestration.md:76` says the orchestration module's test file must never define helpers itself. Move the fixture and assertion helpers into `internal/testutil` or the implementation module that owns each detail. Move the per-target render assertions (e.g. the name/description checks in `TestUpdateWritesAgentUnderSourceNameForBothTargets`) into `internal/render`'s tests. Keep the plan narrow and self-contained.
