# Architecture Guidelines

## Before You Start

If it exists, read `specs/architecture.yaml` before writing code. See [Architecture Format](architecture-outline.md).

## Component Placement

Every piece of code belongs in a specific component. Before writing, ask:

1. **Does an existing component own this concern?** Check `specs/architecture.yaml`. If so, add the code there rather than duplicating logic.
2. **Does the orchestration pattern assign this code to a component type?** Coordination logic belongs in an [orchestration module](glossary.md#orchestration-module). Low-level work belongs in an [implementation module](glossary.md#implementation-module). See [Orchestration Pattern](orchestration.md).
3. **Is there no existing home?** If neither applies, expand an existing component's scope or create a new one. Add a new component to `specs/architecture.yaml` once the code is written.

## When to Create Components

Keep components to a minimum. Create a new component for one of these reasons:

- **Separate orchestration from implementation.** Create one component for coordination and one for the low-level work it delegates. See [Orchestration Pattern](orchestration.md).
- **Separate major concerns.** Give each major concern its own [deep module](glossary.md#deep-module): a simple interface over a complex implementation.
- **Share common logic.** Move logic that several major concerns use into a shared component.

When none of these apply, add the code to an existing component.

## Component Types

Each component is an [implementation module](glossary.md#implementation-module) unless `specs/architecture.yaml` sets `orchestration: true`. The flag makes it an [orchestration module](glossary.md#orchestration-module). A component with no code is neither.

### Orchestration Modules

- A small app typically has one.
- Do not add code to an orchestration module unless its orchestration calls it.
- If an orchestration module is accumulating logic that is not pure coordination, move that logic into an implementation module.
- Split a growing orchestration module along deep concern boundaries.

### Implementation Modules

- Keep each component focused on its declared concern. Do not extend a component's scope without updating `specs/architecture.yaml`.
- Prefer deepening an existing component over creating a new one for the same concern.
- Expose only what callers need.
- If a component meant to hold only pure functions accumulates side-effectful code, move that code into a different implementation module.
- Interfacing with CLI input parsers or HTTP hosting is usually an implementation concern.

## Related

- [Glossary](glossary.md) — definitions of deep, pure, orchestration, and implementation modules
