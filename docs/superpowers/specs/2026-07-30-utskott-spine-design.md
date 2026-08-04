# Utskott spine — refocusing the site around citizen questions

**Date:** 2026-07-30
**Status:** approved (design)
**Area:** information architecture, `riksdag` / `goals` / `votes` / `budget`

## Why

The site drifted from its purpose. A citizen who lands on it today finds:

- a front page labelled `LIVE · 30.07.2026` showing votes from **2026-04-18** and
  speeches from **2025-09-04**
- ~30 routes across four sections named after the state's org chart
  (Riksdag / Region / Kommun / Regering)
- no single place that answers an actual question

The stated goal is narrower and sharper than what got built: give an ordinary
Swede a fighting chance to see what those they elected actually did, so they can
decide whether to keep voting for them. Today that requires deep research *about
the site itself* before any research about politics can begin.

## The selection problem

The obvious fix — feature a few issues (kärnkraft, vårdköer, migration) — is
wrong, and was rejected during design. Choosing which issues appear **is** the
editorial act. `CONTEXT.md` states it directly: bias does not enter through the
raw number, it enters through the selection of what gets highlighted. A curated
issue list publishes a worldview while claiming neutrality.

The resolution: **complete coverage is the neutrality guarantee.** If everything
is present, nothing has been selected. So the site adopts an official, exhaustive
partition of Swedish politics that we did not choose.

## The spine: utskott (parliamentary committees)

The Riksdag's 15 standing committees are an official, exhaustive partition. We
do not pick them; the Riksdag does. Every proposal is handled by exactly one.

Crucially, this taxonomy already joins all three layers of the site with **no new
inference** — verified against live production data:

| Layer | Existing join | Verified coverage |
|---|---|---|
| Promises | `party_goals.relevant_committees` | **100%** of goals, all 8 parties |
| Votes | `beteckning` prefix (`AU9` → `AU`) | every vote row |
| Money | official Riksdag utskott→utgiftsområde allocation | all 27 UO |

Verification run 2026-07-30 against `sdt.runevibe.se`:

```
S: 13/13   M: 10/10   SD: 10/10   C: 9/9
V:  9/9    KD: 10/10  L:  9/9     MP: 10/10
committees seen: AU CU FiU FöU JuU KU MJU NU SfU SkU SoU TU UbU UU
```

"Kärnkraft" is therefore not a featured issue. It is what a reader finds *inside*
Näringsutskottet, next to everything else NU handled. Nothing is promoted and
nothing is omitted.

## The committee page

Each of the 15 committees gets one page, with three rows:

| Row | Source | Nature of the claim |
|---|---|---|
| **LOVAT** | `party_goals` where `relevant_committees` contains this committee | verbatim quote + source link |
| **RÖSTAT** | votes whose `beteckning` starts with this committee code, with each party's position | fact, from the ballots |
| **KOSTAR** | utgiftsområden allocated to this committee, with year-over-year change | raw kronor + delta |

### RÖSTAT shows how parties voted, not whether they kept a promise

An earlier draft proposed curating a direction ("voting Ja here honours this
promise") per vote. Under complete coverage that does not scale — it would mean
hand-curating every vote in 15 committees — and more importantly it reintroduces
judgement.

So RÖSTAT states how each party voted. The promise sits beside it. The reader
connects them. This is both more honest and removes the curation bottleneck
entirely.

### Labels must not overclaim

`party_goals.relevant_committees` is committee-level, so the LOVAT heading reads
"Partimål inom Näringsutskottets område" — not "Löften om kärnkraft". Honest
labelling is free and prevents overclaiming.

### No causal money claims

Showing UO20 down 1,1 mdkr beside UO21 up 4,2 mdkr **must not** imply one funded
the other. The budget data cannot support that, and the inference is exactly the
kind that lets a hostile reader dismiss the site.

Utgiftsområden are listed with their real changes under a neutral heading
("Utgiftsområden som utskottet bereder"). Never "detta finansierades av". The
reader sees both numbers and draws their own link.

## Front page

All 15 committees, each with its plain-language remit (taken from the Riksdag's
own published description, not written by us) and its most recent activity.

**Ordering must be neutral and deterministic**: alphabetical by Swedish committee
name (Arbetsmarknadsutskottet … Utrikesutskottet). Never by "importance",
activity volume, or recency of decision — any of those is a ranking, and ranking
is itself an editorial act.

## Navigation

```
FRÅGOR · Om sajten · DATA
FÖRDJUPNING: Riksdag · Region · Kommun · Regering
```

The four org-chart sections survive intact as a second tier, reached from a
committee page when the reader wants the raw underlying data, or directly.
Apart from `/sok` (see "What is removed"), this is an information-architecture
change only: **no other routes deleted, no features rebuilt.**

Region and Kommun gain value from this rather than losing it — a committee page
can link to the regional KPI that corresponds to its area, connecting national
promises to the level citizens actually feel.

## What is removed

**Search.** `frontend/src/features/search/SearchPage.tsx` (410 LOC) is a
client-side substring filter over one preloaded list of votes
(`SearchPage.tsx:66`). There is no backend search endpoint. It cannot find
speeches, promises, budget lines or politicians. With every committee one click
from the front page, structure replaces it.

Removing it also deletes the `/sok` route and its nav entry.

## Prerequisite: fix the data spine

Nothing above is worth building on stale data. This must land first.

### Cursor bug

`backend/internal/ingestion/workers/votes.go:47` (same pattern in the speeches
worker):

```go
// Update cursor to today
now := time.Now().UTC().Truncate(24 * time.Hour)
w.cursors.Upsert(ctx, Cursor{DataType: "votes", LastDate: &now})
```

The cursor advances to today **unconditionally** — even when the fetch failed
(errors above are logged as warnings and `continue`) and even when zero rows were
returned. Once it drifts, every later run asks Riksdagen "anything since today?",
gets nothing, logs `worker done`, and advances the cursor again. The gap is
skipped permanently.

Production log, 2026-07-30: `"votes: incremental sync","since":"2026-07-30"` —
completed in 3s having fetched nothing.

Fix: advance the cursor to the newest record actually fetched; never on a failed
fetch; leave it unchanged when nothing was returned. Plus a one-time backfill to
recover the skipped window.

### Freshness honesty

Replace the `LIVE · <today>` header with the real data date
("Senaste beslut: 18 apr 2026"). If the pipeline stalls again the UI must say so.
A LIVE badge over stale data is the most damaging possible bug on a site built
for distrustful readers.

## Bugs that break the completeness premise

Both verified empirically on 2026-07-30, both must be fixed:

1. **`MJU` never resolves.** `committeeFromBeteckning`
   (`frontend/src/shared/design.ts:103`) uses
   `/^([A-ZÅÄÖ][a-zåäöÅÄÖ]*U)/`. The character class excludes uppercase `J`, so
   `MJU12` returns `null`. Miljö- och jordbruksutskottet is invisible — and MJU
   appears in the goals of 7 of 8 parties.

2. **`KrU` is missing.** The `COMMITTEES` map (`design.ts:85`) contains 14 of
   15 committees; Kulturutskottet is absent. `KrU5` extracts correctly but maps
   to nothing.

A spine claiming complete coverage cannot silently drop two committees.

## New data

One seeded mapping table: **utskott → utgiftsområde**. This is a transcription of
the Riksdag's own published allocation, not a judgement. It requires a
`docs/data-sources/` entry with source and reproduction steps per CLAUDE.md
rule 11, and must be registered in the source registry.

Everything else derives from data already in the database.

## Explicitly not doing

- **No populism index or any computed verdict.** `party_goals.specificity`
  (`concrete` / `directional` / `rhetorical`) already answers "how much of this
  is talk" as a property of the party's own wording, verifiable against the
  manifesto. A score would be an opinion and would hand critics a weapon.
- **No touching `inferDirection`.** Out of scope, as decided in the
  2026-07-29 alignment work.
- **No rebuilding Region/Kommun.**
- **No deleting routes** other than `/sok`.
- **No editorial issue list.**

## Deferred

`alignmentPct` on the party-goal pages stays as fixed on 2026-07-29 (honest
nulls). With RÖSTAT presenting votes as fact, the computed alignment number may
become redundant; whether to retire it is a separate decision, not part of this
work.

## Phasing

This spec is too large for one implementation plan. It decomposes into three,
each shippable on its own:

1. **Data spine** — cursor fix, backfill, freshness honesty, `MJU`/`KrU` bug
   fixes. Independent of everything else and blocks the rest. Ship first.
2. **Committee pages** — utskott→UO mapping table + data source doc, backend
   endpoints, the committee page with its three rows.
3. **IA switch** — front page becomes the 15 committees, nav restructured to
   Frågor / Fördjupning, `/sok` removed.

Phase 2 is only worth building on phase 1's data. Phase 3 is only worth
switching to once phase 2 has something to switch to.

## Verification

- Every committee in `COMMITTEES` resolves from a real `beteckning`, including
  `MJU` and `KrU`
- All 15 committees reachable from the front page
- Committee page renders all three rows with `SourceMarker` on every value
- Ingestion: cursor does not advance past the newest fetched record; a failed
  fetch leaves it unchanged
- Front page shows real data freshness, not today's date
- `go build -C backend ./...` and `npx tsc --noEmit` clean
