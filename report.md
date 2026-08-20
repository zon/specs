Compose target directory with source-owned layout in targetdir

RelPath builds each path from the target's directory and the source's
per-kind layout. Docs write as-is, since the doc layout already names
the docs directory. All paths, ownership manifests, and removal
behavior are unchanged.

A test pins the composition for every target and kind.

Ralph item 1 completed
