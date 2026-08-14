# Specs

Document formats and standards for spec-driven development with AI coding agents, plus the skills that author them.

This repository holds the opinions: what a spec looks like, how modules are categorized, how a project splits work into iterations, what belongs in a test. It is installed into a project rather than depended on, so a project gets a copy it can read, edit, and diff.

## Install

```bash
just install ../myproject
```

That copies:

```
docs/specs/*.md              ->  <dir>/docs/specs/*.md
.claude/skills/spec-*/       ->  <dir>/.claude/skills/spec-*/
```

The docs land at the same path they occupy here, so every link inside them resolves identically in both repositories — nothing is rewritten on the way in. Installing again replaces both trees and prunes any `spec-` skill this repository no longer publishes. Skills without the `spec-` prefix are left alone.

Then point the target's `AGENTS.md` at the standards:

```markdown
Before writing any code, read [docs/specs/code.md](docs/specs/code.md).
Before writing any tests, read [docs/specs/testing.md](docs/specs/testing.md).
```

## What's Inside

**Formats** — [spec](docs/specs/spec.md), [orchestration](docs/specs/orchestration.md), [architecture](docs/specs/architecture.md), [project](docs/specs/project.md). The first three build on each other: what the system must do, the shape of the logic that does it, and where that logic lives. A project is a plain list of requirements, one per iteration.

**Standards** — [writing code](docs/specs/code.md), [testing](docs/specs/testing.md), [writing requirements](docs/specs/requirements.md), [agent prompts](docs/specs/prompts.md), [glossary](docs/specs/glossary.md).

**Skills** — six skills an agent invokes by name:

| Skill | Writes |
|---|---|
| `spec-write-spec` | `specs/features/<component>/<feature>/spec.md` |
| `spec-write-architecture` | `specs/architecture.yaml` |
| `spec-write-orchestration` | `specs/features/<component>/<feature>/orchestration.md` |
| `spec-write-project` | `projects/<slug>.yaml` |
| `spec-write-skill` | `.claude/skills/<name>/SKILL.md` |
| `spec-review-module` | a findings report, and a project to fix the gaps |

**Prompts** — [outline](docs/specs/outline.md) generates a whole `/specs` tree from an existing codebase.

## Runners

Nothing here executes a project. A project is a file listing requirements; running it is a runner's job. [ralph](https://github.com/zon/ralph) is one — it picks one requirement per iteration, drives a coding agent through it, and records completion in the branch's commit log.

The formats stay runner-neutral: a project is a plain top-level list, and how it is executed is the runner's business.

## Recipes

```bash
just install <dir>   # copy docs and skills into another repo
just list            # show what install would copy
just check           # verify every relative markdown link resolves
```

## License

GPL-3.0
