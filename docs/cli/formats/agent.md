# Agent Format

`zpecs update agents` reads the neutral format for agent definitions.

## File Location

An agent definition lives at `agents/<name>.md` in the source.

## Structure

An agent is a markdown file with frontmatter and a body:

```markdown
---
name: prose-editor
description: Reviews prose against the prose guidelines. Use when the user asks to proofread, edit, or check the writing in docs, comments, or agent prose.
---

# Prose Editor

The instructions an agent follows when called.
```

| Field | Meaning |
|---|---|
| `name` | The identifier, matching the file name |
| `description` | What the agent does, and when to use it |

The body holds the instructions.
