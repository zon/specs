Initialize the repo as a Go module that builds a CLI binary

The module and the CLI entry point are new. AGENTS.md and the justfile record the build and test commands, and the built binary is ignored.

Tests added: a package test that builds and runs the binary, asserting the executable exists and exits cleanly.

Ralph item 16 completed
