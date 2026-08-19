# Refactoring Plans

## State the legacy-manifest rule once

`internal/targetdir/targetdir.go` explains how to handle a manifest without kinds in four comments: on `known`, `Owned`, `SaveOwned`, and `RemoveStale`. State it on `known` and shorten the rest.

## Remove the dead guard in frontmatter.parse

`internal/frontmatter/frontmatter.go:40` checks `len(lines) == 0`, but `strings.Split` always returns one element. Delete the guard.

## Replace the kindName and parseKind switches with a map

`internal/targetdir/targetdir.go:170` `kindName` and `parseKind` are mirror switches over the same words. Replace both with one map and two lookups.
