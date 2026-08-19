# CLI Source Specification

## Purpose
Read skill, agent, and doc definitions from a local directory or from GitHub.

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

### Requirement: Docs
The system SHALL read doc definitions from `docs/zpecs/<name>.md`.

#### Scenario: Doc placement
- GIVEN a source that places a doc at `docs/zpecs/prose.md`
- AND a source that places a doc outside `docs/zpecs/<name>.md`
- WHEN the system reads definitions
- THEN it finds the doc at `docs/zpecs/prose.md`
- AND it does not find the doc outside that path
