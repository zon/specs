# Specs

The document formats and standards for spec-driven development with AI coding agents.

Everything in this directory is installed into a project at `docs/specs/` in the target. The paths are the same in both places, so every link below resolves identically wherever it is read.

## Directory Structure

```
/specs/
├── architecture.yaml
├── README.md
└── features/
    └── <component>/
        └── <feature>/
            ├── spec.md
            └── orchestration.md

/projects/
└── <slug>.yaml
```

`/specs/architecture.yaml` covers the **current** components of the application ([Architecture Format](architecture.md)). `/specs/README.md` is an index of the features below it.

Specs and orchestrations are co-located under `/specs/features`:

- `spec.md` — behavioral requirements and scenarios ([Spec Format](spec.md))
- `orchestration.md` — idealized domain logic ([Orchestration Format](orchestration.md))

Project files live at `/projects/<slug>.yaml` and list the requirements a coding agent works through, one per iteration ([Project Format](project.md)).

See [Component](glossary.md#component) and [Feature](glossary.md#feature) in the glossary.

## Formats

| Document | Purpose |
|---|---|
| [Spec](spec.md) | Behavior contracts — structured requirements and scenarios |
| [Orchestration](orchestration.md) | Idealized domain logic as an implementation contract |
| [Architecture](architecture.md) | The components of an application, in YAML |
| [Project](project.md) | A list of requirements for a coding agent to work through |

## Standards

| Document | Purpose |
|---|---|
| [Writing Code](code.md) | Component placement, and what belongs in each type of component |
| [Testing](testing.md) | Test layers, mock rules, and module boundaries in tests |
| [Writing Requirements](requirements.md) | What makes a good unit of work |
| [Agent Prompts](prompts.md) | How to structure a single-task prompt |
| [Glossary](glossary.md) | Terms used throughout these documents |

## Order of Authorship

For a new feature, author the formats in order:

1. **Spec** — what the system must do, as observable behavior.
2. **Orchestration** — the shape of the domain logic, with modules assigned.

Each step is optional in principle, but an orchestration drafted without a spec has no behavior to shape, and a spec handed to an agent without one leaves it to invent implementation shapes. Skipping a step moves the decision, it does not remove it.

The **architecture** records what the code produced, so it is written after the code, not ahead of it.

A [project](project.md) stands apart from that chain. It lists requirements a runner works through one at a time, whatever their source.
