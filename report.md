Accept a --model option and forward it to opencode

The review command forwards --model to opencode run. Without it, opencode runs with deepseek/deepseek-v4-flash.

Tests added:
- parsing and defaults in cmd/zpecs
- forwarding in internal/review
- default and custom invocation in internal/opencode

Ralph item 2 (Model Option) completed
