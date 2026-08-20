Fold frontmatter into render

Merge the frontmatter parsing code into internal/render, its only
caller, so one component parses and renders agents. Move the frontmatter
tests into render's test file. Remove internal/frontmatter from
specs/architecture.yaml and the filesystem.

The tests cover parsing fields, bodies, and errors.

Ralph item 0 completed
