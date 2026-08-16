# Agent Instructions

This repository holds the spec formats and standards that projects install, and the Go source of the zpecs CLI that installs them.

## Prose

Read [prose guidelines](docs/specs/prose.md) before writing any kind of prose. This includes docs, code comments, git descriptions, or agent communication.

After writing or editing prose, invoke the `prose-editor` subagent to review it. Pass the files you touched as the scope.

## Editing Docs

Documents live in [docs/specs/](docs/specs/README.md) and install to `docs/specs/` in the target repository. The path is identical in both places, which is what makes relative links work everywhere.

- **Link between documents with plain relative links.** From inside `docs/specs/`, that is a bare filename: `[Architecture Format](architecture-outline.md)`.
- **Never rewrite a link to an absolute URL.** Install is a file copy. A raw GitHub URL would pin the target to whatever this repo looked like at install time.
- **Paths belonging to the target project stay unlinked.** `specs/architecture.yaml`, `specs/<path>.md`, and `projects/<slug>.yaml` are resolved wherever a skill runs, so write them as code spans.

## Editing Skills

Author skills in the central `skills/<name>/SKILL.md` directory. Use the `write-skill` skill, or follow the same rules by hand:

- The frontmatter `name` must match the directory name.
- Reference documents with markdown links at their installed path: `[docs/specs/architecture.md](docs/specs/architecture.md)`. Never a bare path, never a code span.
- Reference documentation rather than restating it. A skill that repeats a format goes stale when the format changes.

## Before Committing

```bash
just check
go test ./...
```

`just check` verifies every relative markdown link in the docs and skills resolves. `go test ./...` builds the CLI and runs its tests. Run both after any change.

## Keeping It Neutral

These documents describe how to write specs, architecture, and projects, and how to apply the orchestration pattern to code, not how any particular tool consumes them. When a rule depends on runner behavior, state the rule and note that the runner decides, rather than documenting one runner's flags. Completion state in particular never belongs in a project file.
