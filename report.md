Drop the targetName wrapper type

Replace the targetName type and its constants in cmd/zpecs with the
target package constants and a plain string field. kong validates a
plain string against the enum, so the type only added conversions.
Update the tests that used the type. Remove the completed plan from
refactoring.md.

Ralph item 2 completed
