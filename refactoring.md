# Refactoring Plans

## Inline the checkDir helper

`internal/source/source.go:62` `checkDir` wraps `os.Stat` in a one-line helper called once. Call `os.Stat` in `ReadKinds` and delete the helper.

## Simplify the delimiter search in frontmatter.parse

`internal/frontmatter/frontmatter.go:43` finds the closing `---`. It derives the body start as `i + 2` and slices `lines[1:bodyAt-1]`. Track the closing delimiter's index and slice the block and body around it.

## Drop the targetName wrapper type

`internal/cli/cli.go` `targetName` and the constants `targetClaude` and `targetOpencode` repeat `target.Claude` and `target.Opencode`. kong validates a plain string against `enum`, so the type only adds conversions. Use the target package constants and a string field. Update the tests that use the type.

## Keep orchestration bodies free of format details

The orchestration in `internal/update` still has `updatePair` wrapping `RemoveStale` with `fmt.Errorf("removing stale definitions: %w", ...)` and passing `os.Stdout` to `report.Summary`. `pairs` still embeds the display names as literals. Let `targetdir` wrap its own error, let the report module own the output sink and scope labels, and drop `os.Stdout` from orchestration calls.

## Move CLI tests out of the orchestration module's test file

`cmd/zpecs/main_test.go` defines implementation test helpers (`buildBinary`, `parseUpdateArgs`, `writeSourceFile`, `gitCloneSource`, `captureStdout`, `captureStderr`) and tests CLI plumbing (`TestUnmarshalScope`, `TestParseUpdate`, `TestPrintError*`, `TestBuildProducesRunnableCLIBinary`, `TestBinaryPrintsVersion`). These belong beside the `internal/cli` entry point once the split lands. Keep only tests of `update`, `updatePair`, and `resolveSource` decisions in the orchestration module's test file.

## Extract format details from the update orchestration tests

The update module's test file embeds output paths and rendered-content literals in test bodies. `orchestration.md:21` says to extract these into named test helpers. `orchestration.md:76` says the orchestration module's test file must never define helpers itself. Move the fixture and assertion helpers into `internal/testutil` or the implementation module that owns each detail. Move the per-target render assertions (e.g. the name/description checks in `TestUpdateWritesAgentUnderSourceNameForBothTargets`) into `internal/render`'s tests. Keep the plan narrow and self-contained.
