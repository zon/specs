Remove stale definitions from the target

The update command now removes rendered definitions the source no longer
lists. A full run prunes owned skills and agents. Update skills prunes
only skills, and update agents only agents. The targetdir module deletes
each stale file and drops it from the manifest.

Tests cover a removed skill and a removed agent leaving the target, a
scoped run leaving the other kind alone, and a built-binary run removing
a stale skill.

Ralph item 14 completed
