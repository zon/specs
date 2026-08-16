Create directories the update needs

The update command creates every directory it needs before writing
into a target. A run against a fresh repository builds the target
directory, its skills and agents directories, and each definition's
directory. The manifest directory appears too, even when no definitions
write.

Added tests assert Write creates the directories for a skill and an
agent, SaveOwned creates the target directory for the manifest, and an
update run creates them in a fresh repository.

Ralph item 12 completed
