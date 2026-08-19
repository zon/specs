# Refactoring

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
