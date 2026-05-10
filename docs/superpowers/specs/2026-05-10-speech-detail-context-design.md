# Speech Detail Page — Context, Connections, Transcript

**Date:** 2026-05-10
**Status:** Design (approved by user, pending implementation plan)

## Context

`/anforanden/:id` currently shows just the speech body, the speaker's
name as text, the party as text, and the date. Three problems:

1. **No navigation.** The politician name is not a link; the party
   string is not a link. Users can read the speech but cannot pivot to
   the speaker's profile or the party page from here.
2. **No image.** The site already stores `politician.imageUrl` and
   uses it on `/politicians/:id`, on the new `SpeechRow` component, and
   on the party page. The detail page is the one place it's missing.
3. **No debate context.** A speech is part of a debate, but the page
   shows it in isolation. Users cannot see who answered, who spoke
   before, or what the debate was even about beyond the cryptic
   `topic_heading` (e.g. "Svar på interpellation 2024/25:734 om
   arbetsvillkor inom gigekonomin").

Goal: turn `/anforanden/:id` from a leaf page into a hub. From a single
speech, visitors should be able to (a) jump to the speaker's profile,
(b) jump to the speaker's party, (c) jump to the underlying betänkande
when the speech is part of a betänkande debate, and (d) see the rest
of the debate as a list of related speeches.

## Out of scope

- Per-anförande replies/interactions (Riksdagen does not return a
  structured reply tree; the chronological order is best-effort).
- Inline expand/collapse of other speeches' full text — clicking a
  related speech navigates to its own detail page, same as today.
- AI summary of the debate. Out of scope; if Riksdagen has a `summary`
  on the related document, we surface it; otherwise nothing.
- Moving the existing HTML render of the speech body. Stays where it
  is; we only add structure around it.

## Architecture

Two new layers added on top of existing infrastructure:

1. **Backend: one new endpoint** `GET /api/documents/:dokId` returning
   the parsed `DocumentStatus` (title, type, beteckning, summary,
   bodyHtml). Implemented as a thin chi route on the votes feature
   that delegates to the existing `votes.Service.GetDocumentStatus`,
   which already wraps the Riksdagen `dokumentstatus` API. No new
   business logic — just an HTTP surface.
2. **Frontend: rework SpeechDetailPage** into three vertical sections
   (politician card · current speech · debate transcript) plus a
   compact "Om debatten" header pointing at the related document.

Reuses everything already shipped:
- `<SpeechRow>` for the transcript list.
- `useSpeechesByDocument(dokId)` hook.
- `politician.imageUrl` + initials fallback (same pattern as
  `<SpeechRow>`).
- `votes.Service.GetDocumentStatus`.

## Components

### Backend: `GET /api/documents/:dokId`

Mounted in `backend/internal/votes/adapters/http/handler.go` (votes
feature already owns `dokumentstatus`). Route:

```go
r.Get("/documents/{dokId}", h.getDocument)
```

Handler returns the same JSON shape the existing
`/api/votes/{beteckning}/{punkt}` already builds for
`dokumentstatus`-fed fields. Specifically:

```json
{
  "dokId": "hb091ip453",
  "type": "ip",
  "title": "...",
  "subtitle": "",
  "summary": "",
  "beteckning": "",
  "session": "2024/25",
  "date": "2025-09-04"
}
```

When the document type is `bet` (betänkande), the response includes
the `beteckning` parsed from `dokumentstatus.dokument.beteckning`.
Frontend uses that field to construct a `/beslut/:beteckning` link.

For interpellations, motions, etc., `beteckning` is empty and the
frontend just shows the title without a link.

`bodyHtml` is **not** included — it's only used on the beslut page
and unnecessary here. Adding it would 10x the response size.

### Frontend: `useDocument` hook

```ts
// frontend/src/hooks/useDemocracy.ts
export function useDocument(dokId: string | undefined) {
  return useQuery<RiksdagDocument>({
    queryKey: ["document", dokId],
    queryFn: () => votesApi.getDocument(dokId ?? ""),
    enabled: !!dokId,
    staleTime: 5 * 60 * 1000,
  });
}
```

`votesApi.getDocument` is a new method:

```ts
// frontend/src/features/votes/api.ts
getDocument: (dokId: string) =>
  api.get<RiksdagDocument>(`/documents/${encodeURIComponent(dokId)}`),
```

`RiksdagDocument` type added to `frontend/src/shared/types.ts`:

```ts
export interface RiksdagDocument {
  dokId: string;
  type: string;          // "bet" | "ip" | "mot" | "prop" | ...
  title: string;
  subtitle?: string;
  summary?: string;
  beteckning?: string;   // present for bet/mot
  session?: string;
  date?: string;
}
```

### Frontend: rewritten `SpeechDetailPage` layout

Top to bottom:

```
← Tillbaka

┌──────────────────────────────────────────────────────────┐
│ [80px avatar]  Magdalena Andersson                        │
│                Socialdemokraterna  · Stockholms län       │
│                ↗ /politicians/:id   ↗ /parties/:party     │
└──────────────────────────────────────────────────────────┘

ANFÖRANDE · TIS 04/09/2025
"Vissa nya regler om arbetsvillkor inom gigekonomin" — topic heading

OM DEBATTEN
[bet] Några tillägg och förtydliganden i den nya lagen om
arbetslöshetsförsäkring  →  Läs hela betänkandet

╔════════════════════════════════════════════╗
║ Speech body (existing dangerouslySetInnerHTML)  ║
║ — full HTML rendered inside .riksdagen-doc      ║
╚════════════════════════════════════════════╝

RESTEN AV DEBATTEN · 12 anföranden
[SpeechRow]
[SpeechRow]
... (current speech excluded; chronological asc)
```

Section breakdown:

1. **Politician card** (top, after back link). 80 × 80 avatar, full
   name (link to `/politicians/:id`), party name (link to
   `/parties/:party`), constituency line, party-coloured background
   strip (re-use the colour treatment from the existing
   `PoliticianPage` hero).
2. **Anförande meta row** (mono kicker): `ANFÖRANDE · WEEKDAY DD/MM/YYYY`.
3. **Topic heading** (serif italic, the existing `topicHeading` field
   if present).
4. **Om debatten card.** Renders only when `useDocument(rel_dok_id)`
   resolves. Shows document type chip + title + link. The link target:
   - `bet` (betänkande) with non-empty beteckning → `/beslut/:beteckning`
   - any other type → no link, just title in muted text.
5. **Speech body** — existing HTML render. Unchanged.
6. **Resten av debatten panel.** Title `RESTEN AV DEBATTEN · N anföranden`.
   `<SpeechRow>` list, ordered by `anforandeNummer` ascending,
   filtered to exclude the current speech's id. If the related
   document has fewer than 2 speeches total, the panel is omitted.

When the speech has no `relatedDokId` (orphan), sections 4 and 6 are
skipped.

### Politician metadata

The current `Speech` payload returned by `/api/speeches/:id` already
has `politicianName`, `politicianImageUrl`, `party`, `politicianId`,
but **does not** include `constituency`. To render
`Magdalena Andersson · Stockholms län`, the frontend needs the
constituency too.

Two options:

- **A.** Extend `politicians.NameByID` in
  `backend/internal/politicians/service.go` to a richer
  `ProfileByID` that returns `{firstName, lastName, party,
  constituency, imageUrl}`. Speeches handler calls it once per
  `toDTO` invocation, populates new DTO fields.
- **B.** Frontend issues a second query: `politiciansApi.getById`,
  which already exists and returns the full profile. Adds 1 RTT but
  no backend changes.

**Decision: B.** Adds zero backend change, leverages an endpoint the
politician page is already battle-testing, and TanStack Query dedupes
the call across components. Trade-off: a 2nd network round-trip. The
detail page already does ~3 (speech, speeches-by-document, document)
so adding a 4th is fine.

## Data flow

1. Visitor opens `/anforanden/974`.
2. Frontend calls `/api/speeches/974` → returns `Speech` with
   `politicianId`, `politicianImageUrl`, `relatedDokId`, etc.
3. Three parallel fetches:
   - `/api/politicians/{politicianId}` → full profile (constituency).
   - `/api/documents/{relatedDokId}` → debate context (title, type,
     beteckning).
   - `/api/speeches/by-document/{relatedDokId}` → list of all
     speeches in this debate.
4. Render: politician card from speech + politician profile, debate
   context from document, speech body from speech, transcript list
   from by-document filtered to exclude current id.

All hooks use `enabled: !!dokId` / `enabled: !!politicianId` guards so
no failed call is made when fields are missing.

## Error handling

- `/api/documents/:dokId` returns 404: frontend hides the "Om
  debatten" section silently. Logging stays in the hook's error
  handler (no toast).
- `/api/speeches/by-document/:dokId` returns empty: hide the "Resten
  av debatten" panel.
- `/api/politicians/:id` 404: render the politician card with name +
  party + initials avatar (no link). The speech itself still shows
  `politicianId`, so links to `/politicians/:id` always work even when
  the profile fetch fails — they'll just land on the politician page's
  own 404 state.

## Testing

- TypeScript clean: `cd frontend && npx tsc --noEmit`.
- Backend build: `go build -C backend ./...`.
- Local smoke (`make run`):
  - `/anforanden/974` shows politician avatar + clickable name + clickable party.
  - "Om debatten" card links to the related betänkande.
  - "Resten av debatten" lists 11 other speakers (12 minus current).
- Mobile prod check (375 × 812): zero horizontal overflow on
  `/anforanden/974`. Politician card stacks correctly. Tap targets
  ≥ 24 × 24.
- Curl smoke: `curl -s 'http://localhost:8080/api/documents/hd01au9' | jq .` returns
  the expected betänkande shape (title + beteckning).

## Open questions for implementation phase

- **`relatedDokId` correctness.** Some speeches in our DB may have an
  empty `related_dok_id`. Verify with a quick `SELECT COUNT(*) FROM
  speeches WHERE related_dok_id IS NULL OR related_dok_id = ''` after
  rolling out — if the count is huge, ingestion needs a fix, but the
  frontend should already degrade to "no debate context".
- **Beteckning extraction.** `dokumentstatus.dokument.beteckning` is
  already parsed by `votes.Client.FetchDocumentStatus`. Verify the
  field maps cleanly through to the new endpoint response — if not,
  add it.
- **Reverse navigation.** A speech under an interpellation does not
  link anywhere from the "Om debatten" card. If interpellations get a
  detail page later, swap that empty link target for `/ip/:id` or
  similar.
