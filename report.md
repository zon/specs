Render only what each update command names

The update command reads definitions by scope. A full update reads
skills and agents. Update skills reads skills only, and update agents
reads agents only. The source module exposes a scoped read for each
kind.

Tests cover scoped reads finding only their kind, missing scopes
erroring, in-process runs, and a built-binary run of the scoped
commands against a real source.

Ralph item 15 completed
