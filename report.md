Inline the checkDir helper in internal/source

ReadKinds now calls os.Stat directly, and the checkDir helper is gone.
The existing TestReadKindsErrorsOnMissingSource already covers the
error behavior, so no new tests were added. Removed the completed plan
from refactoring.md. Bumped VERSION to 0.1.10.

Ralph item 0 completed
