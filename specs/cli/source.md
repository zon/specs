# CLI Source Specification

## Purpose
Read skill and agent definitions from a local directory or from GitHub.

## Requirements

### Requirement: Default Source
The system SHALL read definitions from the GitHub repository when the user gives no `--source` flag.

#### Scenario: Local source
- GIVEN a `--source` flag pointing at a local directory
- WHEN the system reads definitions
- THEN it reads them from that directory

### Requirement: Local Source Layout
A local source SHALL mirror the repository layout.

#### Scenario: Misplaced definition
- GIVEN a local source that places a skill outside `skills/<name>/SKILL.md`
- WHEN the system reads definitions
- THEN it does not find that skill

### Requirement: Skills
The system SHALL read skill definitions from `skills/<name>/SKILL.md`.

### Requirement: Agents
The system SHALL read agent definitions from `agents/<name>.md`.
