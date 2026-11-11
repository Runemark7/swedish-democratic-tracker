# Runbook: backfill motions and betänkanden (issue #144)

The daily chain (`fetch-motions`, `fetch-betankanden`) takes only the **current
riksmöte**. Everything older is this operator command. It never runs on deploy
or at startup; run it deliberately, against the database you mean to fill.

```bash
# from backend/, DATABASE_URL set
go run ./cmd/backfill-motions                 # every riksmöte with voteringar, newest first, back to 2002/03
go run ./cmd/backfill-motions -rm 2021/22     # one or more riksmöten (comma separated)
go run ./cmd/backfill-motions -from 2014/15   # stop earlier
go run ./cmd/backfill-motions -sources motioner   # or: betankanden
go run ./cmd/backfill-motions -report         # print the report from stored data; fetches nothing
```

In the image the binary is `/backfill-motions`.

## Behaviour

- One riksmöte at a time, newest first. For each, motions first, then the
  betänkanden the riksmöte's motions point at (förslagspunkter, reservations,
  yrkande links).
- Progress is written to `motion_backfill` per riksmöte and source. A riksmöte
  × source is **complete** when every motion Riksdagen's listing reports is
  held (motioner), or no betänkande its motions point at is left to read
  (betankanden). A restart skips complete rows; an interrupted or failed one is
  redone, skipping every motion already stored. `-force` redoes complete rows.
- Idempotent: every write is the daily worker's own upsert.
- Polite: `-delay` (default 200 ms) between requests, as the workers.
- Ctrl-C stops cleanly; the exit code is non-zero when the run was interrupted
  or a step failed, so a partial run is never mistaken for a complete one.
- After the walk it recomputes the yrkande links of every held betänkande (no
  requests). A votering's "proposed by" is derived from those links when read
  (#141), so this is the origin recompute. `-skip-relink` leaves it out; it can
  be repeated any time by running with `-sources betankanden -rm <rm>`.

## The report

Printed at the end (and by `-report`):

1. Coverage per riksmöte and source: expected, held, complete.
2. Every chamber decision value seen, with counts: `kammaren` (decision on a
   yrkande), `punktbeslut` (decision read from the punkt), `beslutstyp`
   (acklamation / röstning / …) and `vinnare`.
3. Values `Outcome` does not map, marked `EJ MAPPAT`. The site already shows
   such a word exactly as Riksdagen wrote it. To map one, add it to
   `mapOutcomeWord` in `backend/internal/motions/domain/outcome.go` with a test.
4. How many voteringar have at least one decided yrkande linked to their
   förslagspunkt.

The Data section (`/data`) shows the riksmöten covered, from
`GET /api/motions/backfill`.
