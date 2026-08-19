Drop the dead empty-target case in report.Summary

- Summary no longer handles an empty target, since every caller passes one. Only the docs target omits it from the line.
- Remove the test row that covered the empty-target case. The remaining rows cover both branches of the condition.
- Remove the completed plan from refactoring.md.
- go test ./... and just check both pass.

Ralph item 1 completed
