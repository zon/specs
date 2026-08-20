Bump the app version to 0.1.15

The convert command refactor is a patch-level change, so this is a
patch bump. VERSION stays the single source of truth: the justfile
passes it to the binary at build time, and the version-format test
covers the file.

Tests: TestVersionFileHoldsSemver already asserts the VERSION file
holds a valid semver. Both the test suite and just check pass.

Ralph item 1 completed
