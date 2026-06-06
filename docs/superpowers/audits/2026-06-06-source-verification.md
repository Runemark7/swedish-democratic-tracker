# Source Verification Audit — 2026-06-06

Pass two. Tracks every registered source against the verification criteria
defined in `docs/superpowers/specs/2026-06-06-source-verification-design.md`.

## Main table

| id | URL liveness | Schema match | Repro narrative OK? | UI consumers | Backend client | Proposed status | Gaps |
|---|---|---|---|---|---|---|---|
| derived-agenda |  |  |  |  |  |  |  |
| kolada |  |  |  |  |  |  |  |
| kolada-spending |  |  |  |  |  |  |  |
| riksdagen |  |  |  |  |  |  |  |
| scb-befolkning |  |  |  |  |  |  |  |
| scb-kfmandat |  |  |  |  |  |  |  |
| scb-kostndrlt |  |  |  |  |  |  |  |
| scb-ltmandat |  |  |  |  |  |  |  |
| seed-budget-data |  |  |  |  |  |  |  |
| seed-party-goals |  |  |  |  |  |  |  |
| statskontoret-arsutfall |  |  |  |  |  |  |  |
| ted |  |  |  |  |  |  |  |
| wikimedia-svg |  |  |  |  |  |  |  |

## Appendix 1 — Untagged values

UI strings quoting a number/date sourced from external data but lacking `<SourceMarker>`.

| File:line | Value | Should reference source |
|---|---|---|

## Appendix 2 — Untracked backend clients

External HTTP calls / CSV downloads in `backend/internal/` whose target URL
does not map to a registered source.

| Backend file:line | Target URL | Should reference source (or "new") |
|---|---|---|

## Appendix 3 — Sources not yet in registry

Discoveries from Appendix 1/2 that warrant a new MD in `docs/data-sources/`.

| Proposed id | Upstream | Notes |
|---|---|---|
