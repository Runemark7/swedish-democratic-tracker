# Utskott Spine — Phase 3: Information Architecture Switch Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the committee spine the front door — a complete dated feed above a committee index above a party index — restructure navigation into one shared definition with a Fördjupning tier, and retire the surfaces the map ruled out.

**Architecture:** One new backend endpoint (a complete betänkande feed, no committee filter) reusing the existing `/dokumentlista` client. Front page reassembled from three blocks. Navigation collapses from four duplicated lists into one exported constant. The agenda becomes a document index under `/regering`; `/sok` and `/agenda/:id` go; the invented committee→region feed goes.

**Tech Stack:** Go 1.26.1, chi v5, React 19, Vite 6, React Router v7, TanStack Query v5, Tailwind v4.

**Depends on:** Phase 1 (`/api/committees`) and Phase 2 (`/committees/:code` page, shared `GoalCard`). Do not start Phase 3 until a committee page exists — this phase points the front door at it.

## Global Constraints

- **Go 1.26.1, Node 25.9.0.** Parameterised SQL only. `cmd/api/main.go` is the only wiring point.
- **`api/openapi.yaml` first**, then `npm run generate:api`. `frontend/src/shared/types.ts` hand-mirrors it and must be updated too.
- **No ordering that implies precedence.** Alphabetical unless the data has an intrinsic order (a feed is chronological; that is intrinsic).
- **No route deleted other than `/sok` and `/agenda/:id`.** Everything else survives the IA change.
- **Verification gate:** `go build -C backend ./...`, `npx tsc --noEmit`, `npm run lint`, `npm run build` all clean.
- **No frontend component test runner exists.** Verification is `tsc` + `lint` + `build` + a scripted browser check against a real API.
- **Copy is Swedish; code, comments, commits and docs are English.**

## Decisions this plan implements

| Source | Decision |
|---|---|
| [#88](https://github.com/Runemark7/swedish-democratic-tracker/issues/88) | Front page carries **both**: a complete dated feed of the most recent sitting period **above** the committee index. The unit is the **betänkande**, not the förslagspunkt. Each item names and links its utskott. The period is the selection; the calendar defines periods. |
| [#98](https://github.com/Runemark7/swedish-democratic-tracker/issues/98) | **State the facts, infer nothing.** The feed names the dates it covers and the site states there is nothing newer. It never claims recess — Riksdagen publishes no sourceable sitting calendar. |
| [#100](https://github.com/Runemark7/swedish-democratic-tracker/issues/100) | **Party index is a front-page peer** to the committee index. Eight badges, alphabetical. |
| [#101](https://github.com/Runemark7/swedish-democratic-tracker/issues/101) | Fördjupning is a **persistent second tier in the top nav**, `Kommun · Region · Regering · Riksdag`, **numerals dropped**. Nav **defined once**. `/votes` and `/budget` are both committee-scoped links and unscoped destinations, with the `origin_enriched` gate stated in plain Swedish. `/politicians` entered from a ballot. |
| [#102](https://github.com/Runemark7/swedish-democratic-tracker/issues/102) | Agenda becomes a **hand-kept index of whole documents** under `/regering` — no authored description, no per-item status, no point extraction. Off the front page. `/agenda/:id` retires. |
| [#103](https://github.com/Runemark7/swedish-democratic-tracker/issues/103) | **Strictly national.** One sourced sentence on the Region and Kommun landing pages about who decides what. |
| Spec | `/sok` is removed: `SearchPage.tsx` is a client-side substring filter over one preloaded list of votes with no backend search endpoint. |

## A live contradiction this phase must fix

`GetRiksdagFeed` (`backend/internal/votes/service.go:170-184`) hardcodes which committees matter to a level — `region → SoU, TU`; `kommun → UbU, CU, SoU` — and `useRegion` calls it (`frontend/src/hooks/useDemocracy.ts:333`), as does `useMunicipality` (`:483`). So **every region and kommun page already renders riksdag betänkanden from a committee set we picked**, presented as relevant to that place.

That is precisely the invented committee→regional relevance mapping #103 ruled does not honestly exist, and it is in production. Task 5 removes it.

## Carried in from Phase 2: two grains for one record, on linked pages

Phase 2 left a divergence this phase has to decide on. It is recorded here because
Phase 2 deliberately did not resolve it, and the numbers look like a bug to anyone
who meets them cold.

- `/parties/:party` publishes **2 558 av 2 562 omröstningar**. Its coverage query counts
  `DISTINCT beteckning||':'||forslagspunkt` per session.
- `/committees/:code` headers count `DISTINCT votering_id` and sum to **2 576** across all
  committees.
- The 18-row gap is exactly the 18 förslagspunkter in 2022-2026 that Riksdagen decided with
  two separate voteringar. The committee grain counts each of those twice; the coverage
  grain collapses each to one.
- Both numbers are correct for their own grain. Neither is wrong, and neither query is
  broken. Coverage compares against Riksdagen's own förslagspunkt-grained denominator, and
  RÖSTAT must show two rows where the record holds two voteringar with different party
  positions.
- Phase 2 is the first branch to render both on pages that link to each other: a party
  page's committee headings link straight to `/committees/:code`. Its fix was scope-limited
  to naming the grain on the committee page (`Räknat per votering. En förslagspunkt som
  avgjordes av två voteringar räknas som två.`); no query was changed and `/parties/:party`
  does not name its grain.

Whether the two surfaces should report one grain, and if so which, is a Phase 3 decision to
make with the human — it is a question about what "an omröstning" means to a reader, not a
defect to patch.

## File Structure

| File | Responsibility |
|---|---|
| `backend/internal/votes/ports/riksdagen.go` | +`FetchRecentBetankanden` on the client port |
| `backend/internal/votes/adapters/riksdagen/client.go` | Same `/dokumentlista` call, **no `organ` filter** |
| `backend/internal/votes/service.go` | +`GetRecentBetankanden`; **delete** `GetRiksdagFeed` |
| `backend/internal/votes/adapters/http/handler.go` | +`GET /votes/recent`; **delete** `/votes/riksdag-feed` |
| `frontend/src/shared/nav.ts` | **The one** nav definition |
| `frontend/src/App.tsx` | Consumes `nav.ts`; routes; `/sok` and `/agenda/:id` removed |
| `frontend/src/components/MobileNav.tsx` | Consumes `nav.ts` |
| `frontend/src/features/home/HomePage.tsx` | Feed → committee index → party index |
| `frontend/src/features/home/DecisionFeed.tsx` | The dated feed |
| `frontend/src/features/home/CommitteeIndex.tsx` | All committees, alphabetical |
| `frontend/src/features/regering/RegeringPage.tsx` | Document index |
| `frontend/src/features/regions/RegionLandingPage.tsx`, `municipalities/MunicipalityLandingPage.tsx` | The who-decides sentence |
| `frontend/src/features/votes/VotesPage.tsx` | Committee filter + honesty rider |
| `backend/migrations/000036_agenda_document_index.{up,down}.sql` | Drop `description`/`status`, add `url` |
| `docs/data-sources/national-agenda.md` | Rewritten |

Deleted: `frontend/src/features/search/SearchPage.tsx`, `frontend/src/features/riksdag/AgendaDetailPage.tsx`.

---

### Task 1: A complete betänkande feed

**Files:**
- Modify: `backend/internal/votes/ports/riksdagen.go`, `backend/internal/votes/adapters/riksdagen/client.go`, `backend/internal/votes/service.go`, `backend/internal/votes/adapters/http/handler.go`, `api/openapi.yaml`

**Interfaces:**
- Produces: `FetchRecentBetankanden(ctx context.Context, count int) ([]RiksdagDocument, error)`; `GET /api/votes/recent?count=40` → `RiksdagDocument[]` with `title`, `organ`, `date`, `beteckning`.

The existing `FetchDocuments(ctx, organs, count)` already does the right upstream call — `/dokumentlista/?organ=…&typ=bet&utformat=json&sz=N&sort=datum&sortorder=desc` (`client.go:172-174`) — and `RiksdagDocument` already carries `Date`, `Organ` and `Beteckning`. The only change needed is **omitting the organ filter**, which turns a curated feed into a complete one.

- [ ] **Step 1: Write the failing test**

Create `backend/internal/votes/adapters/riksdagen/recent_test.go`:

```go
package riksdagen_test

import (
	"context"
	"os"
	"testing"

	"riksdagskollen/internal/votes/adapters/riksdagen"
)

// Hits the live API. Skipped unless explicitly enabled, like the other smoke
// tests in this package.
func TestFetchRecentBetankanden_SpansManyCommittees(t *testing.T) {
	if os.Getenv("RIKSDAGEN_SMOKE") == "" {
		t.Skip("RIKSDAGEN_SMOKE not set — skipping live API test")
	}
	c := riksdagen.NewClient()
	docs, err := c.FetchRecentBetankanden(context.Background(), 40)
	if err != nil {
		t.Fatalf("FetchRecentBetankanden: %v", err)
	}
	if len(docs) == 0 {
		t.Fatal("no documents returned")
	}
	organs := map[string]bool{}
	for _, d := range docs {
		if d.Date == "" {
			t.Errorf("%s has no date — the feed must state the dates it covers", d.Beteckning)
		}
		if d.Organ == "" {
			t.Errorf("%s has no organ — every item must name its utskott", d.Beteckning)
		}
		organs[d.Organ] = true
	}
	// A complete feed is not a curated one. Five committees across 40 recent
	// betänkanden is a low bar; failing it means an organ filter is still applied.
	if len(organs) < 5 {
		t.Errorf("only %d committees in 40 items — is an organ filter still applied?", len(organs))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && RIKSDAGEN_SMOKE=1 go test ./internal/votes/adapters/riksdagen/ -run TestFetchRecentBetankanden -v`
Expected: FAIL with `c.FetchRecentBetankanden undefined`.

- [ ] **Step 3: Implement**

In `backend/internal/votes/adapters/riksdagen/client.go`, add beside `FetchDocuments`:

```go
// FetchRecentBetankanden returns the most recently published betänkanden across
// every committee, newest first.
//
// Same upstream call as FetchDocuments with the organ filter omitted. That
// omission is the point: a feed filtered to committees we chose would publish a
// worldview while claiming to show what happened. The period is the selection,
// and the calendar defines the period.
func (c *Client) FetchRecentBetankanden(ctx context.Context, count int) ([]ports.RiksdagDocument, error) {
	return c.fetchDocumentList(ctx, "", count)
}
```

Then refactor the body of `FetchDocuments` into `fetchDocumentList(ctx, organParam string, count int)`, which builds the URL with `&organ=` **only when `organParam != ""`**, and have `FetchDocuments` call it with `strings.Join(organs, ",")`. Keep the existing parsing untouched.

Add to the client port in `backend/internal/votes/ports/riksdagen.go`:

```go
	// FetchRecentBetankanden returns recent betänkanden across all committees.
	FetchRecentBetankanden(ctx context.Context, count int) ([]RiksdagDocument, error)
```

Adding a method to that interface breaks every fake implementing it. Find them with `grep -rn "FetchDocuments" backend --include=*_test.go` and add the method to each, returning `nil, nil`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && RIKSDAGEN_SMOKE=1 go test ./internal/votes/adapters/riksdagen/ -run TestFetchRecentBetankanden -v`
Expected: PASS, with at least 5 distinct committees.

- [ ] **Step 5: Service, endpoint, verify, commit**

Add `GetRecentBetankanden(ctx, count int)` to the service delegating to the client. Register `r.Get("/votes/recent", h.recent)` on the votes handler, add the path and a `RecentBetankande` schema to `api/openapi.yaml`, regenerate types.

```bash
go build -C backend ./... && cd frontend && npm run generate:api && npx tsc --noEmit && npm run lint
```

```bash
git add backend/internal/votes/ api/openapi.yaml frontend/src/shared/api-contract.ts
git commit -m "feat: complete betänkande feed across all committees

Same dokumentlista call as the curated feed with the organ filter omitted.
That omission is the decision: a feed filtered to committees we picked
would publish a worldview while claiming to show what happened."
```

---

### Task 2: One nav definition, Fördjupning tier, numerals gone, /sok removed

Four nav lists exist across two files — `App.tsx:191` and `:303`, `MobileNav.tsx:13` and `:22` — and they **have already drifted**: the same route is "Partimål & röstning" in one and "Partier" in the other.

**Files:**
- Create: `frontend/src/shared/nav.ts`
- Modify: `frontend/src/App.tsx`, `frontend/src/components/MobileNav.tsx`
- Delete: `frontend/src/features/search/SearchPage.tsx`

- [ ] **Step 1: Create the single definition**

Create `frontend/src/shared/nav.ts`:

```ts
/**
 * The site's navigation, defined once.
 *
 * Four copies of this existed across App.tsx and MobileNav.tsx and had already
 * drifted — the same route was labelled "Partimål & röstning" in one and
 * "Partier" in the other. One definition makes that impossible rather than
 * policed.
 */
export interface NavItem {
  to: string;
  label: string;
  /** Path prefixes that mark this item active. */
  match: string[];
}

/** The primary tier: the questions the site answers. */
export const PRIMARY_NAV: NavItem[] = [
  { to: "/", label: "FRÅGOR", match: ["/", "/committees"] },
  { to: "/om-sajten", label: "OM SAJTEN", match: ["/om-sajten"] },
  { to: "/data", label: "DATA", match: ["/data"] },
];

/**
 * The second tier: the four levels of government, as reference.
 *
 * Alphabetical. The previous order was numbered I RIKSDAG · II REGION ·
 * III KOMMUN · IV REGERING, and numerals read as precedence — an order neither
 * the Riksdag nor the constitution publishes, which also placed Regering fourth
 * when Riksdag and Regering are both national.
 */
export const FORDJUPNING_NAV: NavItem[] = [
  { to: "/kommun",   label: "Kommun",   match: ["/kommun"] },
  { to: "/region",   label: "Region",   match: ["/region"] },
  { to: "/regering", label: "Regering", match: ["/regering"] },
  { to: "/riksdag",  label: "Riksdag",  match: ["/riksdag", "/votes", "/budget", "/politicians", "/manifestos", "/parties"] },
];

/** True when the pathname sits under the item. */
export function isActive(item: NavItem, pathname: string): boolean {
  return item.match.some((m) =>
    m === "/" ? pathname === "/" : pathname === m || pathname.startsWith(m + "/"),
  );
}
```

- [ ] **Step 2: Consume it in both navs**

In `App.tsx`, delete the local `navPills` array and the riksdag sub-tab array, and render `PRIMARY_NAV` as the pill row with `FORDJUPNING_NAV` as a persistent second row labelled `FÖRDJUPNING`. Do the same in `MobileNav.tsx`. Neither file may define a nav item inline any more.

- [ ] **Step 3: Remove /sok**

```bash
git rm frontend/src/features/search/SearchPage.tsx
grep -rn "SearchPage\|/sok" frontend/src || echo "no references remain"
```

Remove the `/sok` route from `App.tsx`. `SearchPage` was 410 lines of client-side substring filtering over one preloaded list of votes, with no backend search endpoint — it could not find speeches, promises, budget lines or politicians. With every committee one click from the front page, structure replaces it.

- [ ] **Step 4: Verify**

With the API and dev server running, check in a browser at both a wide viewport and ≤640px:

- The top row reads `FRÅGOR · OM SAJTEN · DATA`, the second `FÖRDJUPNING: Kommun · Region · Regering · Riksdag`.
- **No roman numerals anywhere**: `grep -rn "I RIKSDAG\|II REGION\|III KOMMUN\|IV REGERING" frontend/src` returns nothing.
- `/sok` 404s at the router level; nothing links to it.
- Every Fördjupning item highlights on its own section, and `/parties/S` highlights `Riksdag`.

Then `npx tsc --noEmit && npm run lint && npm run build`.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/shared/nav.ts frontend/src/App.tsx frontend/src/components/MobileNav.tsx
git commit -m "refactor: define navigation once, drop the roman numerals

Four nav lists across two files had already drifted on the same route's
label. One definition now feeds desktop and mobile.

Fördjupning becomes a persistent second tier, alphabetical: numerals are
the strongest available ordering signal and the order they asserted is not
one the constitution publishes. /sok goes with it — a client-side substring
filter over one preloaded vote list, which structure replaces."
```

---

### Task 3: The front page — feed, committee index, party index

**Files:**
- Create: `frontend/src/features/home/DecisionFeed.tsx`, `frontend/src/features/home/CommitteeIndex.tsx`
- Modify: `frontend/src/features/home/HomePage.tsx`

- [ ] **Step 1: Build the feed**

Create `DecisionFeed.tsx` consuming `GET /api/votes/recent?count=40`. Requirements:

- **One item per betänkande**, each naming its utskott and linking `/committees/{organ}`, plus its own date via `swedishDate` from `@/shared/dates`.
- **The feed states the span it covers**, from the items' own dates: *"Beslut publicerade {äldsta} – {nyaste}"*. Do not compute or claim a "sitting period" — the dates are the fact.
- **When the newest item is old, say only that.** #98: never claim recess, never infer a calendar. The honest sentence is *"Inget nyare finns i vår inläsning."* with a `SourceMarker sourceId="riksdagen"` on the dates. Do **not** write "riksdagen har sommaruppehåll" — Riksdagen publishes no sourceable sitting calendar and the gap only reaches recess length once the recess is nearly over.
- Empty feed renders the row with the absence attributed to us, not to the Riksdag.

- [ ] **Step 2: Build the committee index**

Create `CommitteeIndex.tsx` consuming `GET /api/committees?period=…`:

- **All committees the record shows**, alphabetical by name, each linking `/committees/{code}`.
- Each shows its plain-language remit. Take it from the Riksdag's own published description; where none is held, show the name alone rather than writing one.
- **No count-based ordering and no "most active" marker.** The `voteringar` count may be displayed as a fact but must not order the list.

- [ ] **Step 3: Reassemble the front page**

In `HomePage.tsx`, order the blocks: **`DecisionFeed` → `CommitteeIndex` → party index** (eight `PartyBadge` links to `/parties/{code}`, alphabetical). Remove `AgendaList` (`:376`) — the agenda moves in Task 4. Keep the existing framing prose.

- [ ] **Step 4: Verify**

In a browser at `/`: the feed appears above the committee index above the party strip; every feed item names a committee and links to it; the committee index lists 16 including UFöU alphabetically; the party strip has 8 badges C…V; no agenda block remains. Then `npx tsc --noEmit && npm run lint && npm run build`.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/home/
git commit -m "feat: front page becomes feed, committees, parties

A complete dated feed of recent betänkanden above the committee index
above the party index. The betänkande is the unit, which makes a complete
week ~10-15 items instead of hundreds of near-identical vote points.

The feed states the dates it covers and, when nothing is newer, says only
that: Riksdagen publishes no sourceable sitting calendar, so claiming
recess would be an unsourced inference."
```

---

### Task 4: The agenda becomes a document index under /regering

**Files:**
- Create: `backend/migrations/000036_agenda_document_index.{up,down}.sql`
- Modify: `backend/internal/riksdag/` (agenda domain/repo/handler), `frontend/src/features/regering/RegeringPage.tsx`, `docs/data-sources/national-agenda.md`, `CLAUDE.md`
- Delete: `frontend/src/components/AgendaList.tsx`, `frontend/src/features/riksdag/AgendaDetailPage.tsx`

`riksdag_agenda` currently holds 5 rows, all seeded `2026-05-19` with `updated_at` never touched, each carrying our paraphrase in `description`, a free-text `source` with **no URL**, and a hand-set `status` (`active`/`in_progress`) asserting how the government is progressing.

- [ ] **Step 1: Write the migration**

`000036_agenda_document_index.up.sql`:

```sql
-- The agenda becomes an index of whole documents.
--
-- It previously held five hand-picked points from documents like Tidöavtalet,
-- each with a description we paraphrased and a status we assigned. Selecting
-- five of hundreds of points is an editorial act, the paraphrase is not the
-- document's words, and `status` was a claim about the present that had been
-- frozen since 2026-05-19. Selecting whole documents is far less bias-prone:
-- the reader gets the primary source and extracts for themselves.
ALTER TABLE riksdag_agenda ADD COLUMN IF NOT EXISTS url TEXT NOT NULL DEFAULT '';
ALTER TABLE riksdag_agenda ADD COLUMN IF NOT EXISTS issuer TEXT NOT NULL DEFAULT '';
ALTER TABLE riksdag_agenda ADD COLUMN IF NOT EXISTS published DATE;

ALTER TABLE riksdag_agenda DROP CONSTRAINT IF EXISTS riksdag_agenda_status_check;
ALTER TABLE riksdag_agenda DROP COLUMN IF EXISTS status;
ALTER TABLE riksdag_agenda DROP COLUMN IF EXISTS description;

-- The five curated points are not documents and cannot be salvaged into them.
DELETE FROM riksdag_agenda;

INSERT INTO riksdag_agenda (title, source, issuer, url, published, sort_order) VALUES
  ('Tidöavtalet', 'Tidöavtalet 2022', 'Moderaterna, Kristdemokraterna, Liberalerna, Sverigedemokraterna',
   'https://via.tt.se/data/attachments/00805/8ce1f0f1-4dbb-4c6c-9dcb-e1e5f1ca47c1.pdf', '2022-10-14', 1),
  ('Budgetpropositionen 2026', 'Prop. 2025/26:1', 'Regeringen',
   'https://www.regeringen.se/rattsliga-dokument/proposition/2025/09/prop.-2025261/', '2025-09-22', 2);
```

Verify both URLs resolve before committing. If either has moved, find the current canonical location — a document index whose links 404 is worse than no index. Add further documents only where the URL is verified.

`000036_agenda_document_index.down.sql`:

```sql
ALTER TABLE riksdag_agenda ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE riksdag_agenda ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE riksdag_agenda DROP COLUMN IF EXISTS published;
ALTER TABLE riksdag_agenda DROP COLUMN IF EXISTS issuer;
ALTER TABLE riksdag_agenda DROP COLUMN IF EXISTS url;
```

- [ ] **Step 2: Update the backend**

Remove `Description` and `Status` from the agenda domain type and its DTO; add `URL`, `Issuer`, `Published`. Update the SELECT columns and the OpenAPI schema. **Delete the `/agenda/{id}` endpoint** if one exists — an item is now a title plus a link, so a detail endpoint returns nothing a reader needs.

- [ ] **Step 3: Render it under /regering**

In `RegeringPage.tsx`, add a section listing the documents: title, issuer, publication date, and an outbound link. Above it, state the selection rule plainly, because the list is ours:

```tsx
<p>
  Regeringens program- och budgetdokument för mandatperioden, som vi har
  registrerat dem. Vi länkar dokumenten i sin helhet och sammanfattar dem inte.
  <SourceMarker sourceId="national-agenda" />
</p>
```

Then delete `AgendaList.tsx` and `AgendaDetailPage.tsx`, and remove the `/agenda/:id` route.

- [ ] **Step 4: Rewrite the data source doc**

Rewrite `docs/data-sources/national-agenda.md` to describe a document index: `kind: seed`, `last_verified` today, `used_by: /regering`. The Begränsningar section must state that the list is **ours and not the government's**, that a missing document means we have not registered it rather than that it does not exist, that we do not summarise or assign status, and that the previous version carried five paraphrased points with statuses frozen since 2026-05-19.

Run `cd frontend && npm run sync:data-sources`.

- [ ] **Step 5: Verify and commit**

```bash
make migrate
docker exec swedish-democratic-tracker-postgres-1 psql -U riksdagskollen -d riksdagskollen -c "SELECT title, issuer, url, published FROM riksdag_agenda ORDER BY sort_order;"
```

Then in a browser: `/regering` lists the documents with working outbound links; `/` no longer shows an agenda block; `/agenda/1` 404s; `/data/national-agenda` renders the rewritten body.

```bash
git add backend/migrations/000036_* backend/internal/riksdag/ frontend/src/features/regering/ \
        docs/data-sources/national-agenda.md api/openapi.yaml frontend/src/shared/api-contract.ts CLAUDE.md
git rm frontend/src/components/AgendaList.tsx frontend/src/features/riksdag/AgendaDetailPage.tsx
git commit -m "refactor: agenda becomes an index of whole documents

Five hand-picked points with paraphrased descriptions and hand-set
statuses become a list of whole documents with links. Selecting documents
is far less bias-prone than selecting points within them: the reader gets
the primary source instead of our reading of it.

status is gone. It asserted how the government was progressing, was never
sourced, and had been frozen at 2026-05-19. /agenda/:id goes with the
paraphrase, since an item is now a title and a link."
```

---

### Task 5: Strictly national — remove the invented committee→level feed

**Files:**
- Modify: `backend/internal/votes/service.go`, `backend/internal/votes/adapters/http/handler.go`, `api/openapi.yaml`
- Modify: `frontend/src/hooks/useDemocracy.ts`, `frontend/src/features/regions/RegionLandingPage.tsx`, `frontend/src/features/municipalities/MunicipalityLandingPage.tsx`

- [ ] **Step 1: Delete the curated feed**

Remove `GetRiksdagFeed` from `service.go:170-184`, the `/votes/riksdag-feed` route and handler, and its path from `api/openapi.yaml`. Remove `fetchRiksdagFeed` from `useDemocracy.ts:69-80` and its two call sites (`:333` in `useRegion`, `:483` in `useMunicipality`), along with whatever renders `feed` on those pages.

This mapping is exactly what #103 ruled does not honestly exist: `region → SoU, TU` is a judgement about which committees matter to a region, made by us and presented as relevance.

- [ ] **Step 2: Add the who-decides sentence**

On both `RegionLandingPage.tsx` and `MunicipalityLandingPage.tsx`, add near the top:

```tsx
<p>
  Riksdagen beslutar om statens budget. Din region beslutar om sin egen, med
  egna skatteintäkter och statsbidrag. De två är inte samma pengar.
</p>
```

For the municipality page, replace "region" with "kommun" throughout the sentence. This states the constitutional division of responsibility — a fact about who decides, not a data link. It exists so a reader arriving from a committee page does not conclude that the Riksdag runs their hospital.

- [ ] **Step 3: Verify and commit**

Browser-check `/region`, `/region/{code}`, `/kommun`, `/kommun/{code}`: no riksdag betänkande feed appears anywhere, the who-decides sentence reads correctly on both landing pages, and nothing 500s from the removed endpoint. Then `go build -C backend ./... && npx tsc --noEmit && npm run lint && npm run build`.

```bash
git add backend/internal/votes/ frontend/src/hooks/useDemocracy.ts frontend/src/features/regions/ \
        frontend/src/features/municipalities/ api/openapi.yaml frontend/src/shared/api-contract.ts
git commit -m "fix: stop presenting picked committees as regionally relevant

Every region page rendered betänkanden from SoU and TU, and every kommun
page from UbU, CU and SoU, because a hardcoded map said those committees
matter to those levels. That is a relevance judgement we invented and
published.

Replaced by a statement of who decides what, which is checkable: the
Riksdag decides the state budget, a region decides its own, and they are
not the same money."
```

---

## Self-Review

**Spec coverage.** The spec's Phase 3 was "front page becomes the 15 committees, nav restructured to Frågor / Fördjupning, `/sok` removed". All three are here — Tasks 3, 2 and 2 respectively — with the front page widened by #88 (feed above the index) and #100 (party index as a peer), and the committee count corrected by #97 from "the 15" to whatever the record shows. Tasks 4 and 5 are not in the spec: they come from #102 and #103, both later.

**Placeholders.** Backend work carries real Go and SQL. Frontend tasks specify every user-visible string verbatim plus the structural constraints, and defer JSX to the implementer following `HomePage`/`PartiesPage` — the same choice as Phase 2, for the same reason. Task 4 Step 1 contains two real URLs and an explicit instruction to verify them before committing rather than trusting them.

**Type consistency.** `NavItem`/`isActive` are defined once in Task 2 and consumed by both navs. `RiksdagDocument` is reused unchanged from `votes/ports/riksdagen.go:41-46` rather than a parallel type. The agenda's new columns (`url`, `issuer`, `published`) are named identically in migration, domain, DTO and OpenAPI.

**Risk worth stating.** Task 1's live smoke test is `RIKSDAGEN_SMOKE`-gated, so CI will not run it — the completeness assertion ("≥5 committees in 40 items") is only checked when someone runs it deliberately. Run it once locally before shipping Task 3, because a silently-still-filtered feed is the failure this whole phase exists to avoid, and it would look completely normal on the page.

**Ordering.** Task 1 before Task 3 (the feed needs its endpoint). Task 2 is independent. Task 4 removes `AgendaList` from `HomePage`, which Task 3 also edits — do Task 3 first, or expect a conflict in that file.
