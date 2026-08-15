# Specs

The document formats and standards for spec-driven development with AI coding agents.

Everything in this directory installs into a project at `docs/specs/` in the target. The paths are the same in both places, so every link below resolves identically wherever it is read.

## Directory Structure

```
/specs/
├── architecture.yaml
├── README.md
└── features/
    └── <component>/
        └── <feature>/
            └── spec.md

/projects/
└── <slug>.yaml
```

`/specs/architecture.yaml` covers the **current** components of the application ([Architecture Format](architecture-outline.md)). `/specs/README.md` is an index of the features below it.

Specs live under `/specs/features`:

- `spec.md` — behavioral requirements and scenarios ([Spec Format](spec.md))

Orchestration is a code pattern: the coding agent writes the coordination logic into [orchestration modules](glossary.md#orchestration-module) during implementation ([Orchestration Pattern](orchestration.md)).

Project files live at `/projects/<slug>.yaml` and list the requirements a coding agent works through, one per iteration ([Project Format](project.md)).

See [Component](glossary.md#component) and [Feature](glossary.md#feature) in the glossary.

## Formats

| Document | Purpose |
|---|---|
| [Spec](spec.md) | Behavior contracts — structured requirements and scenarios |
| [Architecture](architecture-outline.md) | The components of an application, in YAML |
| [Project](project.md) | A list of requirements for a coding agent to work through |

## Standards

| Document | Purpose |
|---|---|
| [Orchestration](orchestration.md) | The pattern that separates coordination logic from implementation detail |
| [Architecture](architecture.md) | Component placement, and what belongs in each type of component |
| [Writing Requirements](requirements.md) | What makes a good unit of work |
| [Agent Prompts](prompts.md) | How to structure a single-task prompt |
| [Glossary](glossary.md) | Terms used throughout these documents |

## Order of Authorship

For a new feature, author the formats in order:

1. **Spec** — what the system must do, as observable behavior.

Orchestration is not authored ahead of the code. The architect decides where the orchestration pattern applies, the coding agent writes the orchestration code, and the architecture records which components are orchestration modules.

The **architecture** records what the code produced, so it is written after the code.

A [project](project.md) stands apart from that chain: its requirements can come from any source.
