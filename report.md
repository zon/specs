Move the orchestration tests into the internal/update module

Move the tests of the update and updatePair decisions out of
cmd/zpecs/main_test.go and into the orchestration module's test file,
calling it with Options instead of driving the CLI.
cmd/zpecs keeps the CLI test helpers and the parser and binary tests.
testutil gains the shared fixtures the moved tests use: source
builders and written-file assertions. Remove the matching plan from
refactoring.md.

Tests: the moved update orchestration tests in internal/update, and
the new testutil fixture tests.

Ralph item 5 completed
