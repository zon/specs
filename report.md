Move format details out of the update orchestration

internal/targetdir now wraps its RemoveStale error. internal/report
owns the output sink and the scope labels. The update orchestration
carries scope words instead of display labels and no longer passes
os.Stdout. Added a test for the wrapped RemoveStale error and one for
the "all" scope label. Removed the completed plan from refactoring.md.
Bumped VERSION to 0.1.9.

Ralph item 4 completed
