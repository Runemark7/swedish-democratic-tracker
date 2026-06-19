# Beslut detail: debate schedule + plain-language summary

**Date:** 2026-06-19
**Feature:** `votes` (backend) + `BeslutDetailPage` (frontend)
**Status:** design approved, pending spec review

## Problem

Some `/beslut/<bet>` pages carry almost no usable content. The document body
renders, but the **Debatten** section is empty and just says "Inga registrerade
anföranden för det här beslutet ännu." The reader gets no sense of *when* the
matter was (or will be) debated, and the **Sammanfattning** often falls back to
the dense `summary` field — hard to read.

Concrete example: `/beslut/KU45` (Bifall, 16 jun). Decided with little/no
chamber debate; the page is a near-empty shell.

## Root finding

`dokumentstatus/{dok_id}.json` exposes a `dokuppgift` block we currently
discard. Three fields are reliable, factual, and sourced:

| `kod`            | Meaning                          | Example value         |
|------------------|----------------------------------|-----------------------|
| `debattdatumtid` | Debate slot (empty if unscheduled) | `2026-06-17 00:00:00` |
| `beslutdatumtid` | Decision datetime                | `2026-06-17 00:00:00` |
| `statustext`     | Lifecycle text                   | `Ärendet är avslutat` |
| `notis`          | "Beslut i korthet" — plain Swedish HTML summary | `<p>KU bedömer att…</p>` |

The detail endpoint already fetches `dokumentstatus` **live** via
`svc.GetDocumentStatus(dokID)` in `votes/adapters/http/handler.go`. So these
fields can be threaded through with **no DB migration and no ingestion change**.

Note on scope (explicitly excluded): the Riksdagen `anforandelista?bet=` filter
is loose and anföranden are keyed by the protokoll `dok_id` / `rel_dok_id`, not
the betänkande `dok_id` — which is why speeches rarely link. Fixing that linkage
is a separate, larger task and is **out of scope** here.

## Design

### Fact / interpretation discipline

Every new label is a raw Riksdagen datum shown verbatim with no editorial
spin. We never assert "this had no debate" (we cannot prove a negative from the
data); we state the sourced date and whether anföranden are registered, and let
the reader conclude.

### Backend (`votes`)

1. `domain.DocumentStatus`: add `DebattDate string`, `BeslutDate string`,
   `StatusText string`, `Notis string` (notis kept as HTML, like `BodyHTML`).
2. `riksdagen/client.go` `FetchDocumentStatus`: decode `dokuppgift.uppgift[]`,
   index by `kod`, populate the four fields. `dokuppgift.uppgift` may decode as
   object or array — handle both (same pattern already used elsewhere).
3. `adapters/http/handler.go`: in **both** response branches (no-votes metadata
   branch and votes-in-DB branch), add `debattDate`, `beslutDate`, `statusText`,
   `notis` to the response map when the `ds` lookup succeeds.

No `api/openapi.yaml` change required — the beslut detail response is an ad-hoc
`map[string]any`, not a typed spec endpoint (matches current handler style).

### Frontend (`BeslutDetailPage.tsx`)

Read the new optional fields off `data`. Two render changes:

**A. Empty Debatten branch** — replace the single fallback line with a
date-aware label. Parse `debattDate` (`YYYY-MM-DD HH:MM:SS`) to a `Date`,
compare to today, format Swedish (`17 juni 2026`):

- `debattDate` in the future → **"Debatt planerad: 17 juni 2026"**
- `debattDate` empty/absent → **"Debatt ännu inte planerad"**
- `debattDate` in the past, no speeches → **"Debatten hölls 17 juni 2026 ·
  inga anföranden registrerade här"**

When speeches *do* exist, the current speaker list is unchanged.

**B. Sammanfattning card** — prefer `notis` when present:

- `notis` present → render its HTML (reuse the existing `DocumentBody`/
  `riksdagen-doc` wrapper pattern; notis is trusted government markup like the
  body) under the existing "Sammanfattning" header.
- else `summary` → current plain-text paragraph.
- else → current "Riksdagen har inte publicerat någon sammanfattning…" line.

### Sourcing

Both additions are Riksdagen `dokumentstatus` data already covered by the
existing `riksdagen` data source — no new `docs/data-sources/` entry, no
`SourceRegistry` regen, no diagram change (CLAUDE.md rule 11 not triggered:
same source, same endpoint).

## Components touched

- `backend/internal/votes/domain/vote.go` — 4 struct fields
- `backend/internal/votes/adapters/riksdagen/client.go` — parse `dokuppgift`
- `backend/internal/votes/adapters/http/handler.go` — 2 response maps
- `frontend/src/features/votes/BeslutDetailPage.tsx` — Debatten + Sammanfattning
- `frontend/src/features/votes/api.ts` / types — optional new fields

## Verification

- `go build -C backend ./...`
- `cd frontend && npx tsc --noEmit`
- Manual: `/beslut/KU45` shows the beslut/debate date label and the "Beslut i
  korthet" notis text; a still-upcoming betänkande shows "Debatt planerad: …".

## Out of scope

- Fixing anförande→betänkande linkage (separate task).
- Any DB migration, ingestion worker, or scheduler change.
- Pulling the chamber-wide planning calendar.
