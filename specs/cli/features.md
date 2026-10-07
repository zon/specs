# CLI Features Specification

## Purpose
Control optional documents, skills, and agents with feature flags, so a run copies only the content a project uses. A docs run records the flags it enables, so the project keeps that content.

## Requirements

### Requirement: Feature Flags
The system SHALL accept one flag per optional feature area. The optional features are orchestration, named `--orchestration`; process, named `--process`; and design, named `--design`.

#### Scenario: Flag omitted
- GIVEN an update run without `--orchestration` and no recorded orchestration feature
- WHEN the system selects optional content
- THEN the orchestration feature is off

#### Scenario: Flag passed
- GIVEN an update run with `--orchestration`
- WHEN the system selects optional content
- THEN the orchestration feature is on

#### Scenario: Process flag passed
- GIVEN an update run with `--process`
- WHEN the system selects optional content
- THEN the process feature is on

#### Scenario: Design flag passed
- GIVEN an update run with `--design`
- WHEN the system selects optional content
- THEN the design feature is on

### Requirement: Feature Defaults
The system MUST treat every feature as off unless the user passes its flag or the docs manifest records it.

#### Scenario: Fresh run
- GIVEN a run with no feature flags and no recorded features
- WHEN the system selects optional content
- THEN it includes no optional content

### Requirement: Recorded Features
The system SHALL record the features a run enables in the docs manifest and include them in later runs that pass no flag.

#### Scenario: Feature recorded
- GIVEN a run with `--orchestration` that writes docs
- WHEN the system writes the docs manifest
- THEN it records the orchestration feature

#### Scenario: Recorded feature persists
- GIVEN the docs manifest records the orchestration feature
- WHEN a later run passes no feature flag
- THEN the system selects optional orchestration content

#### Scenario: Recorded feature reaches other kinds
- GIVEN the docs manifest records the orchestration feature
- WHEN a later run renders a skill and passes no feature flag
- THEN the rendered skill includes its orchestration content

### Requirement: Optional Definitions
The system SHALL omit a definition whose whole content belongs to a disabled feature.

#### Scenario: Optional doc omitted
- GIVEN `--orchestration` is absent and no recorded orchestration feature
- WHEN the system updates docs
- THEN it does not write the orchestration document

#### Scenario: Optional doc restored
- GIVEN the orchestration document was omitted
- WHEN the user runs with `--orchestration`
- THEN the system writes the orchestration document

#### Scenario: Recorded definition is kept
- GIVEN a run with `--orchestration` wrote the orchestration document and recorded the feature
- WHEN the user runs without the flag
- THEN the system writes the orchestration document again

### Requirement: Optional Content
The system SHALL omit content inside a definition that belongs to a disabled feature, and keep the rest of the definition.

#### Scenario: Disabled section
- GIVEN a document with a section marked for the orchestration feature
- WHEN the system renders it without `--orchestration` and no recorded orchestration feature
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
- WHEN the system renders it without `--orchestration` and no recorded orchestration feature
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
- GIVEN the orchestration, process, and design features
- WHEN another feature flag is added
- THEN each flag still selects its own content and defaults to off when unrecorded
