Read the same frontmatter fields for every target

A shared frontmatter reader parses the name, description, and tools fields
from each definition file. It never depends on the target, so claude and
opencode both use the same fields. The reader lives in the new
internal/frontmatter module. The update command calls it for every
definition from a source.

Added tests cover reading all three fields, an inline tools list, a file
without frontmatter, unterminated frontmatter, a missing file, and the
update command reading the same agent for both targets.

Ralph item 4 completed
