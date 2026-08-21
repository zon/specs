Add the zpecs review command

The CLI now accepts zpecs review code, architecture, or prose. Each runs an opencode review of the repository against that scope's guidelines. An unknown scope errors and reviews nothing.

Tests added:
- scope parsing and command recognition tests in cmd/zpecs
- orchestration tests in internal/review
- opencode invocation tests in internal/opencode
- a fake opencode test fixture in internal/testutil

Ralph item 0 (Command Form) completed
