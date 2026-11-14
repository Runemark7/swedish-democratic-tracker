# Runbook: backfill speeches (issue #147)

The daily `fetch-speeches` worker takes the recent riksmöten. Older anföranden
are this operator command. It never runs on deploy or at startup; run it
deliberately, against the database you mean to fill.

```bash
# from backend/, DATABASE_URL set
go run ./cmd/backfill-speeches                    # 2014/15 up to the current riksmöte, newest first
go run ./cmd/backfill-speeches -rm 2017/18        # one or more riksmöten (comma separated)
go run ./cmd/backfill-speeches -from 2014/15 -to 2021/22
go run ./cmd/backfill-speeches -delay 500ms       # be politer
go run ./cmd/backfill-speeches -force             # redo riksmöten recorded as complete
```

In the image the binary is `/backfill-speeches`. #147 covers the two newest
mandate periods plus 2014–2018, which is the default range; older periods follow
in #148 (`-from 2002/03`), sized from the report below.

## Behaviour

- One riksmöte at a time, newest first. For each: the speech lists (one request
  per party: S M SD C V KD L MP FP), a `politicians` row for every speaker who
  has none (from `/personlista/?iid=`, former members as inactive), the
  speeches, the text of each (one request per speech, `-delay` apart), and the
  link to the betänkande each debates (the rules of #146).
- Progress is written to `speech_backfill` per riksmöte. A riksmöte is
  **complete** when every speech the lists report is stored with text and no
  speaker is unresolved. A restart skips complete riksmöten; an interrupted or
  failed one is redone, and only speeches still lacking text are fetched again.
- Idempotent: every write is an upsert; a re-listed speech keeps its text.
- Ctrl-C stops cleanly. The exit code is non-zero when the run was interrupted
  or a riksmöte is not complete.
- A riksmöte holds roughly 10–15 thousand speeches, so at 200 ms per text request
  expect about an hour per riksmöte. The text requests dominate.
- Speakers Riksdagen has no person for are listed (`OKÄNDA TALARE`) and their
  speeches are not stored; the riksmöte stays incomplete until that is resolved.
- Some speeches are listed with the placeholder speaker id `0000000000000` (no
  person; e.g. the ceremony speech of the acting oldest member at the end of
  2007/08). They cannot be stored (`speeches.politician_id` is required), so
  they are reported as `utan talare N` and `speakerlessSpeeches` on
  `/api/speeches/backfill`, and are not part of *förväntat*. They do not block
  completion; any other unresolved speaker still does.
- Run `link-speeches` (or wait for the daily run) after the motions backfill has
  read more betänkanden: the debate-date rule needs the betänkande stored.

## The report

Printed at the end:

1. Coverage per riksmöte: expected, stored, linked, unlinked, missing text, new
   politicians.
2. Storage growth per riksmöte: new rows, text bytes, and the growth of the
   `speeches` table (heap and TOAST, `pg_table_size`) and its indexes
   (`pg_indexes_size`), measured around the riksmöte. The total and the mean per
   riksmöte are what #148 is sized by. The table figure can include dead tuples
   from the text updates until autovacuum has run; the same figures are kept in
   `speech_backfill` and served by `GET /api/speeches/backfill`.

The Data section (`/data`) shows the coverage per riksmöte.
