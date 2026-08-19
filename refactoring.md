# Refactoring Plans

## Drop the dead target case in Summary

`internal/report/report.go:12` `Summary` handles an empty target, but every pair sets one. Keep only the docs case and remove the matching test row.

## State the legacy-manifest rule once

`internal/targetdir/targetdir.go` explains how to handle a manifest without kinds in four comments: on `known`, `Owned`, `SaveOwned`, and `RemoveStale`. State it once on `known` and shorten the rest.

## Remove the dead guard in frontmatter.parse

`internal/frontmatter/frontmatter.go:40` checks `len(lines) == 0`, but `strings.Split` always returns one element. Delete the guard.

## Replace the kindName and parseKind switches with a map

`internal/targetdir/targetdir.go:170` `kindName` and `parseKind` are mirror switches over the same words. Replace both with one map and two lookups.
