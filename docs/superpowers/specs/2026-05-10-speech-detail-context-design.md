# Speech Detail + Politician Profile Tracking

**Date:** 2026-05-10
**Status:** Design (approved by user, pending implementation plan)

## Context

Two pages need to evolve so a citizen can move from a single speech
into deep tracking of the politician who delivered it.

**`/anforanden/:id` — speech detail.** Shows the speech body, the
speaker's name as plain text, the party as plain text, and the date.
Three problems:
1. **No navigation.** Politician name and party are not links.
2. **No image.** The site already stores `politician.imageUrl` and
   uses it everywhere else.
3. **No debate context.** A speech is part of a debate — users cannot
   see who answered, who spoke before, or what the debate was about
   beyond the cryptic `topic_heading`.

**`/politicians/:id` — politician profile.** Today shows the hero,
plus three tabs (Röstningshistorik · Löften · Anföranden). The user
explicitly wants more: ability to **track** a politician — see what
debates they are active in, what topics recur in their work, and what
they are doing right now in present tense. The current Anföranden tab
is a flat list, which doesn't aggregate anything.

Goal: turn `/anforanden/:id` from a leaf page into a hub that pivots
into a deeper politician profile, and rework `/politicians/:id` so the
profile actually surfaces patterns of activity rather than just lists.

## Out of scope

- Notifications / "follow this politician" subscriptions. The
  user-facing word "track" is interpreted as observation via UI
  affordances, not push notifications.
- AI-generated topic clustering. Topic frequency uses raw
  `topic_heading` string deduplication; smarter NLP is deferred.
- Per-anförande replies/interactions (Riksdagen does not return a
  structured reply tree; chronological order is best-effort).
- Inline expand/collapse of other speeches' full text on the speech
  detail page — clicking a related speech navigates to its own detail
  page, same as today.
- Standalone routes for interpellations, motions, propositions. The
  "Om debatten" card on the speech detail page links to
  `/beslut/:beteckning` only when the related document is a betänkande;
  for other doc types it shows the title without a link.
- Politician page hero changes. Only the tab set below the hero
  changes.

## Architecture

One new backend endpoint plus reworked frontend on two pages:

1. **Backend** — `GET /api/documents/:dokId` returning the parsed
   `DocumentStatus` (title, type, beteckning, summary, session, date,
   without bodyHtml). Implemented as a thin chi route on the votes
   feature delegating to the existing `votes.Service.GetDocumentStatus`,
   which already wraps the Riksdagen `dokumentstatus` API. No new
   business logic.
2. **Speech detail page** — politician card (image + clickable name +
   clickable party), "Om debatten" card pointing at the related
   document, the existing speech body, and a "Resten av debatten"
   panel listing all other speeches in the same debate via
   `<SpeechRow>`.
3. **Politician profile page** — replace the existing
   Röstningshistorik · Löften · Anföranden tab set with **Aktivitet ·
   Debatter · Ämnen · Löften**. Aktivitet is a combined chronological
   feed of votes + speeches; Debatter groups speeches by debate;
   Ämnen aggregates topic_heading frequency; Löften stays.

Reuses everything already shipped:
- `<SpeechRow>` for the transcript list and any speech-row needs.
- `useSpeechesByDocument(dokId)` hook.
- `useSpeechesByPolitician(id, limit)` hook.
- `politician.imageUrl` + initials fallback.
- `votes.Service.GetDocumentStatus`.
- The existing `politiciansApi.listVotes(id, ...)` — already used by
  the current Röstningshistorik tab.

## Components

### Backend: `GET /api/documents/:dokId`

Mounted in `backend/internal/votes/adapters/http/handler.go` (the votes
feature already owns dokumentstatus). New route line in `Routes`:

```go
r.Get("/documents/{dokId}", h.getDocument)
```

Handler returns:

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
the `beteckning` field. Frontend uses that to build a
`/beslut/:beteckning` link.

`bodyHtml` is **not** included — only the beslut detail page needs it,
and including it here would 10x the payload.

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

### Speech detail page (`/anforanden/:id`)

Top-to-bottom layout:

```
← Tillbaka

┌──────────────────────────────────────────────────────────┐
│ [80px avatar]  Magdalena Andersson                        │
│                Socialdemokraterna  · Stockholms län       │
│                ↗ /politicians/:id   ↗ /parties/:party     │
└──────────────────────────────────────────────────────────┘

ANFÖRANDE · TIS 04/09/2025
"<topic_heading>"

OM DEBATTEN
[bet] <document title>  →  Läs hela betänkandet (when type=bet)

╔══════════════════════════════════════════════════╗
║ Speech body (existing dangerouslySetInnerHTML)    ║
║ — full HTML rendered inside .riksdagen-doc        ║
╚══════════════════════════════════════════════════╝

RESTEN AV DEBATTEN · N anföranden
<SpeechRow> × N (current excluded, chronological asc)
```

1. **Politician card.** 80 × 80 avatar (or initials with party
   colour), full name (`Link` to `/politicians/:id`), party name
   (`Link` to `/parties/:party`), constituency line. Use the same
   colour treatment as the existing `PoliticianPage` hero.
2. **Anförande meta row** (mono kicker): `ANFÖRANDE · WEEKDAY DD/MM/YYYY`.
3. **Topic heading** (serif italic, the existing `topicHeading` field
   if present).
4. **Om debatten card.** Renders only when `useDocument(rel_dok_id)`
   resolves. Document type chip + title + link. Link target:
   - `bet` with non-empty beteckning → `/beslut/:beteckning`
   - any other type → no link, just title in muted text.
5. **Speech body** — existing HTML render. Unchanged.
6. **Resten av debatten panel.** Title `RESTEN AV DEBATTEN · N anföranden`.
   `<SpeechRow>` list, ordered by `anforandeNummer` ascending, current
   speech filtered out. If the related document has fewer than 2
   speeches total, the panel is omitted.

When the speech has no `relatedDokId` (orphan), sections 4 and 6 are
skipped.

The frontend issues `politiciansApi.getById(politicianId)` to get the
speaker's `constituency` field — the speech payload doesn't include it
and a second query is the smaller change.

### Politician profile page (`/politicians/:id`) — tab rework

Hero stays as-is. The tab strip below changes from
`Röstningshistorik · Löften · Anföranden` to **`Aktivitet · Debatter ·
Ämnen · Löften`**.

#### Aktivitet (default tab)

Combined timeline of votes + speeches, newest first, cap 50.

Composed client-side by merging the existing `politiciansApi.listVotes`
and `speechesApi.listByPolitician` responses, normalising each to:

```ts
type ActivityEvent =
  | { kind: "vote";    date: string; vote: Vote }
  | { kind: "speech";  date: string; speech: Speech };
```

Sort by `date` desc, slice to 50. Render rows of two shapes:
- Vote row: same look as the current Röstningshistorik table — vote
  result chip + title + beteckning + committee.
- Speech row: existing `<SpeechRow>` with `hidePolitician`.

A small mono kicker row above ("Senaste 50 händelser") makes the
chronological framing explicit.

Empty state: italic muted line `Ingen registrerad aktivitet ännu.`

#### Debatter

Speeches grouped by `relatedDokId`. Cap top 20 debates by speech count
(this politician's count). For each debate:

```
[bet] <document title>             5 anföranden · 2025-09-04 → 2025-10-12
                                                         →
```

- Document title comes from `useDocument(relatedDokId)` per row. Each
  row renders an outer `<Link>` to `/beslut/:beteckning` when the doc
  type is `bet`; otherwise the row is inert (no Link wrapper, just a
  div with the title).
- Speech count is `speeches.filter(s => s.relatedDokId === id).length`.
- Date range: min/max `date` across the matching speeches.

Empty state: `Inga debatter har kopplats till anförandena ännu.`

#### Ämnen

Top 10 `topicHeading` strings by frequency. Each row:

```
Svar på interpellation 2024/25:734 …  · 4 anföranden · senast 2025-10-12
```

Pure client-side aggregation off the `useSpeechesByPolitician(id, 200)`
result. Map of `topicHeading → { count, lastDate }`, sort by count
desc, slice 10. Skip rows with empty topicHeading.

Empty state: `Inga teman går att utläsa ännu.`

#### Löften

Unchanged from the existing Löften tab. Move the JSX into the new tab
without modifying its internals.

#### Removed

- `Röstningshistorik` standalone tab — votes are now part of Aktivitet.
- `Anföranden` standalone tab — speeches are also part of Aktivitet,
  and grouped in Debatter, and aggregated in Ämnen. The flat list
  was redundant.

#### URL-driven tab state

Read tab from `?tab=...` (mirroring the pattern from
`PartyDetailPage`). Default is `aktivitet`. Acceptable values:
`aktivitet`, `debatter`, `amnen`, `loften`. Unknown values fall back
to `aktivitet`.

## Data flow

**Speech detail page** (`/anforanden/974`):
1. `useQuery(["speech", 974])` → speech with `politicianId`,
   `relatedDokId`, `politicianImageUrl`.
2. Three parallel queries:
   - `politiciansApi.getById(politicianId)` → constituency.
   - `useDocument(relatedDokId)` → debate context (title, type,
     beteckning).
   - `useSpeechesByDocument(relatedDokId)` → list of all speeches in
     this debate.
3. Render: politician card from speech + politician profile, debate
   context from document, speech body from speech, transcript list
   from by-document filtered to exclude current id.

**Politician page** (`/politicians/:id`, default tab `aktivitet`):
1. Hero data: `politiciansApi.getById(id)` → existing.
2. `politiciansApi.listVotes(id, { pageSize: 100 })` → existing.
3. `speechesApi.listByPolitician(id, 200)` → use the existing hook
   `useSpeechesByPolitician(id, 200)` (raise the limit from the
   default 20 so Debatter and Ämnen aggregations are meaningful).
4. Switching to `Debatter`: extract unique `relatedDokId` values from
   the speech list, top 20 by count. For each, `useDocument(dokId)` —
   TanStack Query dedupes if multiple debates share IDs.

All hooks gate with `enabled: !!id` so no failed call when fields are
missing.

## Error handling

- `/api/documents/:dokId` 404: speech detail hides "Om debatten"
  silently. Politician Debatter tab shows the row with the raw
  `relatedDokId` as a fallback title.
- `/api/speeches/by-document/:dokId` empty: hide "Resten av debatten".
- `politiciansApi.getById(id)` 404: render politician card without
  link decoration (initials avatar, plain text name + party).
- `politiciansApi.listVotes(id)` empty: Aktivitet shows speech-only
  events; Debatter and Ämnen still work.
- `speechesApi.listByPolitician(id)` empty: Aktivitet shows vote-only
  events; Debatter/Ämnen show their empty-state copy.
- A politician with no votes AND no speeches: Aktivitet shows the
  italic "Ingen registrerad aktivitet ännu." line; Debatter and Ämnen
  show their respective empty states.

## Testing

- TypeScript clean: `cd frontend && npx tsc --noEmit`.
- Backend build: `go build -C backend ./...`.
- Local smoke (`make run`):
  - `/anforanden/974` — politician card with avatar, clickable name,
    clickable party, "Om debatten" linking to `/beslut/AU9`,
    "Resten av debatten" listing other speakers.
  - `/politicians/{some-id}` — 4 tabs visible; default Aktivitet renders
    a chronological mix of votes + speeches; Debatter shows top
    debates by speech count; Ämnen shows top topic strings.
- Mobile prod check (375 × 812): zero horizontal overflow on both
  pages. Politician card on speech detail stacks correctly. Tabs
  scroll horizontally if they overflow.
- URL state: `/politicians/:id?tab=debatter` directly opens the
  Debatter tab; refresh keeps it open.

## Open questions for implementation phase

- **`relatedDokId` correctness.** Some speeches in our DB may have an
  empty `related_dok_id`. Verify with a quick `SELECT COUNT(*) FROM
  speeches WHERE related_dok_id IS NULL OR related_dok_id = ''` after
  rolling out — if the count is huge, ingestion needs a fix, but the
  frontend should already degrade to "no debate context".
- **Beteckning extraction.** `dokumentstatus.dokument.beteckning` is
  already parsed by `votes.Client.FetchDocumentStatus`. Verify the
  field maps cleanly through to the new endpoint response.
- **Debatter row count math.** If `useSpeechesByPolitician(id, 200)`
  caps at 200, a politician with 250 speeches will under-count older
  debates. For v1 accept the cap; raising it is a one-line change.
- **Topic noise.** `topic_heading` strings are long and verbose.
  Aggregating raw strings will produce a top-10 list dominated by full
  interpellation titles like "Svar på interpellation 2024/25:734 om
  arbetsvillkor inom gigekonomin". v1 accepts the noise; future work
  may strip the "Svar på interpellation N:M om" prefix or do real NLP.
- **Reverse navigation from interpellations.** A speech under an
  interpellation does not link anywhere from the "Om debatten" card.
  If interpellations get a detail page later, swap that empty link
  target.
