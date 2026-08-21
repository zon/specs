Prompt prose reviews to fix issues in place

The prose review scope now prompts opencode to fix each prose issue it
finds immediately, editing the offending text in place rather than
writing issues to a project. VERSION bumps to 0.7.0.

Tests added:
- prompt coverage of the in-place fix behavior in internal/opencode
- updated record-line helpers in internal/testutil

Ralph item 6 (Prose Review) completed
