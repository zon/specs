Prompt code reviews into refactor projects

The code review scope now prompts opencode to write each code issue to a
refactor-<slug>.yaml project in projects/ rather than editing the code,
updating a matching existing project, and writing no project when it
finds no issues. VERSION bumps to 0.5.0.

Tests added:
- prompt coverage of the three issue behaviors in internal/opencode
- record-line helpers in internal/testutil
- updated forwarding assertions in internal/review

Ralph item 4 (Code Review) completed
