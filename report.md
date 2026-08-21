Prompt architecture reviews into refactor projects

The architecture review scope now prompts opencode to write each
architecture issue to a refactor-<slug>.yaml project in projects/ rather
than editing the code. VERSION bumps to 0.6.0.

Tests added:
- prompt coverage of the architecture issue behavior in internal/opencode
- updated forwarding assertions in internal/review
- updated record-line helpers in internal/testutil

Ralph item 5 (Architecture Review) completed
