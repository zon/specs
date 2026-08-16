Write definitions into each target's directories

The update command now places every definition, keyed by its source
name, into the target's directories. A skill writes its own content to
skills/<name>/SKILL.md. An agent writes its form to agents/<name>.md.
Both sit under a .claude or .opencode root. The path never comes from
the rendered fields. A new internal targetdir module resolves the paths
and writes the files, creating the directories it needs.

Tests cover the four target paths, directory creation, both targets
writing a skill and an agent, and a built-binary run against the claude
target.

Ralph item 11 completed
