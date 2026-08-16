Replace only the files the system wrote

The update command replaces only the files it wrote on a later run. A
file in the target directory it did not write stays unchanged.

The targetdir module keeps a per-target manifest of written paths. The
update command reads it before a run and saves the updated list after.

Tests cover a foreign file surviving a run, a second run replacing an
owned file, the manifest round trip, separate manifests per target, and
a built-binary run.

Ralph item 13 completed
