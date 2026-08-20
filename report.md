Drop the targetName wrapper type

Removed the targetName type and its targetClaude and targetOpencode
constants from internal/cli. The update command's target flag is now a
plain string that kong validates against the enum, and the grammar's
enum and default values come from the target package constants.
Updated the CLI parsing tests to expect the target package constants.
Removed the completed plan from refactoring.md. Bumped VERSION to
0.1.8.

Ralph item 2 completed
