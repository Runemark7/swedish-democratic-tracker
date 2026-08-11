# Handoff — next up: utskott spine Phase 2

**Written:** 2026-08-11. **State of main:** `98ad934`, deployed and verified on prod.

Read this, then read the Phase 2 plan. Everything else here is context that plan cannot carry.

---

## Nothing is at risk

Unlike the previous handoff, there is no unpushed work. Every artifact is on `main`:

- `docs/superpowers/plans/2026-08-05-utskott-spine-phase1-data.md` (done, merged)
- `docs/superpowers/plans/2026-08-06-utskott-spine-phase2-pages.md` ← **your plan**
- `docs/superpowers/plans/2026-08-06-utskott-spine-phase3-ia.md`
- `docs/superpowers/specs/2026-07-30-utskott-spine-design.md`

The Phase 1 SDD ledger was deleted after merge, deliberately — git history is the record. `git log --oneline 8e014a4..7247f36` is the task-by-task trail.

## Do this next

Execute `docs/superpowers/plans/2026-08-06-utskott-spine-phase2-pages.md` with
`superpowers:subagent-driven-development`. Four tasks. **Task 4 must precede Task 3** — the
committee page renders the shared `GoalCard` that Task 4 creates. The plan says so; it is easy to miss.

Phase 2 is also what removes a live defect (below), so it is not only feature work.

## Two honesty defects live in production right now

Both decided, both are Phase 2/3 work, both verifiable in one command.

**1. The goal vote count** — Phase 2 Task 4.

```bash
curl -s "https://sdt.runevibe.se/api/parties/S/goals" | head -c 200   # relevantVotes: 14
```

Party pages say "N matchade omröstningar", and for unmatched goals
`Ej prövad i någon omröstning ännu` **with `<SourceMarker sourceId="riksdagen" />`** — our
keyword-list gap stated as a verdict about a party and attributed to the record.
[#100](https://github.com/Runemark7/swedish-democratic-tracker/issues/100) killed it. No
denominator beside a goal is honest: the keyword count under-claims, and the committee's full
count over-claims (a committee deciding 40 things does not mean one promise was tested 40 times).

**2. An invented committee→region relevance mapping** — Phase 3 Task 5.

```bash
curl -s "https://sdt.runevibe.se/api/votes/riksdag-feed?level=region"   # 5 items, FiU/Vård/Plan
```

`GetRiksdagFeed` (`backend/internal/votes/service.go`) hardcodes `region → SoU, TU` and
`kommun → UbU, CU, SoU`; `useRegion` calls it (`frontend/src/hooks/useDemocracy.ts`). Every region
page publishes *our* judgement about which committees matter to a region.
[#103](https://github.com/Runemark7/swedish-democratic-tracker/issues/103) ruled that mapping does
not honestly exist.

## Facts that cost real time to learn

**The record's grain.** `votes` has **no decision date** — only `system_datum`, which is when
Riksdagen last *touched* the row, and must never be rendered as a decision date. Beteckning
numbering **restarts each riksmöte**, so any `DISTINCT beteckning || forslagspunkt` without
`session` collapses `FiU1` across years: that undercounted every committee by 27% (1 870 vs 2 576)
and shipped through five reviews before the whole-branch review caught it. Use
`count(DISTINCT votering_id)`, or include `session`.

**`ListDistinctByCommitteePrefix` (`votes/adapters/postgres/repository.go`) filters
`origin_enriched = true`.** Do not reuse it for RÖSTAT — enrichment is about who proposed a bill,
and gating on it silently removes real voteringar from a record of how parties voted. It also
orders `beteckning` as text, so `AU10` sorts before `AU9`.

**`budget_years` has a `status` column and no unique constraint on `year`.** Always filter
`status = 'decided'` — the budget feature does, in three places. Omitting it doubles every amount
the day a `proposed` row appears.

**`expenditure_areas` has `sort_order`.** Use it, not a regexp cast on the code — a non-numeric code
makes the cast *error*.

**`shared/types.ts` hand-mirrors the generated `api-contract.ts`.** Every new API field must be added
in both or `tsc` fails. This is open fog on the map; do not fix it mid-task, but expect it.

**`SectionSource` does not exist** despite `CLAUDE.md` rule 11(d) naming it. Only `SourceMarker` is
implemented. Use what exists.

**There is no frontend component test runner** — only `test:scripts` (`node --test scripts/*.mjs`).
Frontend verification is `tsc` + `lint` + `build` + a scripted browser check. Do not write test steps
that cannot run.

## Working in this environment

**The dev Postgres container stops on its own.** `swedish-democratic-tracker-postgres-1`. When
`make migrate` fails with `connection refused`:

```bash
docker start swedish-democratic-tracker-postgres-1
until docker exec swedish-democratic-tracker-postgres-1 pg_isready -U riksdagskollen -d riksdagskollen >/dev/null 2>&1; do sleep 2; done
```

The backend container also applies migrations on startup, so `make migrate` often reports
`no change` even when your migration did apply. Verify against the schema, not the command output.

**DB-backed Go tests need the DSN explicitly**, and skip silently without it:

```bash
cd backend && TEST_DATABASE_URL="postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable" go test ./... -count=1
```

**Editor diagnostics on new Go files are stale roughly every time.** Six false reports in one
session — `undefined: X`, `does not implement Y` — all on freshly created packages, all while
`go build -C backend ./...` was clean. Verify with the compiler, never the squiggle. One was
semantically specific enough to look real (a genuine interface-satisfaction message) and was still
stale, caught mid-edit between adding a port method and its implementation.

**`pkill -f <pattern>` kills your own shell** when the pattern appears in the command line — exit
144, repeatedly. Use `pgrep -f`, then `kill` the specific PID. Local API/vite servers started with
`timeout N` self-terminate; leaving them is usually fine.

**`gh pr merge` prints an error after succeeding** when the tree is dirty — it fails updating the
local checkout, not the merge. `frontend/tsconfig.tsbuildinfo` is tracked and regenerated by every
build, so it is always dirty. Before merging: stash `.claude/settings.local.json` (it is the user's,
pre-existing), `git restore frontend/tsconfig.tsbuildinfo`. Then check `gh pr view N --json state`
rather than trusting the exit code.

**CI is two workflows** — `CI` and `verify-sources`. `gh run list --limit 1` may hand you the wrong
one and make the build jobs look absent. Query by name.

**Wait loops must pin the exact SHA.** A loop that polls `--limit 1` will match a *stale completed*
run and report green for code that no longer exists. This nearly caused a merge on the wrong
evidence:

```bash
SHA=$(git rev-parse HEAD)
until [ "$(gh run list --branch "$B" --limit 1 --json headSha,status --jq ".[0] | select(.headSha==\"$SHA\") | .status")" = "completed" ]; do sleep 20; done
```

## Verifying a deploy

CI green is not deployed. The chain is: CI on main → a `ci: update image tags` commit → ArgoCD sync.
Verify by watching for what actually changed, then assert content:

```bash
curl -s https://sdt.runevibe.se/ | grep -o 'assets/index-[A-Za-z0-9_-]*\.js' | head -1   # bundle flips
curl -s https://sdt.runevibe.se/api/riksdag/coverage                                     # 200 = pod healthy
```

Migrations run on backend startup, so a healthy API after a migration is the proof the migration
applied. Arm the watch **before** the flip — a watch started late records the new bundle as its
baseline and then waits forever.

## The judgement that governs this codebase

Read `CLAUDE.md`'s appended project context if you have not. The operative rule: this site's failure
mode is not crashing, it is **publishing something that looks authoritative and is wrong**. Every
serious defect found in the last two sessions had that shape and all of them passed their tests:

- a LIVE badge showing the browser's clock
- an alignment percentage that quietly favoured the governing bloc
- a coverage denominator frozen while its numerator moved
- "förväntas presenteras" about a manifesto published a month earlier
- five hand-picked government "priorities" with a progress status we assigned
- every committee's vote count 27% low

Before shipping any displayed number, ask what it would take for it to mislead while everything
still goes green.

Two habits that repeatedly earned their cost: **mutation-test the guard** (break the code on
purpose and confirm the test fails — twice this session a test that "passed" would not have caught
its own defect), and **verify claims against the live database or the live site**, not against the
document that asserts them. Two of the plan's own stated numbers were wrong, and both were caught
that way.

## Map #82 has no open tickets

[Map #82](https://github.com/Runemark7/swedish-democratic-tracker/issues/82) reached its
destination — 14 tickets resolved, nothing left to decide. Its **Out of scope** section is the
execution queue; its **Not yet specified** section holds two fog patches (whether `topic` earns its
place as a tag, and whether `shared/types.ts` should exist). Do not open new wayfinder tickets for
Phase 2 — it is planned work, not a decision.
