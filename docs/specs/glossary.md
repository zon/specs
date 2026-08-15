# Glossary

Terms used throughout these documents. A runner that executes projects will have its own vocabulary for the mechanics of doing so — how a requirement is selected, how completion is recorded — and that belongs in the runner's documentation, not here.

## Component

A collection of resources — scripts, runtimes, assets, config, or source code modules — the project is built from. Components are listed in [`specs/architecture.yaml`](architecture-outline.md), each with a path, a description, and whether it is an orchestration module.

## Feature

A coherent slice of user-facing or system-facing behavior — something a user can do, or something the system does on their behalf. Good feature names describe what the system does (`auth`, `payments`, `notifications`), not how it does it (`jwt-handler`, `stripe-client`). If a feature grows too large to read comfortably, split it by sub-feature rather than by implementation detail.

## Requirement

A single behavior the system must have, stated as observable behavior rather than implementation. Requirements live in a [spec](spec.md), where they use RFC 2119 keywords (SHALL, MUST, SHOULD, MAY) to signal strength, and in a [project](project.md), where each one is the unit of a single iteration.

## Scenario

A concrete, testable example of a requirement in action, written in Given/When/Then form. Scenarios live in a spec.

## Project

A YAML or JSON file holding a list of requirements for a coding agent to work through, one per iteration. See [Project Format](project.md).

## Deep Module

A module with a simple interface but complex implementation. Deep modules hide implementation complexity behind a clean, minimal API, providing powerful functionality without exposing internal details. This design principle maximizes the benefit-to-complexity ratio by minimizing the cognitive load on users while maximizing utility.

## Pure Module

A module that contains only value objects and pure functions — code with no side effects. Pure modules do not perform I/O, mutate shared state, or call external services. They are fully testable with unit tests alone, with no need for mocks or integration setup.

## Implementation Module

A module that contains concrete technical implementation details and low-level operations. Implementation modules execute specific tasks such as database queries, API calls, cryptographic operations, file I/O, or data transformations. These modules provide the actual "how" of executing operations rather than coordinating what operations to execute.

Each implementation module covers a single deep concern — one cohesive area of functionality with a simple interface over hidden complexity.

## Orchestration Module

A module that contains only domain logic for coordinating other modules. Orchestration modules define workflows, sequence steps, enforce business rules, and delegate to implementation modules. They say what should happen and when, never how: no string construction, no format literals, no I/O, no external calls, no helper utilities.

A small app typically contains a single orchestration module. As it grows, the orchestration module should be split along deep concern boundaries — each resulting module coordinates one deep concern.
