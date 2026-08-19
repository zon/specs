# Testing

## Overview

Tests split into two layers: unit and integration.
- **Unit** — individual functions and packages in isolation
- **Integration** — full request or execution paths end-to-end within the process

Examples below are Go.

## Conventions

### Assertions

Use one assertion library across the whole suite for every assertion: `github.com/stretchr/testify` (`assert` and `require`).

### Test structure

Use table-driven tests with named subtests (`t.Run()`). Use a managed temporary directory for any filesystem interaction (`t.TempDir()`) so cleanup is automatic.

### Isolation

Unit tests may run code with real side effects.

Invoke a real dependency in a test only when it is safe, local, and cheap. Version control and local CLI tools qualify, but only against an isolated temporary directory, **never** against the real repository. Create the directory with the temp-directory helper and initialise a fresh repo inside it before the test runs.

Never invoke a real dependency in a test when it:

- mutates state outside the test (remote APIs, clusters, deployments, notifications)
- costs money per call
- requires credentials the suite does not own
- reaches the network for anything the test's correctness depends on

### Dependency unit tests

Unit tests for real external dependencies must be small, focused, and cheap. They exist only to verify the lowest-level interface of the real implementation, not to exercise full workflows.

- Test only the minimal surface needed to confirm the real dependency works (e.g. a single command round-trip).
- Use the shortest possible inputs. For a metered dependency such as an AI CLI, hard-code the cheapest available model and a trivial prompt like `"say hi"`.
- These tests must not accumulate: one or two cases per real dependency are enough.

### Module boundaries

Only implementation modules may hold real dependency implementations. See [Architecture Guidelines](architecture.md).

| Module type | May contain |
|---|---|
| Orchestration | interfaces, orchestration functions, tests |
| Implementation | real dependency implementations, tests |

### CLI command validation

Test the commands defined in specs against the expected format. Use the command's own help output to confirm the structure matches the specification:

```go
func TestCommand(t *testing.T) {
    cmd := exec.Command("go", "run", "./cmd/myapp", "subcommand", "name", "--help")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)
    assert.Contains(t, string(output), "Expected help text from spec")
}
```

This catches structural issues such as:
- Incorrect command names (e.g. duplicated subcommand names)
- Missing subcommands
- Mismatched help text or flags

Run these tests as part of the standard test suite to ensure the CLI structure stays aligned with specifications.
