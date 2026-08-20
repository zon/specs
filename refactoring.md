# Refactoring Plans

## Inline the checkDir helper

`internal/source/source.go:62` `checkDir` wraps `os.Stat` in a one-line helper called once. Call `os.Stat` in `ReadKinds` and delete the helper.

## Simplify the delimiter search in frontmatter.parse

`internal/frontmatter/frontmatter.go:43` finds the closing `---`. It derives the body start as `i + 2` and slices `lines[1:bodyAt-1]`. Track the closing delimiter's index and slice the block and body around it.

## Drop the targetName wrapper type

`cmd/zpecs/main.go:59` `targetName` and the constants `targetClaude` and `targetOpencode` repeat `target.Claude` and `target.Opencode`. kong validates a plain string against `enum`, so the type only adds conversions. Use the target package constants and a string field. Update the tests that use the type.

## Split the update orchestration from the entry point

`cmd/zpecs` mixes entry-point code with orchestration. The kong grammar, `main`, `run`, `usageText`, `printError`, `scope`, and `cliVars` are an implementation concern (architecture.md:42). They stay in `cmd/zpecs`, which keeps a real concern rather than passing through to another module. `update`, `updatePair`, `resolveSource`, and `pairs` are coordination. Move the coordination into an orchestration module (e.g. `internal/update`) and record it in `specs/architecture.yaml`.

## Keep orchestration bodies free of format details

The orchestration in `cmd/zpecs` builds strings and passes infrastructure values: `main.go:211` wraps `RemoveStale` with `fmt.Errorf("removing stale definitions: %w", ...)`, `main.go:221` passes `os.Stdout` to `report.Summary`, and `pairs` at `main.go:82` embeds the display names as literals. Let `targetdir` wrap its own error, let the report module own the output sink and scope labels, and drop `os.Stdout` from orchestration calls.

## Move the orchestration tests into the update module's test file

`cmd/zpecs/main_test.go` defines implementation test helpers (`buildBinary`, `parseUpdateArgs`, `writeSourceFile`, `gitCloneSource`, `captureStdout`, `captureStderr`) and tests CLI plumbing (`TestUnmarshalScope`, `TestParseUpdate`, `TestPrintError*`, `TestBuildProducesRunnableCLIBinary`, `TestBinaryPrintsVersion`). These stay beside the `cmd/zpecs` entry point. Move only the tests of `update`, `updatePair`, and `resolveSource` decisions into the orchestration module's test file.
