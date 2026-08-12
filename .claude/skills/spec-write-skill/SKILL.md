---
name: spec-write-skill
description: Writes an agent skill to .claude/skills and makes it available to OpenCode and Claude Code. Use when the user wants to add a new skill or edit an existing one.
---

# Write Skill

Write a skill that a coding agent can invoke by name in any repository where it is installed.

## Steps

1. **Read the [OpenCode Skills docs](https://opencode.ai/docs/skills/)** describing the skill file format. The frontmatter `name` must match the directory name, and the `description` is what an agent matches against when deciding whether to invoke the skill — write it as the trigger, naming the situations that should call it.

2. **Read the spec** at `specs/features/<component>/<feature>/spec.md` describing what the skill should do, if one exists. If not, ask the user what the skill's task and output are.

3. **Read [docs/specs/prompts.md](docs/specs/prompts.md)** — a skill is a prompt that has to locate its own context, so the same structure applies: one task, numbered steps, explicit output.

4. **Write the skill** to `.claude/skills/<name>/SKILL.md`.

   - **Reference documentation rather than repeating it.** Link to the doc and let the agent read it when needed. A skill that restates a format goes stale the moment the format changes.
   - **Reference documents at their installed path.** Documents in this repository live at `docs/specs/<file>.md` and are installed to the same path in the target repository, so a relative markdown link resolves in both places. Use markdown links, never bare paths or code spans.
   - **Target-repository paths need no link.** Paths like `./specs/features/...` or `specs/architecture.yaml` are resolved in whatever project the skill runs in.
   - **End with a report step** telling the agent what to tell the user — the file path written and a one-line summary.

5. **Report** the skill path and the situations it will trigger on.
