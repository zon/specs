# CLI Rendering Specification

## Purpose
Map skill and agent definitions to each target runner's format.

## Requirements

### Requirement: Shared Frontmatter
The system SHALL read the same frontmatter fields for every target.

#### Scenario: Same fields for both targets
- GIVEN an agent definition with `name`, `description`, and `tools` fields
- WHEN the system renders it for both targets
- THEN it reads the same fields for each

### Requirement: Claude Agent Naming
The system SHALL use the `name` field for a claude agent.

#### Scenario: Claude agent
- GIVEN an agent with `name: prose-editor`
- WHEN the system renders it for claude
- THEN the rendered agent keeps `name: prose-editor`

### Requirement: OpenCode Agent Naming
The system SHALL use `mode: subagent` for an opencode agent and drop its `name`.

#### Scenario: OpenCode agent
- GIVEN an agent with `name: prose-editor`
- WHEN the system renders it for opencode
- THEN the rendered agent has `mode: subagent`
- AND it has no `name` field

### Requirement: Claude Tools
The system SHALL map the `tools` list to a `tools` list for a claude agent.

#### Scenario: Claude tools
- GIVEN an agent with `tools: [read, edit]`
- WHEN the system renders it for claude
- THEN the rendered agent lists `read` and `edit` in its `tools`

### Requirement: OpenCode Tools
The system SHALL map the `tools` list to deny rules for every other tool for an opencode agent.

#### Scenario: OpenCode tools
- GIVEN an agent with `tools: [read, edit]`
- WHEN the system renders it for opencode
- THEN the rendered agent denies every tool except `read` and `edit`

### Requirement: Agent Body
The system SHALL make the agent's body the rendered agent's system prompt.
