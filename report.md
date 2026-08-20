Inline the checkDir helper in internal/source

ReadKinds calls os.Stat directly and the one-line helper is gone.
The matching plan leaves refactoring.md. VERSION bumps to 0.1.11.

Tests: TestReadKindsErrorsOnMissingSource now asserts os.ErrNotExist. TestReadKindsSelectsSingleKind uses require.Len.

Ralph item 0 completed
