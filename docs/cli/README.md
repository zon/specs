# CLI

The `zpecs` CLI renders this repository's definitions into a project's `.claude` or `.opencode` directory. It syncs the standards docs into the project's `docs/zpecs/`.

It is a Go program that parses arguments with [kong](https://github.com/alecthomas/kong).

## Commands

One command takes an optional scope:

- `update` — render skill and agent definitions
- `update skills` — render skill definitions only
- `update agents` — render agent definitions only
- `update docs` — sync the standards docs from the source into the target's `docs/zpecs/`

All four take the same flags, but `update docs` ignores `--target`:

| Flag | Meaning |
|---|---|
| `--target` | Render for `opencode` (default) or `claude` |
| `--source` | Path to a local directory, or a repository URL to clone. Omit it to read from GitHub |
| `--version` | Print the version and quit |

`ZPECS_SOURCE` sets the source when `--source` is absent.

## Definitions

Definitions are the neutral source of truth. They live in the GitHub repository:

```
skills/<name>/SKILL.md
agents/<name>.md
docs/zpecs/<name>.md
```

Skills follow the [Agent Skills spec](https://agentskills.io/specification). Agents follow the [Agent Format](agent-format.md). Docs are copied verbatim.

## Source

A local source must have the same layout as the repository.

## Target

The commands must run inside a git repository, locate the repo root, and write there:

- `claude` — `.claude/skills/<name>/SKILL.md` and `.claude/agents/<name>.md`
- `opencode` — `.opencode/skills/<name>/SKILL.md` and `.opencode/agents/<name>.md`
- `docs` — `docs/zpecs/<name>.md`. Docs are target-independent

When a repo's `docs/zpecs/` files came from a copy, there is no `.zpecs` manifest. The first `update docs` run does not touch them. It uses the same owned-file semantics as the other targets.

It creates missing directories and replaces the files it wrote, leaving other files alone. Stale definitions stop appearing.

## Rendering

Both targets read the same frontmatter fields but map them differently:

- claude agents use `name` and drop `mode`
- opencode agents use `mode` and drop `name`, defaulting to `subagent` when the definition omits it
- the `tools` list maps to `tools` in claude agents. In opencode agents it maps to deny rules for every tool not in the list
