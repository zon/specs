Record the merged git-operations component in specs/architecture.yaml

The architecture document now lists internal/gitops, the component that
locates the repository root and clones a remote source, and no longer
lists internal/repo or internal/clone. All descriptions stay single
sentences per the architecture format guide.

Tests: no tests were added. The existing suite passes unchanged.

Ralph item 1 completed
