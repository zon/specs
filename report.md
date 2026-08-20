Move target names into internal/source

The target names claude, opencode, and docs are now defined once in
internal/source, and the internal/target package is removed. Its callers
read the constants from internal/source.

Tests: source gains a test pinning the target name constants. The
existing update, targetdir, render, and report suites cover the moved
constants through the behavior they drive.

Ralph item 0 completed
