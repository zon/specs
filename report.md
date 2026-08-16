Read skill and agent definitions from a local source

A local source mirrors the repository layout, so the update command reads
skill definitions from skills/<name>/SKILL.md and agent definitions from
agents/<name>.md. The reader does not find a definition in any other
location, and a missing source directory errors. A new internal/source
module performs the read, and the update orchestration calls it when a
--source flag names a directory.

Added tests cover a skill at the canonical path, a misplaced skill staying
unfound, the same for agents, a missing source directory erroring, and a
built-binary run against a source.

Ralph item 1 completed
