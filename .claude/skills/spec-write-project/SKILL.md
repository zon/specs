---
name: spec-write-project
description: Creates and validates a project file defining work for a coding agent to execute, one item per iteration, drawn from a feature's spec, architecture, and orchestration.
---

# Write Project

Create a well-formed project file based on the user's description of the work to be done.

A project is a YAML or JSON file containing an array of work items. Each element of the array is an **item**: one unit of work, one iteration. The item must be self-contained — the agent that implements it sees the item and the project file, not the spec or orchestration it was drawn from. No completion field belongs in the file; completion is the runner's record, not the project's.

## Steps

1. **Understand the work.** If the user's request is vague, ask clarifying questions:
   - What feature or change does this project cover?
   - Does it target a documented feature directory under `specs/features/`?
   - Does the work require a version bump?

2. **Locate the feature directory** if the project targets a documented feature. Feature directories live under `specs/features/<component>/<feature>/` and may contain any of `spec.md`, `orchestration.md`, and `architecture.yaml` — all optional.

3. **Read the project format docs** to refresh your understanding:
   - [docs/specs/project.md](docs/specs/project.md)

4. **Read the coding and testing standards** so items are consistent with how this codebase is written and tested:
   - [docs/specs/code.md](docs/specs/code.md)
   - [docs/specs/testing.md](docs/specs/testing.md)
   - The repository's own `AGENTS.md` or `CLAUDE.md`, for project-specific rules such as which dependencies must never run in tests, and which files a version bump has to touch.

5. **Check the module category** for every module the items will touch. Read `specs/architecture.yaml`. If the project targets a feature and `<feature-dir>/architecture.yaml` is present, read that too — it describes modules introduced or changed by the feature. Look up the `category` field for each affected module path. The category's `signatures` and `orchestration` flag (defined in [docs/specs/architecture.md](docs/specs/architecture.md)) determine what form the code and tests must take. Apply those constraints when writing `items`, `code`, and `tests` for each item.

6. **Draft orchestration-based items.** If `<feature-dir>/orchestration.md` is present, read it and create an item for each implementation shape it defines. Source `code` and `tests` entries exclusively from it — never invent shapes. Every helper called from a `body` gets its own item with a fully-specified `code` entry.

7. **Draft scenario-based items.** If `<feature-dir>/spec.md` is present, read it and add its scenarios to any matching items from step 6. If a scenario doesn't correspond to an orchestration item, create a new item for it with `scenarios` only.

8. **Draft remaining items** for any work not covered by the orchestration or spec — additional constraints, edge cases, operational requirements, and the version bump if needed.

9. **Write the file** at `./projects/<slug>.yaml`, following the format in [docs/specs/project.md](docs/specs/project.md) and the guidance in [docs/specs/requirements.md](docs/specs/requirements.md). Use the conventional shape — a top-level mapping whose `requirements` list holds the item array — and give every item a `slug` so commit messages name the work. Each item needs at least one of `items`, `scenarios`, `code`, or `tests`.

10. **Validate.** Confirm the file parses and that the item array resolves to at least one non-empty item. Empty entries — null, `false`, `0`, blank strings, `{}`, `[]` — are typically dropped during resolution, so never leave a placeholder entry in the list. If the project's runner provides a validation command, use it — for the `ralph` runner that is `ralph validate ./projects/<slug>.yaml`.

11. **Report** the file path, the item query the runner needs (`.requirements` for the conventional shape), and a one-line summary of what the project covers.
