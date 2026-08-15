# Writing Code

## Before You Start

Read `specs/architecture.yaml` before writing any code if it exists — it is the component map for the entire codebase. If it does not exist yet, the code you write is what the architecture will record; add it once written. See [Architecture Format](architecture.md).

## Component Placement

Every piece of code belongs in a specific component. Before writing, ask:

1. **Does an existing component own this concern?** Check `specs/architecture.yaml`. If a component already covers the concern, add the code there rather than duplicating logic or creating a parallel path.
2. **Does an orchestration file assign this code to a specific component?** Orchestration files (`specs/features/<component>/<feature>/orchestration.md`) are implementation contracts — the `**Module:**` annotations are binding. Place the code in the component named by the orchestration. See [Orchestration Format](orchestration.md#module-structure).
3. **Is there no existing home?** If neither an existing component nor an orchestration file covers the concern, determine whether it belongs in an existing component (by expanding its scope) or in a new one. If a new component is needed, add it to `specs/architecture.yaml` once the code is written.

## Component Types

Each component is either an [implementation module](glossary.md#implementation-module) or, when marked `orchestration: true` in `specs/architecture.yaml`, an [orchestration module](glossary.md#orchestration-module). A component that is not code — scripts, runtimes, assets, or config — is neither.

### Orchestration Modules

- Do not add code to an orchestration module that an orchestration file does not ask for. If an orchestration file does not call it, it does not belong there.
- If an orchestration module is accumulating logic that is not pure coordination, move that logic into an implementation module.

### Implementation Modules

- Keep each component focused on its declared concern. Do not extend a component's scope without updating `specs/architecture.yaml`.
- Prefer deepening an existing component over creating a new one for the same concern.
- Expose only what callers need; keep internal details unexported.
- If a component holds only pure functions and accumulates side-effectful code, move that code into a different implementation module.

## Related

- [Testing](testing.md) — what belongs in tests, and where mocks live
- [Glossary](glossary.md) — definitions of deep, pure, orchestration, and implementation modules
