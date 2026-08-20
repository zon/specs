# zpecs

Common standards, agents, and skills for spec-driven development with AI coding agents. Apply them to any repository with the `zpecs` CLI.

Repositories install these rather than depend on them, so a repository gets a copy it can read, edit, and diff.

## Standards

The formats and guidelines that shape how a repository writes and organizes its specs, architecture, and requirements. They install to the target's `docs/zpecs/`. See [Specs](docs/zpecs/README.md) for the full list of formats and guidelines.

## Agents

Coding agents that plan, write, review, test, and polish code. `zpecs update agents` renders them into a repository's `.claude/agents/` or `.opencode/agents/`.

| Agent | Does |
|---|---|
| `code-architect` | Plans implementation through writer and reviewer subagents |
| `architecture-reviewer` | Reviews the repo's architecture against the standards |
| `code-writer` | Writes the code and tests for one step |
| `component-reviewer` | Reviews a single component against the repo's standards |
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

