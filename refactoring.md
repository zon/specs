# Refactoring Plans

## Inline the checkDir helper

`internal/source/source.go:93` `checkDir` wraps `os.Stat` in a one-line helper called once. Call `os.Stat` in `ReadKinds` and delete the helper.

## Simplify the delimiter search in frontmatter.parse

`internal/frontmatter/frontmatter.go:43` finds the closing `---`. It derives the body start as `i + 2` and slices `lines[1:bodyAt-1]`. Track the closing delimiter's index and slice the block and body around it.

## Keep orchestration bodies free of format details

The orchestration in `internal/update` builds strings and passes infrastructure values: `update.go:81` wraps `RemoveStale` with `fmt.Errorf("removing stale definitions: %w", ...)`, `update.go:91` passes `os.Stdout` to `report.Summary`, and `pairs` at `update.go:53` embeds the display names as literals. `update.go:83` passes the render closure `func(d source.Definition) (string, error) { return render.Definition(d, p.target) }` to `targetdir.WriteAll`. Let `targetdir` wrap its own error, let the report module own the output sink and scope labels, and drop `os.Stdout` from orchestration calls.

## Move the orchestration tests into the update module's test file

`cmd/zpecs/main_test.go` defines implementation test helpers (`buildBinary`, `parseUpdateArgs`, `writeSourceFile`, `gitCloneSource`, `captureStdout`, `captureStderr`) and tests CLI plumbing (`TestParseUpdate`, `TestPrintError*`, `TestBuildProducesRunnableCLIBinary`, `TestBinaryPrintsVersion`). These stay beside the `cmd/zpecs` entry point. The resolveSource tests already moved with the function. Move the tests of the `update` and `updatePair` decisions into the orchestration module's test file. Extract the seed-file path literal `filepath.Join(dir, "seed")` that `TestResolveSourceClonesRemote` asserts into a helper imported from an implementation module, such as `testutil`, when executing the move.
