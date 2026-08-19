# Refactoring

## Collapse the source readers

`readSkills`, `readAgents`, and `readDocs` build definitions the same way. Only the glob pattern and the name derivation differ. Collapse them into one helper that takes the kind, the pattern, and a name function. A fourth kind then needs no fourth reader.

Scope: `internal/source/source.go`.
Verify: `go test ./...`.

## Share the agent render frame

`ClaudeAgent` and `OpencodeAgent` write the same frame around the body: the `---` lines, the frontmatter fields, a blank line, and a trailing newline. Only the fields differ. Extract that frame into a helper and have both renderers call it.

Scope: `internal/render/render.go`.
Verify: `go test ./...`.

## Align testing.md with the assertion library

`docs/testing.md` mandates `testify` for every assertion, but the suite asserts with the standard library's `testing` package. Update the doc to describe the standard library style: `t.Fatalf`, `t.Helper`, `t.Run`, and table-driven cases. Drop the `testify` requirement and the `require`/`assert` example.

Scope: `docs/testing.md`.
Verify: `go test ./...`.
