# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for the Riksdagskollen project.

ADRs document significant technical decisions, the context behind them, and the options that were considered. They are written primarily for future developers (human or AI) to understand *why* things are the way they are.

## Format

Each ADR follows this structure:

- **Title**: Short descriptive name
- **Status**: Proposed | Accepted | Deprecated | Superseded
- **Context**: The situation and forces at play
- **Options**: Alternatives considered with pros/cons
- **Decision**: What was chosen and why
- **Consequences**: What follows from the decision

## Naming

Files are named `NNN-short-title.md` (e.g., `001-party-goals-data-strategy.md`).

## Index

| ADR | Title | Status |
|---|---|---|
| [001](001-party-goals-data-strategy.md) | Party goals data strategy | Accepted. Concern A superseded by [003](003-curation-in-chat-via-write-api.md); Concern B stands |
| [002](002-loftesstatus-is-vote-based.md) | Löftesstatus is vote-based, never kept or broken | Accepted |
| [003](003-curation-in-chat-via-write-api.md) | Curation happens in the bot chat, through a token-protected write API | Accepted |

Domain vocabulary lives in [`CONTEXT.md`](../CONTEXT.md) at the repo root. The current product spec is [`docs/spec/riksdagskollen-motioner-loften.md`](../docs/spec/riksdagskollen-motioner-loften.md).

Reference: https://adr.github.io/
