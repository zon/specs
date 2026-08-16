Read definitions from GitHub by default

The update command reads definitions from the GitHub repository when
the user gives no --source flag. A --source flag keeps reading from the
local directory it names.

A new clone module copies the repository into a temp dir, and the update
command removes the clone after the run. ZPECS_SOURCE overrides the
default repository, mostly for tests.

Tests cover reading from the default source in-process and as a built
binary, and a local source winning over the default.

Ralph item 0 completed
