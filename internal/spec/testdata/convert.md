# Convert

## Purpose
Turn a spec markdown file into a JSON document.

### Requirement: Command Form
The system SHALL accept a path to a spec file as its only argument.

#### Scenario: Path argument
- GIVEN a path to a spec file
- WHEN the command runs
- THEN it reads the spec at that path

### Requirement: Output
The system SHALL print one JSON object to stdout.
