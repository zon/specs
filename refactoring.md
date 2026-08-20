# Refactoring Plans

## Inline the checkDir helper

`internal/source/source.go:62` `checkDir` wraps `os.Stat` in a one-line helper called once. Call `os.Stat` in `ReadKinds` and delete the helper.

## Simplify the delimiter search in frontmatter.parse

`internal/frontmatter/frontmatter.go:43` finds the closing `---`. It derives the body start as `i + 2` and slices `lines[1:bodyAt-1]`. Track the closing delimiter's index and slice the block and body around it.

## Drop the targetName wrapper type

`cmd/zpecs/main.go:59` `targetName` and the constants `targetClaude` and `targetOpencode` repeat `target.Claude` and `target.Opencode`. kong validates a plain string against `enum`, so the type only adds conversions. Use the target package constants and a string field. Update the tests that use the type.

## Split the CLI entry point from the orchestration

`cmd/zpecs` mixes entry-point code with orchestration. The kong grammar, `main`, `run`, `usageText`, `printError`, `scope`, and `cliVars` are an implementation concern (architecture.md:42), while `update`, `updatePair`, `resolveSource`, and `pairs` are coordination. Move the entry point into an implementation module (e.g. `internal/cli`) and move the coordination into an orchestration module (e.g. `internal/update`). Record both in `specs/architecture.yaml`.

## Keep orchestration bodies free of format details

The orchestration in `cmd/zpecs` builds strings and passes infrastructure values: `main.go:211` wraps `RemoveStale` with `fmt.Errorf("removing stale definitions: %w", ...)`, `main.go:221` passes `os.Stdout` to `report.Summary`, and `pairs` at `main.go:82` embeds the display names as literals. Let `targetdir` wrap its own error, let the report module own the output sink and scope labels, and drop `os.Stdout` from orchestration calls.

## Move CLI tests out of the orchestration module's test file

`cmd/zpecs/main_test.go` defines implementation test helpers (`buildBinary`, `parseUpdateArgs`, `writeSourceFile`, `gitCloneSource`, `captureStdout`, `captureStderr`) and tests CLI plumbing (`TestUnmarshalScope`, `TestParseUpdate`, `TestPrintError*`, `TestBuildProducesRunnableCLIBinary`, `TestBinaryPrintsVersion`). These belong beside the `internal/cli` entry point once the split lands. Keep only tests of `update`, `updatePair`, and `resolveSource` decisions in the orchestration module's test file.
