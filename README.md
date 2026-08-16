# zpecs

Document formats and standards for spec-driven development with AI coding agents, plus the skills that author them.

Projects install it rather than depend on it, so a project gets a copy it can read, edit, and diff.

## Install

Run the [CLI](docs/cli/README.md) inside the target repository to render the skill and agent definitions into its `.claude` or `.opencode` directory:

```bash
zpecs update --target claude
zpecs update --target opencode
```

Copy the standards docs into the target's `docs/specs/`, then point the target's `AGENTS.md` at them:

```markdown
Before writing any code, read [docs/specs/architecture.md](docs/specs/architecture.md).
```

## What's Inside

**Formats** — [spec](docs/specs/specs.md), [architecture](docs/specs/architecture-outline.md), [project](docs/specs/project.md). A spec defines what the system must do. The architecture records where the code lives. A project is a plain list of requirements, one per iteration.

**Standards** — [orchestration](docs/specs/orchestration.md), [architecture guidelines](docs/specs/architecture.md), [writing requirements](docs/specs/requirements.md), [agent prompts](docs/specs/prompts.md), [glossary](docs/specs/glossary.md).

**Skills** — six skills an agent invokes by name:

| Skill | Writes |
|---|---|
| `write-spec` | `specs/<path>.md` |
| `write-architecture` | `specs/architecture.yaml` |
| `write-project` | `projects/<slug>.yaml` |
| `prose-editor` | fixes for violations in docs, comments, or agent prose |
| `write-skill` | `skills/<name>/SKILL.md` |
| `write-agent` | `agents/<name>.md` |

## Runners

Nothing here executes a project. Running it is a runner's job. [ralph](https://github.com/zon/ralph) is one — it picks one requirement per iteration, drives a coding agent through it, and records completion in the branch's commit log.

The formats stay runner-neutral.

## Recipes

```bash
just check   # verify every relative markdown link resolves
```

## License

GPL-3.0
