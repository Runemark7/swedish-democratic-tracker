# Curation happens in the bot chat, through a token-protected write API, with all data in Postgres

**Status:** Accepted, 2026-10-27. Supersedes ADR-001, Concern A. Scope confirmed 2026-10-28.

There is no admin UI. The Redaktör approves Löften and Föreslagna kopplingar by answering the bot in chat, and the bot writes those decisions through a backend write API. The API is protected by a token that only the bot holds. Every change is recorded in an append-only audit table (who, on whose behalf, when, old value and new value, and the chat reference), so any change can be traced and undone. All editorial data lives only in Postgres, never in repo data files or migrations. This removes a UI and a login to build and secure, keeps one source of truth that survives redeploys, and keeps every change reviewable without hand-written SQL against production.

The write API, the audit table and all manual confirmation ship in Phase 2 (Löften). Phase 1 (Motioner) makes only links that are certain without a person: a reference in Riksdagen's data, a cited motion beteckning, or identical yrkande wording. From Phase 2 the same API and Veckobatch also decide the motion links and Återkommande krav that Phase 1 could not make automatically.

## Considered Options

- **A hidden admin page behind a login.** Rejected by the owner: curation already happens in chat.
- **A versioned data file in the repo, with approvals as data PRs (or seed migrations, as in ADR-001).** Rejected. It splits data between the repo and Postgres.
- **The bot writing SQL directly.** Rejected. It gives no validation, no idempotency and no guaranteed audit entry.
