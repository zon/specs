# CLI Sync Specification

## Purpose
Write rendered definitions into a target repository and keep it in sync with the source.

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

### Requirement: Missing Directories
The system SHALL create directories it needs that do not exist.

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

### Requirement: Command Scope
The system SHALL render what each command names.

#### Scenario: Full update
- GIVEN `update`
- WHEN the system runs
- THEN it renders skills and agents

#### Scenario: Skills only
- GIVEN `update skills`
- WHEN the system runs
- THEN it renders skills and not agents

#### Scenario: Agents only
- GIVEN `update agents`
- WHEN the system runs
- THEN it renders agents and not skills
