# zpecs

Document formats and standards for spec-driven development with AI coding agents, plus the skills that author them.

This repository holds the opinions: what a spec looks like, how modules are categorized, how a project splits work into iterations, what belongs in a test. Projects install it rather than depend on it, so a project gets a copy it can read, edit, and diff.

## Install

Run the [CLI](docs/cli/README.md) inside the target repository to render the skill and agent definitions into its `.claude` or `.opencode` directory:

```bash
zpecs update --target claude
zpecs update --target opencode
```

Copy the standards docs into the target's `docs/specs/`, then point the target's `AGENTS.md` at them:

```markdown
Before writing any code, read [docs/specs/code.md](docs/specs/code.md).
Before writing any tests, read [docs/specs/testing.md](docs/specs/testing.md).
```

## What's Inside

**Formats** — [spec](docs/specs/spec.md), [orchestration](docs/specs/orchestration.md), [architecture](docs/specs/architecture.md), [project](docs/specs/project.md). The first three build on each other: what the system must do, the shape of the logic that does it, and where that logic lives. A project is a plain list of requirements, one per iteration.

**Standards** — [writing code](docs/specs/code.md), [testing](docs/specs/testing.md), [writing requirements](docs/specs/requirements.md), [agent prompts](docs/specs/prompts.md), [glossary](docs/specs/glossary.md).

**Skills** — eight skills an agent invokes by name:

| Skill | Writes |
|---|---|
| `write-spec` | `specs/features/<component>/<feature>/spec.md` |
| `write-architecture` | `specs/architecture.yaml` |
| `write-orchestration` | `specs/features/<component>/<feature>/orchestration.md` |
| `write-project` | `projects/<slug>.yaml` |
| `review-module` | a findings report, and a project to fix the gaps |
| `prose-editor` | fixes for violations in docs, comments, or agent prose |
| `write-skill` | `skills/<name>/SKILL.md` |
| `write-agent` | `agents/<name>.md` |

**Prompts** — [outline](docs/specs/outline.md) generates a whole `/specs` tree from an existing codebase.

## Runners

Nothing here executes a project; running it is a runner's job. [ralph](https://github.com/zon/ralph) is one — it picks one requirement per iteration, drives a coding agent through it, and records completion in the branch's commit log.

The formats stay runner-neutral.

## Recipes

```bash
just check   # verify every relative markdown link resolves
```

## License

GPL-3.0
