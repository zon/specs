Write every definition at the repository root

The update command resolves the git repository root from the working
directory and writes every definition there. A run inside a subdirectory
still lands at the root. Outside a repository, the command errors before
it reads or writes anything.

A new internal/repo module finds the root by walking up until a .git
entry appears. The targetdir module writes under the root instead of the
working directory.

Tests cover finding the root from a subdirectory, erroring outside a
repository without writing anything, and nested runs writing at the
root, in-process and as a built binary.

Ralph item 10 completed
