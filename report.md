Remove internal/target from the architecture

The target names claude, opencode, and docs now live in internal/source,
so specs/architecture.yaml no longer lists internal/target as a
component.

Tests: none added. The source suite already pins the target name
constants, and the remaining architecture entries match the code.

Ralph item 1 completed
