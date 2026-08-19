# Refactoring Plans

## Split the fromAST state machine

`internal/spec/spec.go:55` `fromAST` walks the AST, tracks title, purpose, and requirement state, and fills bodies and steps. It is nested four levels deep. Split it into `applyHeading` and `applyContent` over a small state struct.

## Drop the dead target case in Summary

`internal/report/report.go:12` `Summary` handles an empty target, but every pair sets one. Keep only the docs case and remove the matching test row.

## Drop the unused bool from Write

`internal/targetdir/targetdir.go:87` `Write` returns whether it wrote the file, but `WriteAll` ignores it. Return only an error and always record the path in `owned`. Rewrite the tests that assert on the bool.

## State the legacy-manifest rule once

`internal/targetdir/targetdir.go` explains how to handle a manifest without kinds in five comments: on `known`, `manifestName`, `Owned`, `SaveOwned`, and `RemoveStale`. State it once on `known` and shorten the rest.

## Remove the dead guard in frontmatter.parse

`internal/frontmatter/frontmatter.go:40` checks `len(lines) == 0`, but `strings.Split` always returns one element. Delete the guard.

## Move URL detection out of the orchestration module

`cmd/zpecs/main.go:230` `resolveSource` checks for `"://"` inline. Add `IsRemote` to `internal/clone` and call it, keeping `main.go` pure delegation.

## Replace the kindName and parseKind switches with a map

`internal/targetdir/targetdir.go:172` `kindName` and `parseKind` are mirror switches over the same words. Replace both with one map and two lookups.

## Share the git fixture helpers

`cmd/zpecs/main_test.go:224`, `internal/clone/clone_test.go:12`, and `internal/repo/repo_test.go:12` each build a temp git repo. Extract a `testutil` package with `GitRepo` and `RunGit` helpers.
