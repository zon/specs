Expose update, update skills, and update agents commands

The CLI now parses its command line and dispatches to the update command,
which resolves a skills, agents, or full scope from its arguments. Unknown
commands and scopes error with usage on stderr and a non-zero exit.

Tests added cover scope resolution, command recognition and rejection at the
run level, and a built-binary run of each command plus rejection of an unknown
command.

Ralph item 17 completed
