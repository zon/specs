Move the per-kind file layout into internal/source

RelPath returns a definition's path in its kind's layout:
skills/<name>/SKILL.md for skills, agents/<name>.md for agents, and
docs/zpecs/<name>.md for docs. Unknown kinds fall back to the agent
path. The architecture entry for internal/source now records the layout.

Tests for RelPath cover the skill, agent, and doc paths.

Ralph item 0 completed
