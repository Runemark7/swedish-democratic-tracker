# Home Page — Riksdag Activity Overview

**Date:** 2026-05-05
**Status:** Design (approved by user, pending implementation plan)

## Context

The current home page (`frontend/src/features/home/HomePage.tsx`) presents a
three-chamber identity (Riksdag · Region · Kommun) via a `ChambersHero`
nested-box visual, a 3-column `StatusCell` strip, and a 24-hour `PulseStrip`.
A first-time visitor sees no value proposition, no active content, and no
call to action. Region/Kommun cells are particularly weak because the site
cannot geolocate the visitor — generic regional/municipal data has low
signal for an unknown user.

Goal: replace the home page so a visitor immediately understands **what the
Riksdag is currently working on** — beslut, debatter, and what the parties
are saying. National data has uniform signal for every Swedish voter, so
the home page focuses on Riksdag activity. Region/Kommun remain reachable
via the main nav but no longer surface on `/`.

The deeper editorial intent (per user): in a representative democracy a
voter delegates power to a party expecting it to act on its stated goals.
The home page does not score "say-vs-do" populism directly, but by
surfacing what parties **say** (anföranden) next to what they **vote for**
(beslut, omröstningar) it gives the visitor the raw material to judge for
themselves.

## Out of scope

- Region/Kommun content on `/`
- A computed populism / say-vs-do score (deferred — visitor judges from raw activity)
- Per-visitor personalisation (no geolocation, no logged-in users)
- Mobile bottom tab bar or other navigation overhaul
- New theme tokens or palette changes

## Page structure

```
┌─────────────────────────────────────────────────────────────┐
│ HERO                                                         │
│ KAMMARE ETT · RIKSDAGEN                                      │
│ Vad fokuserar riksdagen på?                                 │
│ Beslut, debatter och vad partierna driver — i realtid.      │
├──────────────────────────────┬──────────────────────────────┤
│ 1. Beslut idag               │ 2. Aktuella debatter         │
├──────────────────────────────┼──────────────────────────────┤
│ 3. Veckans omröstningar      │ 5. Kommande beslut           │
├──────────────────────────────┴──────────────────────────────┤
│ 4. Vad partierna säger  (full width, 8 party cards)         │
└─────────────────────────────────────────────────────────────┘
```

Mobile (≤640 px): all panels collapse to a single column in the order 1, 2,
3, 5, 4. Panel 4 stays last because the 8-card grid is the heaviest section.

### Hero

Tight, no decorative chart or counter. Three text rows:

- **Kicker** — `KAMMARE ETT · RIKSDAGEN` — `var(--font-mono)`, 10 px, letter-spacing 2 px, `var(--color-fg-muted)`.
- **Headline** — `Vad fokuserar riksdagen på?` — `var(--font-serif)`, 44 px desktop / 28 px mobile, `var(--color-fg)`.
- **Subtitle** — `Beslut, debatter och vad partierna driver — i realtid.` — `var(--font-body)`, 13 px, `var(--color-fg-muted)`.

Padding: `40px 32px 24px` desktop / `20px 14px 16px` mobile.

### Panel chrome (shared)

- 1 px `var(--color-border)` outer border, no border-radius (matches existing chamber pages).
- Mono kicker header inside the panel: `BESLUT IDAG`, `AKTUELLA DEBATTER`, etc — same style as kickers on Riksdag/Region/Kommun pages.
- Up to 5 items visible. If `data.length > 5`, show muted "Visa alla →" link as last row routing to the relevant index page.
- Empty state: italic muted line, e.g. `Inga beslut idag — kammaren har sommaruppehåll.` Use the same line height as a normal item so panel heights stay similar.

### Panel 1 — Beslut idag

Source: existing `useRiksdag().liveVotes` filtered to `time.startsWith("Idag")` (existing convention in mock + feed).

Item row shape (matches Riksdag page Puls):

```
● HH:MM   Title (ellipsis)        [Pill: Bifall/Avslag]   margin
```

Click → `/votes/{beteckning}/{punkt}` (existing route, link via existing `beslutHref` helper).

### Panel 2 — Aktuella debatter

Source: speeches feature already runs a daily worker (`backend/internal/speeches`). The HTTP adapter currently exposes politician-scoped speech queries, **but no global "recent speeches" endpoint exists**. Plan to add:

```
GET /api/speeches/recent?limit=20
→ [{ id, politician_id, politician_name, party, topic, snippet, occurred_at }]
```

Item row shape:

```
HH:MM   Politician name           [party color dot] Topic snippet
```

Click → speech detail page (new route `/anforanden/:id` — minimal stub page in this plan; full detail UI is a separate spec).

### Panel 3 — Veckans omröstningar

Source: votes resolved in the last 7 days plus any with a known scheduled date in the next 7 days. Backend currently exposes the riksdag feed used for `liveVotes` but has no scheduled-vote endpoint. For v1, **show the last 7 days of resolved votes only** and accept the "kommande" half as a future enhancement noted in the spec. Header label stays `VECKANS OMRÖSTNINGAR`.

Item row shape:

```
DD/MM   Title (ellipsis)          [Pill: Bifall/Avslag]   margin
```

Click → `/votes/{beteckning}/{punkt}`.

### Panel 4 — Vad partierna säger

Eight party cards in a 4×2 grid on desktop, 2×4 grid on mobile. One card per riksdagsparti (S, M, SD, C, V, KD, L, MP) — fixed order matching Hemicycle convention.

Card layout:

```
┌────────────────────────┐
│ ▌ S  Socialdemokraterna │   ← party color strip + short + name (mono)
│                         │
│ Magdalena Andersson     │   ← politician (serif 14)
│ "Vi måste prioritera... │   ← speech excerpt (1-2 lines, body 12, italic)
│ TIS 13:42 · Om vården   │   ← time + topic (mono 9 muted)
└────────────────────────┘
```

Source: derive client-side from a larger fetch via the new `/api/speeches/recent?limit=100` endpoint — group by `party` and pick the newest per party. Avoids a second backend route. The 100-row payload is small (~30 KB) and is fetched once, shared with Panel 2 (which uses the top 5).

If a party has no recent speech (rare): show muted card with party name + `Inget anförande senaste 30 dagarna`.

Card click → speech detail page.

### Panel 5 — Kommande beslut

Source: existing `useRiksdag().agenda` array — already used on `/riksdag` as the AGENDA card. Reuses the existing `<AgendaList>` component from `frontend/src/components/AgendaList.tsx` to avoid divergent agenda visuals across the site. Item shape (per `<AgendaList>`):

```
DD/MM   [Committee badge]   Title (ellipsis)
```

Click → `/votes` (the index page). Per-agenda-item detail is out of scope for v1.

## Component changes

| File | Change |
|---|---|
| `frontend/src/features/home/HomePage.tsx` | Rewrite. Drop `ChambersHero`, `StatusCell`, `PulseStrip`. Add new hero + 5 panel sections. Reuse `Pill`, `AgendaList`, `useMediaQuery`, `useRiksdag`. |
| `frontend/src/features/home/components/PanelCard.tsx` | New. Shared panel chrome (header + items + "Visa alla" footer + empty state). |
| `frontend/src/features/home/components/PartySpeechCard.tsx` | New. Card used by Panel 4. |
| `frontend/src/hooks/useDemocracy.ts` | Add `useRecentSpeeches(limit)` hook backed by new endpoint. |
| `backend/internal/speeches/adapters/http/handler.go` | Add `GET /api/speeches/recent?limit=N` route. Existing repository already stores speeches; add list-all-recent query. |
| `backend/internal/speeches/service.go` | Expose `ListRecent(ctx, limit)`. |
| `backend/internal/speeches/ports/repository.go` | Add `ListRecent(ctx, limit) ([]Speech, error)` interface method. |
| `backend/internal/speeches/adapters/postgres/repository.go` | Implement `ListRecent` with `ORDER BY occurred_at DESC LIMIT $1`. |
| `api/openapi.yaml` | Add `/speeches/recent` schema, regenerate `frontend/src/shared/api-contract.ts`. |
| `frontend/src/App.tsx` | Add route `/anforanden/:id` → minimal `SpeechDetailPage` stub. |
| `frontend/src/features/speeches/SpeechDetailPage.tsx` | New stub: shows full speech text + back link. Sufficient for v1. |

`mockRegion` / `mockKommun` imports get removed from `HomePage.tsx`. Their definitions stay in `frontend/src/mock/democracy.ts` for now (used elsewhere) — orphan removal is a separate cleanup.

## Data flow

1. Visitor lands on `/`.
2. `HomePage` calls `useRiksdag()` (existing; provides `liveVotes` + `agenda` for panels 1, 3, 5).
3. `HomePage` calls `useRecentSpeeches(20)` (new).
4. From the speech list, derive:
   - Panel 2 items: top 5 by `occurred_at` descending.
   - Panel 4 items: group by `party`, pick newest per party, render in fixed party order.
5. Each panel renders via `<PanelCard>` with its own item renderer.

All panels degrade gracefully on empty data (existing `EMPTY_BUDGET`-style pattern).

## Error handling

- Network failure on `useRiksdag` → existing skeleton already in `HomePage` covers this.
- Network failure on `useRecentSpeeches` → panels 2 and 4 show their empty-state copy. Panels 1, 3, 5 still render from `useRiksdag`.
- Per-panel errors do not block the page; each panel handles its own loading + error state.

## Mobile considerations

- Single column.
- Panel order on mobile: 1, 2, 3, 5, 4 (Panel 4 last because it's heaviest at 8 cards).
- Panel padding: `16` mobile / `24` desktop.
- Item rows truncate titles with `text-overflow: ellipsis`.
- Hero kicker stays 10 px on mobile (already small enough).

## Testing

- TypeScript: `cd frontend && npx tsc --noEmit` passes.
- Backend: `go build -C backend ./...` passes.
- Visual mobile (375 × 812 prod with new build): no horizontal overflow on `/`.
- Visual desktop (1280 × 800): hero + 4 panel quads visible above fold.
- Click each panel item type once and confirm route lands correctly.
- Empty state: temporarily mock empty `liveVotes` and `recentSpeeches` to verify the empty copy renders.

## Open questions for implementation phase

- Speech detail page (`/anforanden/:id`) is a stub here. A richer design lives in a separate future spec. v1 should be functional but minimal.
- Panel 3 "planerad" votes deferred until backend exposes scheduled-vote dates. Header stays `VECKANS OMRÖSTNINGAR`; for v1 it shows the last 7 days only.
