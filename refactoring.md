# Refactoring Plans

## Inline the checkDir helper

`internal/source/source.go:93` `checkDir` wraps `os.Stat` in a one-line helper called once. Call `os.Stat` in `ReadKinds` and delete the helper.

## Simplify the delimiter search in frontmatter.parse

`internal/frontmatter/frontmatter.go:43` finds the closing `---`. It derives the body start as `i + 2` and slices `lines[1:bodyAt-1]`. Track the closing delimiter's index and slice the block and body around it.

## Move the orchestration tests into the update module's test file

`cmd/zpecs/main_test.go` defines implementation test helpers (`buildBinary`, `parseUpdateArgs`, `writeSourceFile`, `gitCloneSource`, `captureStdout`, `captureStderr`) and tests CLI plumbing (`TestParseUpdate`, `TestPrintError*`, `TestBuildProducesRunnableCLIBinary`, `TestBinaryPrintsVersion`). These stay beside the `cmd/zpecs` entry point. The resolveSource tests already moved with the function. Move the tests of the `update` and `updatePair` decisions into the orchestration module's test file.
