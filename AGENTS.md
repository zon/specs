# Agent Instructions

This repository holds the spec formats and standards that projects install. Everything in it is documentation — there is no build and no source code.

## Prose

Read [prose guidelines](docs/specs/prose.md) before writing any kind of prose. This includes docs, code comments, git descriptions, or agent communication.

After writing or editing prose, invoke the `prose-editor` subagent to review it. Pass the files you touched as the scope.

## Editing Docs

Documents live in [docs/specs/](docs/specs/README.md) and install to `docs/specs/` in the target repository. The path is identical in both places, which is what makes relative links work everywhere.

- **Link between documents with plain relative links.** From inside `docs/specs/`, that is a bare filename: `[Architecture Format](architecture.md)`.
- **Never rewrite a link to an absolute URL.** Install is a file copy; there is no link rewriting step, and a raw GitHub URL would pin the target to whatever this repo looked like at install time.
- **Paths belonging to the target project stay unlinked.** `specs/architecture.yaml`, `specs/features/<component>/<feature>/`, and `projects/<slug>.yaml` are resolved wherever a skill runs, so write them as code spans.

## Editing Skills

Author skills in the central `skills/<name>/SKILL.md` directory. Use the `write-skill` skill, or follow the same rules by hand:

- The frontmatter `name` must match the directory name. Skills that `just install` ships keep the `spec-` prefix, because the installer only copies and prunes prefixed skills.
- Reference documents with markdown links at their installed path: `[docs/specs/code.md](docs/specs/code.md)`. Never a bare path, never a code span.
- Reference documentation rather than restating it. A skill that repeats a format goes stale when the format changes.

## Before Committing

```bash
just check
```

This verifies every relative markdown link in the docs and skills resolves. It is the only test this repository has, so run it after any edit that touches a link or renames a file.

## Keeping It Neutral

These documents describe how to write specs, architecture, orchestrations, and projects — not how any particular tool consumes them. When a rule depends on runner behavior, state the rule and note that the runner decides, rather than documenting one runner's flags. Completion state in particular never belongs in a project file.
