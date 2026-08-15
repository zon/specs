# CLI

The `zpecs` CLI renders this repository's skill and agent definitions into a project's `.claude` or `.opencode` directory.

It is a Go program that uses [kong](https://github.com/alecthomas/kong) to parse arguments.

## Commands

Three commands:

- `update` — render skill and agent definitions
- `update skills` — render skill definitions only
- `update agents` — render agent definitions only

All three take the same flags:

| Flag | Meaning |
|---|---|---|
| `--target` | Render for `opencode` (default) or `claude` |
| `--source` | Path to a local directory to read from; omitted, reads from GitHub |

## Definitions

Definitions are the neutral source of truth. They live in the GitHub repository:

```
skills/<name>.md
agents/<name>.md
```

See the [Skill Format](formats/skill.md) and [Agent Format](formats/agent.md).

## Source

A local source must have the same layout as the repository.

## Target

The commands must run inside a git repository, locate the repo root, and write there:

- `claude` — `.claude/skills/<name>/SKILL.md` and `.claude/agents/<name>.md`
- `opencode` — `.opencode/skills/<name>/SKILL.md` and `.opencode/agents/<name>.md`

It creates missing directories and replaces the definitions it wrote last time, so definitions removed from the source stop appearing and other files stay untouched.

## Rendering

Both targets use the same frontmatter fields, but not identically:

- claude agents use `name`
- opencode agents use `mode: subagent` and drop `name`

The renderer maps each definition to the target's fields.
