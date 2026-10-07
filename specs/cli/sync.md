# CLI Sync Specification

## Purpose
Write rendered definitions into a target repository and keep it in sync with the source. [Feature flags](features.md) filter optional content before a run writes it. A docs run also indexes the standards it writes in the repository's `AGENTS.md`.

## Requirements

### Requirement: Repository Root
The system SHALL run inside a git repository and write to its root.

#### Scenario: Nested invocation
- GIVEN the command runs in a subdirectory of a git repository
- WHEN the system writes definitions
- THEN it writes them at the repository root

#### Scenario: Outside a repository
- GIVEN the command runs outside a git repository
- WHEN the system runs
- THEN it errors and writes nothing

### Requirement: Target Paths
The system SHALL write rendered definitions to the target's directories, keyed by each definition's source name.

#### Scenario: Definition names the path
- GIVEN a source file at `agents/prose-editor.md`
- WHEN the system writes it to either target
- THEN it writes it under the `prose-editor` name regardless of the rendered fields

#### Scenario: Claude target
- GIVEN `--target claude`
- WHEN the system writes a skill and an agent
- THEN it writes them to `.claude/skills/<name>/SKILL.md` and `.claude/agents/<name>.md`

#### Scenario: OpenCode target
- GIVEN `--target opencode`
- WHEN the system writes a skill and an agent
- THEN it writes them to `.opencode/skills/<name>/SKILL.md` and `.opencode/agents/<name>.md`

#### Scenario: Docs target
- GIVEN `update docs`
- WHEN the system writes a doc
- THEN it writes it to `docs/zpecs/<name>.md`

### Requirement: Missing Directories
The system SHALL create directories it needs that do not exist.

### Requirement: Manifest
The system SHALL record the paths it wrote under a target in that target's `.zpecs` manifest. The docs manifest also records the features the run enabled, as [Feature Flags](features.md) describes.

#### Scenario: Ownership recorded
- GIVEN a run writes a definition to a target
- WHEN the system saves the manifest
- THEN the manifest lists the written path

#### Scenario: Empty manifest removed
- GIVEN a target with no owned paths and no recorded features
- WHEN the system saves the manifest
- THEN no manifest remains

### Requirement: Owned Files
The system SHALL replace only the files it wrote before and leave other files alone.

#### Scenario: Foreign file survives
- GIVEN a file in the target directory the system did not write
- WHEN the system runs
- THEN that file is unchanged

### Requirement: Stale Definitions
The system SHALL stop listing definitions the source no longer provides.

#### Scenario: Removed skill
- GIVEN a skill the source previously listed
- WHEN the source no longer lists it
- AND the system runs
- THEN the rendered skill is removed from the target

### Requirement: Agents Section
The system SHALL keep the repository's `AGENTS.md` in sync with a `## Zpecs` section when a run includes docs. It SHALL render the source's `docs/agents-section.md` template, replace an existing section in place, append the section when the document has none, and leave the rest of the document alone. A source without the template leaves `AGENTS.md` alone.

#### Scenario: Section added
- GIVEN a repository with no `## Zpecs` section
- WHEN the system updates docs
- THEN it appends the rendered section to `AGENTS.md`

#### Scenario: Section replaced
- GIVEN a repository whose `AGENTS.md` has a `## Zpecs` section
- WHEN the system updates docs
- THEN it replaces only that section

#### Scenario: Other content kept
- GIVEN an `AGENTS.md` with content beyond the `## Zpecs` section
- WHEN the system updates docs
- THEN the rest of the document is unchanged

#### Scenario: Optional link gated
- GIVEN the section template links to an orchestration document
- WHEN the system updates docs without `--orchestration` and no recorded feature
- THEN the link does not appear in the section

#### Scenario: Skills-only run
- GIVEN an `update skills` run
- WHEN the system runs
- THEN it leaves `AGENTS.md` alone

#### Scenario: No template
- GIVEN a source without `docs/agents-section.md`
- WHEN the system updates docs
- THEN it leaves `AGENTS.md` alone

### Requirement: Command Scope
The system SHALL render what each command names.

#### Scenario: Full update
- GIVEN `update`
- WHEN the system runs
- THEN it renders skills for the target
- AND it syncs docs
- AND it does not render agents

#### Scenario: Full update with agents
- GIVEN `update --agents`
- WHEN the system runs
- THEN it renders skills and agents for the target
- AND it syncs docs

#### Scenario: Skills only
- GIVEN `update skills`
- WHEN the system runs
- THEN it renders skills and not agents

#### Scenario: Agents only
- GIVEN `update agents`
- WHEN the system runs
- THEN it renders agents and not skills, even without `--agents`

#### Scenario: Docs only
- GIVEN `update docs`
- WHEN the system runs
- THEN it syncs docs and not skills or agents
