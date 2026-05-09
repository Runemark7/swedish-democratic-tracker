# Debate Speakers + Party Detail Page

**Date:** 2026-05-09
**Status:** Design (approved by user, pending implementation plan)

## Context

`/beslut/:beteckning` already inlines the full Riksdagen betänkande HTML
(commit `812fd2b`). What it does not surface is **who said what** during
the debate — the most human, citizen-relevant slice of any Riksdagen
decision. Speeches exist in the database (`speeches` table, populated
by the daily worker; full prose backfilled by the new
`enrich-speech-text` worker). They are simply not connected to the
beslut page.

Two further connections are missing:

1. **Politician → speeches.** A visitor reading
   `/politicians/:id` cannot see what that politician has actually
   said in the chamber.
2. **Party → speeches.** No `/parties/:party` page exists at all
   today; party-level navigation goes straight to
   `/parties/:party/goals`. The user wants an aggregate "what active
   politicians in this party are saying" view.

Goal: tie the existing speech corpus to three pages so a citizen can
trace `beslut → speaker → other things they have said → other people
in their party`.

## Out of scope

- AI summarisation. The spec mentions it as a hybrid fallback for
  Riksdagen's `summary` field but **the AI path is deferred**. v1
  shows Riksdagen's text when present and an empty-state line when
  not.
- New party-level scorecards / aggregate alignment metrics. The
  party detail page reuses the existing `/api/parties` data.
- Editorial highlighting of "key quotes". v1 shows full snippets, no
  AI extraction.
- Cross-debate threading (e.g. "this politician said the opposite
  three weeks ago"). Future enhancement.

## Architecture

Three layers added on top of existing infrastructure:

1. **Backend endpoints** (3 new routes on the speeches HTTP handler):
   - `GET /api/speeches/by-document/:dokId`
   - `GET /api/speeches/by-politician/:intressentId?limit=20`
   - `GET /api/speeches/by-party/:party?limit=50`
2. **Frontend hooks** that wrap each endpoint via `useQuery`.
3. **UI surfaces** on three pages:
   - `BeslutDetailPage` — adds a `SAMMANFATTNING` block + `DEBATTEN`
     speaker list above the existing `HELA BETÄNKANDET` HTML render.
   - `PoliticianPage` — adds an `ANFÖRANDEN` card listing the
     politician's 10 most recent anföranden.
   - **New** `PartyDetailPage` at `/parties/:party` — hero +
     three-tab interface (Mål · Anföranden · Politiker).

The speakers row design (avatar + name + party chip + snippet + link)
is reused as a single shared `SpeechRow` component so all three pages
look identical down to the row. This keeps the DRY pressure off the
three callers.

## Components

### `SpeechRow` (new, shared)

`frontend/src/features/speeches/SpeechRow.tsx`. Visual:

```
┌──────────────────────────────────────┐
│ [40px avatar] Eva Lindh   (S)  18/04 │
│ "Herr talman! Den här frågan..."     │
│ Läs hela →                           │
└──────────────────────────────────────┘
```

- Avatar: 40 × 40 circular `<img>` from `politician.imageUrl`. If
  missing, render initials with the party-coloured background — same
  fallback already in `PoliticianCard`.
- Name: serif 14, links to `/politicians/:intressentId`.
- Party chip: tiny pill in `PARTY_COLORS[party]`.
- Date: mono 10 muted (`DD/MM`).
- Snippet: body 13, two-line clamp via `-webkit-line-clamp: 2`.
- Footer link: mono 11 accent, `Läs hela →`, links to
  `/anforanden/:id`.

Component takes a `Speech` plus an optional avatar URL (politician
endpoint must include `imageUrl` — see Backend section below).

### `BeslutDetailPage` additions

Order on the page becomes:

1. Hero (existing)
2. Outcome strip (existing)
3. Riksdagen summary block — existing `data.summary` rendered, or
   muted line `Riksdagen har inte publicerat någon sammanfattning för
   det här beslutet.` when empty.
4. Party-vote breakdown (existing)
5. **NEW** `DEBATTEN — N TALARE` section: list of `SpeechRow`
   components ordered chronologically (`anforande_nummer` ascending).
   First 5 visible by default; `Visa alla N anföranden` button reveals
   the rest in-place. Empty state line: `Inga registrerade anföranden
   för det här beslutet ännu.`
6. Riksdagen link (existing, repositioned to live below the speakers)
7. **Existing** `HELA BETÄNKANDET` HTML body

Data: `useSpeechesByDocument(dokId)`. Calls
`/api/speeches/by-document/:dokId`.

### `PoliticianPage` additions

New card after the existing voting record, matching existing
`.sdt-card` chrome:

```
ANFÖRANDEN · Senaste 10 inlägg i kammaren
─────────────
<SpeechRow> × 10
─────────────
Visa alla anföranden →
```

Footer link routes to a new sub-page or simply scrolls to a longer
list — for v1, just append a `?showAll=1` query param the same page
reads to render up to 50 instead of 10.

Data: `useSpeechesByPolitician(intressentId, limit)`.

### `PartyDetailPage` (new)

Route `/parties/:party`. File
`frontend/src/features/parties/PartyDetailPage.tsx`.

Layout:

```
┌──────────────────────────────────────┐
│ [colour bar] S  Socialdemokraterna   │
│ 107 mandat · oppositionsledare       │
└──────────────────────────────────────┘

[ Mål ]  [ Anföranden ]  [ Politiker ]   ← tab strip

(active tab content fills the rest of the page)
```

- **Mål tab** — reuses the existing list rendered today on
  `/parties/:party/goals` (call `useQuery(["party-goals", party])`
  identical to `PartyGoalsPage`).
- **Anföranden tab** — list of `SpeechRow` for the party, max 50,
  newest first. Backed by `useSpeechesByParty(party, 50)`. Empty
  state: `Inga registrerade anföranden från ${partyName} ännu.`
- **Politiker tab** — reuses the existing `PoliticianCard` from
  `PoliticiansPage` (extract into a shared component if it isn't
  already), filtered to active members of this party. Backed by
  `useQuery(["politicians", { party }])` against the existing
  `/api/politicians?party=X` endpoint.

Tab state lives in the URL via `?tab=mal|anforanden|politiker`
(default `mal`). Mobile: tabs scroll horizontally, no other layout
change.

A small `← Alla partier` link at the top routes back to `/parties`.

The existing `/parties/:party/goals` route stays alive — it just
becomes one of the things `/parties/:party` links to internally. We
do not redirect.

### Backend endpoints

All three live on the existing speeches handler at
`backend/internal/speeches/adapters/http/handler.go`. Each follows
the shape of the existing `/speeches/recent`:

```
GET /speeches/by-document/:dokId
GET /speeches/by-politician/:intressentId?limit=20
GET /speeches/by-party/:party?limit=50
```

Each returns `[]SpeechDTO` (the existing DTO; `SpeechText` omitted
unless full-detail endpoint). The DTO needs **one new field**:

```go
PoliticianImageURL string `json:"politicianImageUrl,omitempty"`
```

The `PoliticianLookup` interface (defined in the speeches handler)
gains a method to return the image URL alongside the name:

```go
type PoliticianLookup interface {
    NameByID(ctx context.Context, intressentID string) (string, error)
    ImageURLByID(ctx context.Context, intressentID string) (string, error)
}
```

Implemented in `politicians.Service` against the existing
`Politician.ImageURL` column.

Repository methods on `SpeechRepository`:

```go
ListByDocument(ctx context.Context, dokID string) ([]*domain.Speech, error)
ListByParty(ctx context.Context, party string, limit int) ([]*domain.Speech, error)
```

`ListByPolitician` already exists; only the HTTP route is new.

Postgres queries:

- `WHERE rel_dok_id = $1 ORDER BY anforande_nummer ASC` for `ListByDocument`.
- `WHERE party = $1 ORDER BY date DESC, id DESC LIMIT $2` for `ListByParty`.
- `WHERE politician_id = $1 ORDER BY date DESC, id DESC LIMIT $2` for the existing list method (already implemented; just exposed via HTTP).

### OpenAPI

`api/openapi.yaml` gains the three new paths. Schema reuses the
existing `Speech` component plus an optional `politicianImageUrl`
property. After updating, regenerate
`frontend/src/shared/api-contract.ts` via `npm run generate:api`.

## Data flow

1. User opens `/beslut/AU9`. Frontend calls existing
   `/api/votes/AU9/1` → gets `dokId = "hd01au9"`. New call
   `/api/speeches/by-document/hd01au9` → list of speeches.
2. Each `SpeechRow` renders directly from the response.
3. Click politician name → routes to `/politicians/:intressentId`,
   which calls `/api/speeches/by-politician/:id?limit=10` to populate
   the new card.
4. Click party chip → routes to `/parties/:party`, default Mål tab.
   Switch to Anföranden tab → calls `/api/speeches/by-party/:party`.
5. Click `Läs hela →` on any row → existing `/anforanden/:id` route.

All speech endpoints are GET, idempotent, cached by TanStack Query
with 60-second `staleTime` (matches existing speech queries).

## Error handling

- 404 from `/speeches/by-document/:dokId` (no speeches found):
  return empty array `[]`, frontend renders the empty-state line.
- Network failure: each panel shows a small italic muted line
  `Kunde inte hämta anföranden.` Other panels on the page keep
  rendering (per existing pattern).
- Politician without image: `imageUrl == ""` triggers the initials
  fallback in `SpeechRow`, no error state.

## Testing

- **TypeScript:** `cd frontend && npx tsc --noEmit` clean.
- **Backend:** `go build -C backend ./...` clean.
- **Curl smoke:** all three endpoints return JSON arrays for valid
  inputs, empty arrays for unknown ids.
- **Visual smoke (375 × 812 mobile):**
  - `/beslut/AU9` shows the speakers section above HELA BETÄNKANDET.
  - `/politicians/:id` shows the new ANFÖRANDEN card.
  - `/parties/S` renders hero + tab strip; tab switch is keyboard
    + tap-friendly; tab state survives a refresh (URL query param).
- **No horizontal overflow** on any of the three pages at 375 × 812.

## Open questions for implementation phase

- **`rel_dok_id` accuracy.** Earlier testing showed
  `?rel_dok_id=hd01au3` returned interpellation speeches mixed in
  with the betänkande debate. If the speakers list shows obviously
  off-topic speeches for a given beslut, add a filter on
  `kammaraktivitet` (e.g. only `Beslut`/`Debatt`) — defer the call
  until we see real data on the AU9-class examples.
- **Image URL freshness.** `Politician.ImageURL` comes from
  Riksdagen's static image CDN. Some URLs return 404 in practice; the
  `<img onerror>` fallback should swap to initials silently. Verify
  during implementation.
- **Goals query reuse.** Confirm the `useQuery` for
  `/api/parties/:party/goals` can be called from both
  `PartyGoalsPage` and the new `PartyDetailPage` without duplicate
  fetches (TanStack Query dedupes by `queryKey`, so this should
  Just Work). Don't break the existing standalone goals page.
- **Tab styling.** The site has no existing tab pattern. Either
  build a small `<TabStrip>` shared component (mono pills with
  bottom-border indicator) or inline the tab markup in
  `PartyDetailPage` for v1. Lean toward the inline approach — extract
  only when a second tabbed page appears.
