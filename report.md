Delegate the convert command's read-and-write sequence to a helper

cmd/zpecs now only parses the command line and calls the named helper,
matching the update command's shape. internal/spec owns the sequence
in Convert, which reads a spec file and prints it as JSON. The other
spec entry points became private.

Tests: internal/spec covers Convert printing a file as JSON, and
erroring on missing and malformed files without output. The existing
cmd/zpecs convert tests pass unchanged.

Ralph item 0 completed
