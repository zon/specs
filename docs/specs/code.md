# Writing Code

## Before You Start

Read `specs/architecture.yaml` before writing any code — it is the module map for the entire codebase. See [Architecture Format](architecture.md).

If you are working on a feature, also read its architecture file at `specs/features/<component>/<feature>/architecture.yaml` if one exists. A feature architecture file describes the modules introduced or modified by that feature: their paths, roles, and whether each is an orchestration or implementation module.

## Module Placement

Every piece of code belongs in a specific module. Before writing, ask:

1. **Does an existing module own this concern?** Check `specs/architecture.yaml`. If a module already covers the concern, add the code there rather than duplicating logic or creating a parallel path.
2. **Does an orchestration file assign this code to a specific module?** Orchestration files (`specs/features/<component>/<feature>/orchestration.md`) are implementation contracts — the `**Module:**` annotations are binding. Place the code in the module named by the orchestration. See [Orchestration Format](orchestration.md#module-structure).
3. **Is there no existing home?** If neither an existing module nor an orchestration file covers the concern, determine whether it belongs in an existing module (by expanding its scope) or requires a new module. If a new module is needed, update `specs/architecture.yaml` before writing the code.

## Categories

Each module declares a `category`, and the category's `description` and `signatures` in `specs/architecture.yaml` are the authoritative statement of what may live there. The glossary defines the common categories — [pure](glossary.md#pure-module), [orchestration](glossary.md#orchestration-module), [implementation](glossary.md#implementation-module). The rules below say what to do when code does not fit the category it landed in. For any category a project defines beyond these, derive the constraints from its own `description` and `signatures`.

### Pure Modules

- If a pure module accumulates side-effectful code, move that code into an implementation module.

### Orchestration Modules

Modules marked `orchestration: true` in `specs/architecture.yaml` are orchestration modules.

- Do not add code to an orchestration module that an orchestration file does not ask for. If an orchestration file does not call it, it does not belong there.
- If an orchestration module is accumulating logic that is not pure coordination, move that logic into an implementation module.

### Implementation Modules

- Keep each module focused on its declared concern. Do not extend a module's scope without updating `specs/architecture.yaml`.
- Prefer deepening an existing module over creating a new one for the same concern.
- Expose only what callers need; keep internal details unexported.

## Related

- [Testing](testing.md) — what belongs in tests, and where mocks live
- [Glossary](glossary.md) — definitions of deep, pure, orchestration, and implementation modules
