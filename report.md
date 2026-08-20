Split the CLI entry point from the update orchestration

Split cmd/zpecs into internal/cli (entry point, kong grammar, argument
parsing) and internal/update (orchestrated update pipeline). Recorded
both components in specs/architecture.yaml and updated the cmd/zpecs and
testutil entries. Removed the completed plan from refactoring.md and
updated the next plan's references. The CLI tests moved beside
internal/cli. The orchestration tests moved into internal/update's test
file. They drive the pipeline through update.Run. Added a shared
WriteSourceFile helper to internal/testutil.

Ralph item 3 completed
