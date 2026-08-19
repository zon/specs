# CLI Convert Specification

## Purpose
Turn a spec markdown file into a JSON document that keeps its structure.

## Requirements

### Requirement: Command Form
The system SHALL accept a path to a spec file as its only argument.

#### Scenario: Path argument
- GIVEN the path `specs/cli/sync.md`
- WHEN the command runs
- THEN it reads the spec at that path

### Requirement: Output
The system SHALL print one JSON object to stdout. The object has a `title` field for the top-level heading, a `purpose` field for the purpose text, and a `requirements` field listing each requirement.

#### Scenario: Whole spec
- GIVEN a spec file with a title, a purpose, and one requirement
- WHEN the command runs
- THEN it prints one JSON object
- AND the object carries the title, purpose, and requirements

### Requirement: Requirements
The system SHALL capture each `### Requirement:` heading as a requirement. Each requirement has a `name` field from the heading and a `body` field from its prose.

#### Scenario: Requirement with scenarios
- GIVEN a requirement that has scenarios
- WHEN the command runs
- THEN the requirement appears in the JSON with its scenarios

### Requirement: Scenarios
The system SHALL capture each `#### Scenario:` heading as a scenario. Each scenario has a `name` field from the heading and a `steps` list of the step lines as text.

#### Scenario: Steps in order
- GIVEN a scenario with GIVEN, WHEN, THEN, and AND lines
- WHEN the command runs
- THEN each line appears in order as a step

### Requirement: Missing File
The system SHALL error when the argument does not point at a file.

#### Scenario: Nonexistent path
- GIVEN a path that does not exist
- WHEN the command runs
- THEN it errors and prints no JSON

### Requirement: Malformed File
The system SHALL error when the file does not follow the spec format.

#### Scenario: Missing title
- GIVEN a file with no top-level heading
- WHEN the command runs
- THEN it errors

#### Scenario: Missing requirement name
- GIVEN a requirement heading with no name
- WHEN the command runs
- THEN it errors
