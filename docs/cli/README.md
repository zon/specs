# CLI

The `zpecs` CLI renders this repository's definitions into a project's `.claude` or `.opencode` directory.

It is a Go program that parses arguments with [kong](https://github.com/alecthomas/kong).

## Commands

One command takes an optional scope:

- `update` — render skill and agent definitions
- `update skills` — render skill definitions only
- `update agents` — render agent definitions only

All three take the same flags:

| Flag | Meaning |
|---|---|---|
| `--target` | Render for `opencode` (default) or `claude` |
| `--source` | Path to a local directory. Omit it to read from GitHub |

## Definitions

Definitions are the neutral source of truth. They live in the GitHub repository:

```
skills/<name>/SKILL.md
agents/<name>.md
```

Skills follow the [Agent Skills spec](https://agentskills.io/specification). Agents follow the [Agent Format](agent-format.md).

## Source

A local source must have the same layout as the repository.

## Target

The commands must run inside a git repository, locate the repo root, and write there:

- `claude` — `.claude/skills/<name>/SKILL.md` and `.claude/agents/<name>.md`
- `opencode` — `.opencode/skills/<name>/SKILL.md` and `.opencode/agents/<name>.md`

It creates missing directories and replaces the files it wrote, leaving other files alone. Stale definitions stop appearing.

## Rendering

Both targets read the same frontmatter fields but map them differently:

- claude agents use `name`
- opencode agents use `mode: subagent` and drop `name`
- the `tools` list maps to `tools` in claude agents, and to deny rules for every tool not in the list in opencode agents
