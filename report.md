Resolve the repository root from the process working directory

update.Run no longer passes a working-directory literal to repo.Root,
so the update orchestration body holds only named step calls, domain
conditions, and return values. repo.Root resolves the process working
directory itself instead of taking the directory as an argument.

Tests now chdir into the directory under test before calling repo.Root,
covering the repository root, a nested subdirectory, and a directory
outside any repository.

Ralph item 0 completed
