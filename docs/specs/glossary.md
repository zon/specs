# Glossary

Terms used throughout these documents. A runner that executes projects will have its own vocabulary for the mechanics — how it selects a requirement, how it records completion — and that belongs in the runner's documentation, not here.

## App

A runnable the project ships — a server, worker, or CLI. Apps can group [specs](specs.md) by the runtime that serves them.

## Component

A collection of resources — scripts, runtimes, assets, config, or source code modules — that build the project. [`specs/architecture.yaml`](architecture-outline.md) lists each component with a path, a description, and whether it is an orchestration module.

## Feature

A coherent slice of user-facing or system-facing behavior — something a user can do, or something the system does on their behalf. Good feature names describe what the system does (`auth`, `payments`, `notifications`), not how it does it (`jwt-handler`, `stripe-client`). If a feature grows too large to read comfortably, split it by sub-feature rather than by implementation detail. A [spec](specs.md) path ends in a feature.

## Requirement

A single behavior the system must have. State it as observable behavior, not implementation. Requirements live in a [spec](specs.md), where they use RFC 2119 keywords (SHALL, MUST, SHOULD, MAY) to signal strength, and in a [project](project.md), where each one is one iteration.

## Scenario

A concrete example of a requirement in action, in Given/When/Then form. Scenarios live in a spec.

## Project

A YAML or JSON file holding a list of requirements for a coding agent to work through, one per iteration. See [Project Format](project.md).

## Deep Module

A module with a simple interface over a complex implementation.

## Pure Module

A module that contains only value objects and pure functions — no I/O, no shared state, no external calls.

## Implementation Module

A module that does the low-level work — database queries, API calls, file I/O, data transformations. Each covers a single deep concern. See [Architecture Guidelines](architecture.md).

## Orchestration Module

A module that coordinates other modules. A small app typically has one. Split it along deep concern boundaries as it grows. See [Orchestration Pattern](orchestration.md).
