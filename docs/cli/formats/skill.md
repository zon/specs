# Skill Format

`zpecs update skills` reads the neutral format for skill definitions.

## File Location

A skill definition lives at `skills/<name>.md` in the source.

## Structure

A skill is a markdown file with frontmatter and a body:

```markdown
---
name: spec-prose-editor
description: Reviews prose against the prose guidelines. Use when the user asks to proofread, edit, or check the writing in docs, comments, or agent prose.
---

# Prose Editor

The steps an agent follows when the skill runs.
```

| Field | Meaning |
|---|---|
| `name` | The identifier, matching the file name |
| `description` | What the skill does, and when to use it |

The body holds the instructions.
