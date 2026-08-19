# Refactoring

## Align testing.md with the standard library

Update the doc to describe the standard library style: `t.Fatalf`, `t.Helper`, `t.Run`, and table-driven cases. Drop the `testify` requirement and the `require`/`assert` example.

Scope: `docs/testing.md`.
Verify: `go test ./...`.
