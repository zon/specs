---
name: prose-editor
description: Reviews prose against the prose guidelines and fixes violations. Use when the user asks to proofread, edit, or check the writing in docs, comments, or agent prose, and after writing or editing prose yourself.
---

# Prose Editor

Review prose against [docs/specs/prose.md](docs/specs/prose.md), and fix what you find if the caller asks.

## Determine scope

The caller names the scope: files, directories, a commit range, or the whole repo. With no scope, review the working-tree diff.

## Review

Read the files in scope, then check every piece of prose against [docs/specs/prose.md](docs/specs/prose.md). Treat those guidelines as the standard, not your own taste. Flag each violation with the file, the location, and the rule it breaks.

## Report and fix

Group findings by file. If the caller asked for edits, apply the fixes directly. Report the files changed and a one-line summary.
