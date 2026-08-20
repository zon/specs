Move update orchestration into internal/update

Move the update coordination out of cmd/zpecs into a new internal/update
orchestration module and record it in specs/architecture.yaml. cmd/zpecs
keeps the entry point, kong grammar, and argument parsing. The scope
vocabulary moved into internal/source as source.Scope, so it is defined
once, beside the kinds it selects, per the architecture review. Remove
the completed plan from refactoring.md and update the remaining plans'
references.

Tests: move TestUnmarshalScope to internal/source and the resolveSource
tests to internal/update. CLI parser and plumbing tests stay beside the
entry point. Format-detail cleanup (report sink, stale-removal error
wrapping, pair labels, render closure) stays planned in refactoring.md
for a later item.

Ralph item 3 completed
