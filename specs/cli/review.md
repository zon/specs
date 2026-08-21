# CLI Review Specification

## Purpose
Run an opencode review of the repository against the code, architecture, or prose standards.

## Requirements

### Requirement: Command Form
The system SHALL accept a scope of `code`, `architecture`, or `prose` for the review command.

#### Scenario: Code scope
- GIVEN `zpecs review code`
- WHEN the command runs
- THEN it reviews the repository against the code guidelines

#### Scenario: Architecture scope
- GIVEN `zpecs review architecture`
- WHEN the command runs
- THEN it reviews the repository against the architecture guidelines

#### Scenario: Prose scope
- GIVEN `zpecs review prose`
- WHEN the command runs
- THEN it reviews the repository against the prose guidelines

#### Scenario: Unknown scope
- GIVEN a scope the command does not name
- WHEN the command runs
- THEN it errors and reviews nothing

### Requirement: Repository Scope
The system SHALL review the repository the command runs in.

#### Scenario: Nested invocation
- GIVEN the command runs in a subdirectory of a git repository
- WHEN the command runs
- THEN it reviews the repository at its root

### Requirement: Model Option
The system SHALL accept a `--model` option and forward its value to opencode. When the user gives no value, the model is `deepseek/deepseek-v4-flash`.

#### Scenario: Default model
- GIVEN no `--model`
- WHEN the command runs
- THEN opencode runs with the model `deepseek/deepseek-v4-flash`

#### Scenario: Custom model
- GIVEN `--model anthropic/claude-sonnet-4-5`
- WHEN the command runs
- THEN opencode runs with that model

### Requirement: Variant Option
The system SHALL accept a `--variant` option and forward its value to opencode. When the user gives no `--model`, the variant defaults to `high`; when the user gives a `--model`, the variant carries no value.

#### Scenario: Default variant
- GIVEN no `--model` and no `--variant`
- WHEN the command runs
- THEN opencode runs with the variant `high`

#### Scenario: Model picks the variant
- GIVEN `--model anthropic/claude-sonnet-4-5` and no `--variant`
- WHEN the command runs
- THEN opencode runs with no variant

#### Scenario: Explicit variant
- GIVEN `--variant minimal`
- WHEN the command runs
- THEN opencode runs with the variant `minimal`

### Requirement: Code Review
The system SHALL prompt opencode to write each code issue it finds to a `refactor-<slug>.yaml` project rather than editing the code.

#### Scenario: Issue becomes a project
- GIVEN `zpecs review code`
- AND opencode finds a code issue
- WHEN the command runs
- THEN a `refactor-<slug>.yaml` project appears in `projects/`

#### Scenario: Existing project is updated
- GIVEN `zpecs review code`
- AND a matching `refactor-<slug>.yaml` project already exists
- AND opencode finds another issue in the same refactor
- WHEN the command runs
- THEN the existing project gains the issue

#### Scenario: No issues
- GIVEN `zpecs review code`
- AND opencode finds no code issues
- WHEN the command runs
- THEN no refactor project is written

### Requirement: Architecture Review
The system SHALL prompt opencode to write each architecture issue it finds to a `refactor-<slug>.yaml` project rather than editing the code.

#### Scenario: Issue becomes a project
- GIVEN `zpecs review architecture`
- AND opencode finds an architecture issue
- WHEN the command runs
- THEN a `refactor-<slug>.yaml` project appears in `projects/`

### Requirement: Prose Review
The system SHALL prompt opencode to fix each prose issue it finds immediately.

#### Scenario: Issue is fixed
- GIVEN `zpecs review prose`
- AND opencode finds a prose issue
- WHEN the command runs
- THEN the offending text is edited in place
