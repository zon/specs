Move format details out of the update orchestration

targetdir wraps its RemoveStale error. report owns the output sink and
the scope labels. The orchestration renders through a named function
instead of an inline closure. It no longer passes os.Stdout or builds
display names and error messages. Remove the matching plan from
refactoring.md.

Tests: the RemoveStale error-wrapping test in targetdir, the Summary
sink and label tests in report, and the updatePair sink test in update,
which captures the report through the testutil helper.

Ralph item 4 completed
