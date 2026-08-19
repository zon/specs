# Refactoring

## Store the kind in the ownership manifest

The `.zpecs` manifest lists written paths. `RemoveStale` recovers each
path's kind with `pathKind`, which matches path strings. Store the kind
with each path instead. Write one `kind path` line per file. Read the
kind back in `RemoveStale`. Delete `pathKind`.

A manifest written before the change has no kinds. Skip those lines in
`RemoveStale`. The next full run rewrites the manifest in the new
format.

Scope: `internal/targetdir/targetdir.go` and its test.
Verify: `go test ./internal/targetdir/`.

## Run one pipeline per target

`update` runs two passes for the `all` scope. `updateScope` treats docs
as a special target. Give each kind a home target instead. Skills and
agents belong to the `--target` runner. Docs belong to `docs/zpecs`.
Turn a scope into a list of target and kind pairs. Run the sync pipeline
once per pair.

Delete the `scopeAll` branch in `update` and the docs cases in
`updateScope`. Keep the current report text.

Scope: `cmd/zpecs/main.go`.
Verify: `go test ./...`.
