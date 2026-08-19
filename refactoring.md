# Refactoring Plans

## State the legacy-manifest rule once

`internal/targetdir/targetdir.go` explains how to handle a manifest without kinds in four comments: on `known`, `Owned`, `SaveOwned`, and `RemoveStale`. State it on `known` and shorten the rest.

## Remove the dead guard in frontmatter.parse

`internal/frontmatter/frontmatter.go:40` checks `len(lines) == 0`, but `strings.Split` always returns one element. Delete the guard.

