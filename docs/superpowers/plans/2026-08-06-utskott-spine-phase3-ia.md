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

## Carried in from Phase 2: deferred minors

Every item below was found by a review during Phase 2, triaged **ship-as-is**, and merged
knowingly in #112. None is a live defect. They are recorded here because the execution
ledger they lived in was git-ignored scratch, so without this section the only trace of
them would be gone. Treat this as a menu for Phase 3, not a backlog that must be emptied.

**Performance**

- `upper(beteckning)` in the committee votes predicate defeats `idx_votes_beteckning`
  (`btree (beteckning, forslagspunkt)`), so each request does two sequential scans of
  ~899 000 rows. Measured 0.31–0.40 s per committee. A functional index on
  `upper(beteckning)` is the obvious fix and is a migration, not a query change.

**Correctness that today's data does not exercise**

- **The two committee counts can diverge by construction.** The header counts
  `DISTINCT (beteckning, votering_id)` (`committees/adapters/postgres/repository.go`), while
  RÖSTAT groups by `votering_id` alone and takes `min(beteckning)`. A `votering_id` carrying
  two beteckningar would count twice in the header and once in RÖSTAT. Separately, RÖSTAT's
  inner join filters `party <> '-'`, so a votering whose every ballot is unaffiliated is
  counted by the header and dropped by RÖSTAT. Zero rows of either kind exist today, and no
  test pins the invariant. They agreed for all 16 committees when checked.
- The votes repository's Go-side committee filter is case-sensitive while its SQL predicate
  is case-insensitive; only the HTTP handler's `domain.Canonical` call keeps the two layers
  aligned.
- `domain.Canonical` returns an unrecognised code verbatim (its own test pins
  `Canonical("ZZU") == "ZZU"`). A future committee code outside `canonicalCodes` would
  therefore answer zero rows for a lowercase URL and rows for the exact spelling — the kind
  of casing gate the committees handler exists to remove.
- The empty-canonical-code guard runs *before* the period check, so
  `/api/committees/%/votes?period=2018-2022` answers `200 {"items":[],"total":0}` where the
  sibling endpoint 404s. It survives only for codes `Canonical` rejects outright; a
  plausible-but-unknown code such as `ZZU` correctly 404s.

**Test hygiene**

- The `limit > 200` cap has no test — deleting it leaves all four handler tests green.
- `TestListByCommittee_UnknownPeriod`'s message assertion sits inside
  `if err := json.Unmarshal(...); err == nil`, so an unparseable body silently skips the
  message check (the status assertion still fires).
- `repo.calls != 0` in `TestListByCommittee_MissingPeriod` cannot fail under deletion of the
  guard it documents, because `PeriodExists("")` short-circuits a layer down. It is
  defence-in-depth against a future reordering, not the discriminating assertion.

**Architecture**

- The goals HTTP adapter imports `committees/domain` for `Canonical`, and now `matching/domain`
  as well. Neither breaks the "never import an adapter into a domain" rule and both match
  existing precedent, but the coupling is accumulating in one handler.
- Go's nil slice marshals to JSON `null`. One handler crashed the frontend on it during
  Phase 2 and was fixed; whether other handlers can return `null` where the contract says
  array was never audited.

**Presentation**

- `GoalVotesPage` match cards now carry no `SourceMarker` at all. Removing the misattributed
  Riksdagen marker was the ruling, but it left genuinely record-derived fields (beteckning,
  förslagspunkt, proposal origin) untagged, against CLAUDE.md rule 11(d).
- Both rows of a double-decided förslagspunkt link to the same `/votes/{bet}/{punkt}`, which
  merges the two voteringar — so the note telling the reader they are different decisions
  cannot be followed through to see the difference.
- `ORDER BY … pos.party, pt.votering_id` interleaves a duplicate pair's rows. The map-based
  fold reassembles them deterministically, so this is cosmetic.
- The mandate-period and coverage line still sits on the party page's Mål tab, which now
  shows no votes at all.

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

The existing `FetchDocuments(ctx, organs, count)` already does the right upstream call — `/dokumentlista/?organ=…&typ=bet&utformat=json&sz=N&sort=datum&sortorder=desc` (`client.go:172-174`) — and `RiksdagDocument` already carries `Date`, `Organ` and `Beteckning`.

> **Corrected 2026-08-14. `organ` is inert; the feed is already complete.** This task
> originally said the only change needed was "omitting the organ filter, which turns a
> curated feed into a complete one". Tested directly against `data.riksdagen.se`:
> `organ=SoU` returns 74 783 hits whose organs are `AU, JuU, UbU`; `organ=SoU,TU` behaves
> the same; omitting `organ` is identical. **Riksdagen ignores the parameter on this
> endpoint.**
>
> Two consequences. First, the change is still worth making — an inert parameter that looks
> like a filter is worse than no parameter — but it changes nothing about what the feed
> returns, and the commit message must not claim otherwise. Second, the completeness
> assertion in the test below (*"is an organ filter still applied?"*) **cannot fail**: it
> passes whether or not the filter is present, because the filter does nothing. It is
> replaced in Step 1 by a test that can fail.
>
> It also means the defect Task 5 removes is worse than #103 described: region pages never
> showed "SoU and TU", they showed the newest betänkanden from any committee at all,
> labelled as regionally relevant.

- [ ] **Step 1: Write the failing test**

Add to `backend/internal/votes/adapters/riksdagen/client_test.go`, which is already
`package riksdagen` and already has the `newTestClient(srv.URL)` helper and the pattern for
asserting a query parameter is **absent** (see `TestFetchVotes_…`'s check on `p`):

```go
// The organ parameter must not be sent at all.
//
// Riksdagen ignores `organ` on /dokumentlista — organ=SoU returns betänkanden
// from AU, JuU and UbU — so no assertion about the organs that come *back* can
// detect whether we filtered. The only thing that can fail is what we send.
func TestFetchRecentBetankanden_SendsNoOrganFilter(t *testing.T) {
	var q url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dokumentlista":{"dokument":[
			{"titel":"T","organ":"SoU","datum":"2026-08-12","beteckning":"SoU1"}]}}`))
	}))
	defer srv.Close()

	docs, err := newTestClient(srv.URL).FetchRecentBetankanden(context.Background(), 40)
	if err != nil {
		t.Fatalf("FetchRecentBetankanden: %v", err)
	}
	if _, ok := q["organ"]; ok {
		t.Errorf("client sent organ=%q; a complete feed must send no organ filter", q.Get("organ"))
	}
	if got := q.Get("typ"); got != "bet" {
		t.Errorf("typ = %q, want \"bet\"", got)
	}
	if got := q.Get("sz"); got != "40" {
		t.Errorf("sz = %q, want \"40\"", got)
	}
	if len(docs) != 1 || docs[0].Organ != "SoU" || docs[0].Date != "2026-08-12" {
		t.Errorf("parsed %+v, want one SoU doc dated 2026-08-12", docs)
	}
}

// FetchDocuments must keep sending its organ parameter unchanged, even though
// Riksdagen ignores it: this test exists so the refactor that adds
// FetchRecentBetankanden cannot silently change the older call's behaviour.
func TestFetchDocuments_StillSendsOrgan(t *testing.T) {
	var q url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dokumentlista":{"dokument":[]}}`))
	}))
	defer srv.Close()

	if _, err := newTestClient(srv.URL).FetchDocuments(context.Background(),
		[]string{"SoU", "TU"}, 5); err != nil {
		t.Fatalf("FetchDocuments: %v", err)
	}
	if got := q.Get("organ"); got != "SoU,TU" {
		t.Errorf("organ = %q, want \"SoU,TU\"", got)
	}
}
```

**Mutation-test both** before moving on: make `FetchRecentBetankanden` pass a non-empty
organ and confirm the first test fails; drop the organ from `FetchDocuments` and confirm the
second fails. Restore both.

Optionally keep a live smoke test in `smoke_test.go` behind `RIKSDAGEN_SMOKE` asserting only
that every returned document carries a `Date` and an `Organ`. Do **not** assert a committee
count: `organ` is inert upstream, so such an assertion passes with or without a filter and
would be a guard that cannot fail. Note also that the upstream intermittently returns zero
hits for an identical query, so a live test must not treat an empty result as a failure.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/votes/adapters/riksdagen/ -run 'TestFetchRecentBetankanden|TestFetchDocuments_StillSendsOrgan' -v -count=1`
Expected: FAIL with `c.FetchRecentBetankanden undefined`. No network access needed — both tests run against `httptest`.

- [ ] **Step 3: Implement**

In `backend/internal/votes/adapters/riksdagen/client.go`, add beside `FetchDocuments`:

```go
// FetchRecentBetankanden returns the most recently published betänkanden across
// every committee, newest first.
//
// Sends no organ parameter. A feed filtered to committees we chose would
// publish a worldview while claiming to show what happened; the period is the
// selection, and the calendar defines the period.
//
// This does not change what Riksdagen returns — see the note on FetchDocuments.
func (c *Client) FetchRecentBetankanden(ctx context.Context, count int) ([]ports.RiksdagDocument, error) {
	return c.fetchDocumentList(ctx, "", count)
}
```

Then refactor the body of `FetchDocuments` into `fetchDocumentList(ctx, organParam string, count int)`, which builds the URL with `&organ=` **only when `organParam != ""`**, and have `FetchDocuments` call it with `strings.Join(organs, ",")`. Keep the existing parsing untouched.

**Record the inert parameter where the next developer will hit it.** Put this on
`FetchDocuments`, whose callers still pass organs and may reasonably assume it filters:

```go
// NOTE: Riksdagen ignores `organ` on /dokumentlista. Verified 2026-08-14:
// organ=SoU returns 74 783 hits whose organs include AU, JuU and UbU, and
// omitting the parameter gives the identical result. Callers that need a
// specific committee must filter the returned documents themselves — passing
// organs here selects nothing.
```

Add to the client port in `backend/internal/votes/ports/riksdagen.go`:

```go
	// FetchRecentBetankanden returns recent betänkanden across all committees.
	FetchRecentBetankanden(ctx context.Context, count int) ([]RiksdagDocument, error)
```

Adding a method to that interface breaks every fake implementing it. Find them with `grep -rn "FetchDocuments" backend --include=*_test.go` and add the method to each, returning `nil, nil`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/votes/adapters/riksdagen/ -run 'TestFetchRecentBetankanden|TestFetchDocuments_StillSendsOrgan' -v -count=1`
Expected: PASS — no `organ` parameter sent by the new call, `organ=SoU,TU` still sent by the old one. Confirm from `-v` that both tests ran.

- [ ] **Step 5: Service, endpoint, verify, commit**

Add `GetRecentBetankanden(ctx, count int)` to the service delegating to the client. Register `r.Get("/votes/recent", h.recent)` on the votes handler, add the path and a `RecentBetankande` schema to `api/openapi.yaml`, regenerate types.

```bash
go build -C backend ./... && cd frontend && npm run generate:api && npx tsc --noEmit && npm run lint
```

```bash
git add backend/internal/votes/ api/openapi.yaml frontend/src/shared/api-contract.ts
git commit -m "feat: complete betankande feed across all committees

Same dokumentlista call with no organ parameter. The intent is that a feed
filtered to committees we picked would publish a worldview while claiming
to show what happened.

It does not change what comes back. Riksdagen ignores organ on this
endpoint — organ=SoU returns betankanden from AU, JuU and UbU — so the
existing 'curated' feed was never curated. Sending a parameter that looks
like a filter and is not is worse than sending none, which is what this
removes. The test asserts on what we send, since nothing about what comes
back can distinguish the two."
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

> **Corrected 2026-08-14, after the task review. The prescribed copy was itself the defect.**
>
> This step's wording — a block headed *senaste besluten* and a span sentence reading
> *"Beslut publicerade {äldsta} – {nyaste}"* — treats every item in the feed as a decision.
> Verified against the live upstream: **15 of 40 items are `status: "planerat"`,
> `beslutad: 0`, `rm: "2026/27"`** — planned betänkanden for the coming riksmöte, not
> written, debated or decided. They are the newest by `datum`, so five of them occupied the
> **top eight rows** under that heading, two carrying a `justeringsdag` four months in the
> future.
>
> Separately, **`datum` is never the decision date**: on all 24 decided items `beslutsdag`
> falls 1–6 days later. So every date rendered under a "Beslut" heading was the wrong date
> for the event named. And *"Inget nyare finns i vår inläsning."* was contradicted on the
> same page render by `useRecordCoverage().lastDecisionDate`, which was newer.
>
> **Ruling: the unit is the betänkande**, which is what #88 actually decided. The block and
> the span sentence say betänkanden, not beslut. `status`, `beslutad` and `beslutsdag` are
> carried through the client (which discarded them), a decided row is dated by `beslutsdag`,
> a planned row makes no date claim and is labelled as not yet decided, the span is computed
> from decided rows only, and the "nothing newer" sentence is scoped to this feed rather
> than to our reading as a whole.

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

> **ALREADY DONE — DO NOT EXECUTE. Verified 2026-08-14.**
>
> PR [#110](https://github.com/Runemark7/swedish-democratic-tracker/pull/110) (`98ad934`)
> completed this task in full: the migration shipped as `000037_agenda_document_index`
> (`000036` was already taken by `utskott_uo_fk`), `riksdag_agenda` carries `url`, `issuer`
> and `published` with `status` and `description` dropped, `AgendaList.tsx` and
> `AgendaDetailPage.tsx` are deleted, and `docs/data-sources/national-agenda.md` is
> rewritten.
>
> **Running the migration below would be destructive.** It does `DELETE FROM riksdag_agenda`
> and re-inserts the two URLs written here — and both of those URLs now 404, while the two
> actually shipped return 200:
>
> | | in this plan | shipped, working |
> |---|---|---|
> | Tidöavtalet | `via.tt.se/data/attachments/00805/…pdf` — **404** | `liberalerna.se/wp-content/uploads/tidoavtalet-…pdf` — **200** |
> | Budgetprop. 2026 | `regeringen.se/…/prop.-2025261/` — **404** | `regeringen.se/…/2025261/` — **200** |
>
> Someone already followed this task's own instruction to verify the URLs before committing
> and corrected them. Executing the task would undo that and leave a document index whose
> links are dead — which this task's Step 1 calls "worse than no index". The text is kept
> below as the record of what was decided and why.

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
- Modify: `frontend/src/hooks/useDemocracy.ts`
- Modify (the who-decides sentence): `frontend/src/features/regions/RegionLandingPage.tsx`, `frontend/src/features/municipalities/MunicipalityLandingPage.tsx`
- Modify (**this is where the feed actually renders**): `frontend/src/features/regions/RegionDetailPage.tsx`, `frontend/src/features/municipalities/MunicipalityDetailPage.tsx`

> **Corrected 2026-08-14.** The file list above originally named only the *landing* pages.
> The who-decides sentence does belong there, but the feed is rendered on the **detail**
> pages, as `liveVotes` — `RegionDetailPage.tsx:425` and `MunicipalityDetailPage.tsx:449`.
> `useDemocracy.ts` sets `liveVotes: feed` at `:369` (region) and `:519` (kommun).
>
> **`HomePage.tsx:60` and `RiksdagPage.tsx:616` also read `liveVotes` and must not be
> touched.** Theirs comes from `useRiksdag`, which sources it from the votes API, not from
> `/votes/riksdag-feed`. Removing the endpoint does not affect them; editing them would
> break two working pages.

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
git commit -m "fix: stop presenting national betankanden as regionally relevant

Every region and kommun page rendered a riksdag betankande feed because a
hardcoded map said certain committees matter to certain levels. That is a
relevance judgement we invented and published.

It was worse than the map suggested. Riksdagen ignores the organ parameter
on /dokumentlista, so the map selected nothing: the pages showed whichever
betankanden were newest, from any committee, presented as relevant to that
region or kommun.

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
