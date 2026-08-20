# Code Guidelines

We want clear, minimal code that is easy to understand and replace. Code should meet our requirements and standards but in the simplest possible way.

## Writing

* Each function and block should only do one specific thing
* Match current requirements and standards instead of the past or future
* Avoid functions with deep nesting. Return early or consider better architecture to keep code shallow and easy to read
* Code should be self-evident. Only write comments to summarize when clear code is infeasible

## Refactoring

* Gradually move towards ideal architecture and implementation. Always look for clearer, more minimal solutions
* Always refactor within the scope of work
* Ask for a project to be written when refactoring is out of the work scope. Name it `refactor-<slug>.yaml`

### Refactoring projects

Before asking for a project, check whether a similar one already exists in `./projects/`. If it does, ask for it to be updated instead of writing a new one. Write the project with the `write-project` skill.