Merge git operations into one component
The update flow now locates the repository root and clones a remote
source through a single component, internal/gitops, that owns both git
operations. The repository-root lookup and the clone functions move
verbatim out of internal/repo and internal/clone, internal/update calls
the merged component, and the old packages are removed. Behavior is
unchanged.

Tests: the repo and clone test suites move into internal/gitops, keeping
every test. The existing update suite passes unchanged.

Ralph item 0 completed
