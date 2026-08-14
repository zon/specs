# Project Format

A project is a YAML or JSON file holding a list of requirements. Each entry is one requirement — one unit of work, one iteration of a coding agent.

## File Location

Project files live at `./projects/<slug>.yaml`. The base name is the slug, which a runner conventionally uses as the branch name. See [Directory Structure](README.md#directory-structure).

## Shape

```yaml
- Reports can be exported as CSV from GET /reports/:id/export
- A request for a non-existent report ID returns 404
- A malformed report ID returns 400 with a descriptive error message
- Apply a semver minor bump to the app version
```

## Writing Requirements

One entry is one iteration, so work that needs three rounds is three entries. The agent sees the selected requirement and the rest of the file, so write each entry to stand on its own. See [Writing Requirements](requirements.md).

## Version Bumps

If the repo uses versioning, add an entry for the bump. Name the level, not the target version — the agent finds the current version and applies the bump. Each versioned resource is bumped on how its own interface changes:

- **patch** — bug fixes, refactoring, small internal changes
- **minor** — new features added in a backwards-compatible way
- **major** — breaking changes to the API, CLI, or behavior
