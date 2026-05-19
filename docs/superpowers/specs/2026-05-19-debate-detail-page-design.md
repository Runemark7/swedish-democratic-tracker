# Debate Detail Page — Design Spec

**Date:** 2026-05-19  
**Status:** Approved

## Problem

`SpeechDetailPage` shows one politician's statement in isolation. When the speech is a "Svar på interpellation", the user has no way to:
- Read the original interpellation text
- Navigate to a dedicated debate view
- See all participants in that debate

The "Om debatten" block only shows the document title; there is no link for non-betänkande document types.

---

## Architecture

### Backend — new endpoint + enriched domain

**1. New domain type** in `backend/internal/votes/domain/vote.go`:

```go
type Intressent struct {
    IntressentID string `json:"intressentId"`
    Name         string `json:"name"`
    Party        string `json:"party"`
    Role         string `json:"role"` // e.g. "Upphovsman", "Svar"
}
```

`DocumentStatus` gains `Intressenter []Intressent`.

**2. Update `FetchDocumentStatus`** in `backend/internal/votes/adapters/riksdagen/client.go`:
- Parse full `dokintressent.intressent[]` list: `intressent_id`, `namn`, `partibet`, `roll`
- Remove the current single-intressent shortcut

**3. New endpoint** `GET /documents/:dokId/full` in `backend/internal/votes/adapters/http/handler.go`:
- Calls `svc.GetDocumentStatus(ctx, dokID)`
- Returns full payload including `bodyHtml` and `intressenter`
- Existing `GET /documents/:dokId` is unchanged (lightweight, no bodyHtml)

**4. `api/openapi.yaml`:**
- Add `RiksdagDocumentFull` schema (extends `RiksdagDocument` with `bodyHtml` + `intressenter`)
- Add `GET /documents/{dokId}/full` path

---

### Frontend — new page + hook

**Route:** `/debatt/:dokId` → `DebateDetailPage`  
**File:** `frontend/src/features/speeches/DebateDetailPage.tsx`

**New hook:** `useDocumentFull(dokId)` added to `useDemocracy.ts`  
Calls `GET /documents/:dokId/full`, enabled when `dokId` is truthy.

**`DebateDetailPage` layout (top → bottom):**

| Block | Content |
|---|---|
| Back link | `← Tillbaka` |
| Header | Doc type badge (`INTERPELLATION`) + title + date |
| Summary | `summary` field if present |
| Interpellation text | `bodyHtml` via `dangerouslySetInnerHTML` with `riksdagen-doc` CSS class |
| Debattdeltagare | `intressenter` list — avatar initial + name + party badge, each linking to `/politicians/:intressentId` |
| Alla anföranden | Chronological list via `useSpeechesByDocument(dokId)`, rendered with existing `SpeechRow` |
| Source link | `Källa: riksdagen.se →` linking to `https://www.riksdagen.se/sv/dokument-och-lagar/dokument/{type}/{beteckning}/` |

**`SpeechDetailPage` change:**  
In the "Om debatten" block, add `→ Se hela debatten` link to `/debatt/:relDokId` for all document types (current code only links for `bet` type).

**App.tsx:** Add route `<Route path="/debatt/:dokId" element={<DebateDetailPage />} />`.

**Regenerate types:** `cd frontend && npm run generate:api` after openapi update.

---

## Data Flow

```
User visits /anforanden/208
  → SpeechDetailPage fetches speech (relDokId = "ip_dok_id_xyz")
  → "Om debatten" block shows title + "→ Se hela debatten" link

User clicks "Se hela debatten"
  → /debatt/ip_dok_id_xyz
  → DebateDetailPage calls GET /documents/ip_dok_id_xyz/full
  → Renders: interpellation text + debattdeltagare + all speeches
```

---

## Files Changed

| File | Change |
|---|---|
| `backend/internal/votes/domain/vote.go` | Add `Intressent` type; add `Intressenter []Intressent` to `DocumentStatus` |
| `backend/internal/votes/adapters/riksdagen/client.go` | Parse full intressenter in `FetchDocumentStatus` |
| `backend/internal/votes/adapters/http/handler.go` | Add `getDocumentFull` handler + register route |
| `api/openapi.yaml` | Add `RiksdagDocumentFull` schema + `/documents/{dokId}/full` path |
| `frontend/src/shared/api-contract.ts` | Regenerated (do not hand-edit) |
| `frontend/src/hooks/useDemocracy.ts` | Add `useDocumentFull` hook |
| `frontend/src/features/speeches/DebateDetailPage.tsx` | New page |
| `frontend/src/features/speeches/SpeechDetailPage.tsx` | Add "Se hela debatten" link in "Om debatten" block |
| `frontend/src/App.tsx` | Add `/debatt/:dokId` route |

---

## Out of Scope

- Caching debate pages in the database (always proxied live from Riksdagen)
- Supporting document types beyond `ip` (frågasvar `fr`, motioner `mot`) — same code works but not specifically tested
- AI summarisation of debates
