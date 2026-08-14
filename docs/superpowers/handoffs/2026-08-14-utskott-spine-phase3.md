# Handoff — utskott spine Phase 3, mid-execution

**Written:** 2026-08-14. **Branch:** `feat/utskott-spine-phase3`, 9 commits ahead of
`main` (`2c78164`). **Phase 2 is merged and live** (PR #112).

Read this, then read `docs/superpowers/plans/2026-08-06-utskott-spine-phase3-ia.md`.
Everything below is what that plan cannot tell you.

---

## Where execution stopped

| Task | State |
|---|---|
| 1 — complete betänkande feed endpoint | **Complete**, reviewed clean (`e5c125e`..`4020921`) |
| 2 — one nav definition, Fördjupning, `/sok` gone | **Complete**, reviewed clean (`4020921`..`fe756a9`) |
| 3 — front page | Implemented + **fix wave applied**, `a40880f`, `8d7e5e9`, `d83180a`. **Its scoped re-review has NOT been run.** |
| 4 — agenda → document index | **DO NOT EXECUTE.** Already shipped by PR #110. See below. |
| 5 — remove the invented committee→level feed | **Not started.** |

**Your immediate next step** is the scoped re-review of Task 3's fix wave — the range
`a40880f..d83180a` — because those fixes were written by the controller after the fix
agent hit the spend limit, and so have had **no independent review at all**. Then Task 5,
then the final whole-branch review.

## Nothing is lost, but the ledger is not in git

The execution ledger lives at
`.superpowers/sdd/2026-08-06-utskott-spine-phase3-ia/progress.md`, which is **gitignored
scratch**. It holds the full finding-by-finding trail. This document carries everything
from it that outlives the branch; if the two disagree, the ledger is more detailed and
this one is more durable.

## Task 4 is done, and running it would do damage

PR #110 (`98ad934`) completed it. The migration shipped as `000037` because `000036` was
taken, `riksdag_agenda` carries `url`/`issuer`/`published` with `status`/`description`
dropped, both frontend files are deleted, and the data-source doc is rewritten.

The plan's Step 1 does `DELETE FROM riksdag_agenda` and re-inserts two URLs. **Both of the
plan's URLs 404. Both of the shipped ones return 200.** Someone already followed the task's
own instruction to verify them and fixed them. The task is marked do-not-execute in the plan
with the evidence; do not undo that.

## Three premises in this plan did not survive contact

1. **Riksdagen ignores `organ` on `/dokumentlista`.** `organ=SoU` returns 74 783 hits whose
   organs include `AU`, `JuU`, `UbU`; omitting it is identical. The "curated" feed was never
   curated. Task 1's original test asserted "≥5 committees means no filter", which passes
   either way — a guard that cannot fail. It was replaced with `httptest` cases asserting
   what the client *sends*, which is the only thing that can differ.
2. **Task 5's file list names the wrong pages.** The feed renders on
   `RegionDetailPage.tsx:425` and `MunicipalityDetailPage.tsx:449` as `liveVotes`, not on the
   landing pages the plan lists (those get the who-decides sentence).
   **`HomePage.tsx` and `RiksdagPage.tsx` also read `liveVotes` but source it from
   `useRiksdag`/the votes API — leave them alone.**
3. **Task 3's prescribed copy was itself the defect** — see below.

## The defect Task 3's fix wave removed, because Task 5 inherits its shape

`/dokumentlista` mixes decided betänkanden with merely *planned* ones and sorts both on
`datum`. On 2026-08-14, **15 of 40 were `status: "planerat"`, riksmöte 2026/27**, five of
them in the top eight rows — under a heading reading `SENASTE BESLUTEN`. Two carried a
`justeringsdag` four months in the future. And `datum` is never the decision date:
`beslutsdag` fell 1–6 days later on **all 24** decided rows.

Rulings applied: the unit is the betänkande (per #88), decided rows are dated by
`beslutsdag`, planned rows make **no date claim** and say `ännu inte beslutat`, the span is
computed from decided rows only, and the "nothing newer" sentence is scoped to this document
list — it previously contradicted `lastDecisionDate` from coverage **on the same render**.

`Decided()` is keyed on `beslutsdag` being non-empty. `beslutad` is deliberately not decoded:
it arrives with an inconsistent JSON type. The guard is mutation-tested — dropping
`beslutsdag` from the decode fails
`TestFetchRecentBetankanden_DistinguishesPlannedFromDecided`.

## Environment traps that cost real time

- **`npx vite --port N` binds `[::1]` only.** `curl localhost` picks IPv6 and works; a
  browser resolving `localhost` to IPv4 gets a degenerate empty response that looks exactly
  like the feature being broken. Use `--host 127.0.0.1`. This cost ~20 minutes.
- **The browser caches a legitimate empty API response** (`content-length: 2`, no cache
  headers) and keeps serving it after the data is fixed. Send `cache-control: no-store` from
  any stub proxy, and close the page between runs.
- **Riksdagen's `dokumentlista` is intermittently empty** for an identical query — 0 hits,
  then 40 seconds later, and once for several minutes straight. Never treat an empty
  response as a statement about the Riksdag, and expect the empty path to fire during
  testing whether you want it to or not.
- **`pkill -f` / `pgrep -f` match your own shell** and kill it (exit 144). Find servers with
  `ss -ltnp | grep -E ':(5199|8099)'` and `kill` the exact PID.
- **`gh pr merge` prints errors after succeeding** — it fails updating the local checkout
  when the tree is dirty (`.claude/settings.local.json` always is). Verify with
  `gh pr view N --json state`, then `git checkout main && git merge --ff-only origin/main`.
- **`.playwright-mcp/` is in `.gitignore` but its files are tracked.** gitignore does not
  apply to tracked files — `rm -rf` there deletes ~110 committed artifacts. I did this and
  restored with `git restore`. Run `git ls-files` before deleting anything.
- **The monthly spend limit has terminated three subagents mid-task.** Each died before
  writing anything, so nothing was corrupted, but plan for it: check `git status` after any
  agent failure, and be ready to implement in the controller session.

## Deferred minors (Phase 3)

- `speechesApi.listRecent` is now unused — its only caller was the hook Task 3 removed. Left
  deliberately; an api module is a catalogue of the backend's endpoints.
- Task 2's browser pass never visited `/committees/:code`; the case was proven correct by
  simulating `isActive` against the shipped arrays instead.
- Task 2's commit message justifies removing `/sok` partly with "every committee one click
  from the front page", which only became true when Task 3 landed.

Phase 2's deferred minors are committed in the Phase 3 plan under "Carried in from Phase 2",
along with the two-grains divergence that phase left for this one to decide.

## The judgement that governs this codebase

This site's failure mode is not crashing, it is **publishing something that looks
authoritative and is wrong**. Every serious defect found across these three phases passed
its tests and its gates:

- a committee vote count 27% low
- a coverage denominator frozen while its numerator moved
- an alignment percentage that quietly favoured the governing bloc
- five hand-picked "priorities" with a progress status we assigned
- a transient 500 rendering as "we hold no goals" for a committee that has nine
- **planned betänkanden for a future riksmöte published as the most recent decisions**

Two habits keep earning their cost. **Mutation-test every guard** — break the code on
purpose and watch the test fail, because this repo has repeatedly produced tests that could
not catch their own defect. And **verify against the live system, not the document that
asserts it** — six defects in Phase 2's plan and three in Phase 3's were found that way,
including two sets of URLs and one 27% undercount that had already shipped once.
