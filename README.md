# zpecs

Common standards, agents, and skills for spec-driven development with AI coding agents. Apply them to any repository with the `zpecs` CLI.

Repositories install these rather than depend on them, so a repository gets a copy it can read, edit, and diff.

## Standards

The formats and guidelines that shape how a repository writes and organizes its specs, architecture, and requirements. They install to the target's `docs/zpecs/`.

**Formats**

- [Spec](docs/zpecs/specs.md) — behavior contracts: structured requirements and scenarios
- [Architecture](docs/zpecs/architecture-outline.md) — the components of an application, in YAML
- [Project](docs/zpecs/project.md) — a list of requirements for a coding agent to work through

**Guidelines**

- [Orchestration](docs/zpecs/orchestration.md) — separates coordination logic from implementation detail
- [Architecture](docs/zpecs/architecture.md) — component placement, and what belongs in each component type
- [Writing Requirements](docs/zpecs/requirements.md) — what makes a good unit of work
- [Agent Prompts](docs/zpecs/prompts.md) — how to structure a single-task prompt
- [Dependencies](docs/zpecs/dependencies.md) — when to use one instead of writing your own
- [Prose Guidelines](docs/zpecs/prose.md) — how to write and organize prose
- [Glossary](docs/zpecs/glossary.md) — terms used throughout these documents

## Agents

Coding agents that plan, write, review, test, and polish code. `zpecs update agents` renders them into a repository's `.claude/agents/` or `.opencode/agents/`.

| Agent | Does |
|---|---|
| `code-architect` | Plans implementation through writer and reviewer subagents |
| `code-writer` | Writes the code and tests for one step |
| `code-reviewer` | Reviews code against the repo's standards |
| `code-tester` | Runs the test suite and reports results |
| `prose-editor` | Reviews prose against the guidelines and fixes violations |

## Skills

Skills an agent invokes by name to write new specs, architecture, projects, and definitions. `zpecs update skills` renders them into a repository's `.claude/skills/` or `.opencode/skills/`.

| Skill | Writes |
|---|---|
| `write-spec` | `specs/<path>.md` |
| `write-architecture` | `specs/architecture.yaml` |
| `write-project` | `projects/<slug>.yaml` |
| `write-skill` | `skills/<name>/SKILL.md` |
| `write-agent` | `agents/<name>.md` |

## Install and Use

Build and install the CLI from source:

```bash
go install ./cmd/zpecs
```

Run the [CLI](docs/cli/README.md) inside the target repository:

```bash
zpecs update --target claude
zpecs update --target opencode
```

To touch one area only, pass a scope:

```bash
zpecs update skills
zpecs update agents
zpecs update docs
```

Then point its `AGENTS.md` at the docs:

```markdown
Before writing any code, read [docs/zpecs/architecture.md](docs/zpecs/architecture.md).
```

