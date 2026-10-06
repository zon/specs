# CLI Features Specification

## Purpose
Control optional documents, skills, and agents with feature flags, so a run copies only the content a project uses.

## Requirements

### Requirement: Feature Flags
The system SHALL accept one flag per optional feature area. The first optional feature is orchestration, named `--orchestration`.

#### Scenario: Flag omitted
- GIVEN an update run without `--orchestration`
- WHEN the system selects optional content
- THEN the orchestration feature is off

#### Scenario: Flag passed
- GIVEN an update run with `--orchestration`
- WHEN the system selects optional content
- THEN the orchestration feature is on

### Requirement: Feature Defaults
The system MUST treat every feature as off unless the user passes its flag.

#### Scenario: Fresh run
- GIVEN a run with no feature flags
- WHEN the system selects optional content
- THEN it includes no optional content

### Requirement: Optional Definitions
The system SHALL omit a definition whose whole content belongs to a disabled feature.

#### Scenario: Optional doc omitted
- GIVEN `--orchestration` is absent
- WHEN the system updates docs
- THEN it does not write the orchestration document

#### Scenario: Optional doc restored
- GIVEN the orchestration document was omitted
- WHEN the user runs with `--orchestration`
- THEN the system writes the orchestration document

#### Scenario: Omitted definition is removed
- GIVEN a run with `--orchestration` wrote the orchestration document
- WHEN the user runs without the flag
- THEN the system removes the document as stale

### Requirement: Optional Content
The system SHALL omit content inside a definition that belongs to a disabled feature, and keep the rest of the definition.

#### Scenario: Disabled section
- GIVEN a document with a section marked for the orchestration feature
- WHEN the system renders it without `--orchestration`
- THEN the section does not appear
- AND the rest of the document is unchanged

#### Scenario: Enabled section
- GIVEN the same document
- WHEN the system renders it with `--orchestration`
- THEN the section appears

### Requirement: Optional Links
The system SHALL omit a link to a disabled feature's content, so no written definition links to content the run did not write.

#### Scenario: Dead link removed
- GIVEN a document that links to the orchestration document
- WHEN the system renders it without `--orchestration`
- THEN the link does not appear in the written document

#### Scenario: Link returns with its feature
- GIVEN the same document
- WHEN the system renders it with `--orchestration`
- THEN the link appears again

### Requirement: Feature Isolation
Feature flags SHALL only select optional content. They SHALL NOT change a run's scope, target, or file ownership.

#### Scenario: Scope unchanged
- GIVEN `update skills --orchestration`
- WHEN the system runs
- THEN it writes skills only

### Requirement: Extensible Features
The system SHALL add a new feature flag without changing how it selects other features.

#### Scenario: Adding a feature
- GIVEN the orchestration feature
- WHEN a second feature flag is added
- THEN each flag still selects its own content and defaults to off
