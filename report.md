Accept a --variant option and forward it to opencode

The review command forwards --variant to opencode run. Without a --model, the variant defaults to high. With a --model, no variant is passed, and an explicit --variant always applies.

Tests added:
- parsing, defaults, and the empty variant in cmd/zpecs
- forwarding in internal/review
- default, explicit, combined, and omitted invocation in internal/opencode
- record-line helpers in internal/testutil

Ralph item 3 (Variant Option) completed
