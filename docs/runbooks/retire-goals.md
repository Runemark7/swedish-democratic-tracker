# Retire goals, scorecards and old promises (#164, migration 000064)

Migration `000064_retire_goals` drops the goals/promises feature that Löften
(#156) replaces. It touches nothing else: not `loften`, `valmanifest`,
`kopplingar`, `veckobatch*` or `aendringslogg`.

## Exactly what is dropped

| Kind | Name |
|---|---|
| Materialized view | `party_scorecards` (indexes `idx_party_scorecards_pk`, `idx_party_scorecards_party`) |
| Table | `goal_vote_matches` |
| Table | `promise_vote_matches` |
| Table | `promises` |
| Table | `party_goals` |
| Sequence | `goal_vote_matches_id_seq`, `promise_vote_matches_id_seq`, `promises_id_seq`, `party_goals_id_seq` (owned by the tables) |

The 2022 goals (72 rows) plus the one undated goal are already archived Löften
with the same id (migration 000057). The **seven 2026 goals exist only in
`party_goals`**, so the dump below is their only copy.

## Guard

The migration aborts (`retire goals: N party_goals rows have no archived
loften`) if any non-2026 `party_goals` row has no `loften` row with the same id
and `visibility = 'archived'`. Nothing is dropped in that case.

## Before deploy: take the dump

```bash
pg_dump "$DATABASE_URL" --format=custom --file=retire-goals-$(date +%F).dump \
  -t party_goals -t goal_vote_matches -t promises -t promise_vote_matches \
  -t party_goals_id_seq -t goal_vote_matches_id_seq \
  -t promises_id_seq -t promise_vote_matches_id_seq
```

`pg_dump` major version must match the server (18). Keep the file off the VPS
too. The view holds no data of its own and is rebuilt by the down migration.

## Rolling back / restoring

1. Recreate the schema: `migrate -path backend/migrations -database "$DATABASE_URL" down 1`
   (applies `000064_retire_goals.down.sql`: empty tables, indexes, constraints,
   sequences and the `party_scorecards` view).
2. Load the rows, data only, in dependency order:
   ```bash
   pg_restore --data-only --disable-triggers -d "$DATABASE_URL" \
     -t party_goals -t goal_vote_matches -t promises -t promise_vote_matches \
     -t party_goals_id_seq -t goal_vote_matches_id_seq \
     -t promises_id_seq -t promise_vote_matches_id_seq \
     retire-goals-YYYY-MM-DD.dump
   ```
   `--data-only` also restores the sequence positions (`setval`).
3. `REFRESH MATERIALIZED VIEW party_scorecards;`
4. Check: `SELECT count(*) FROM party_goals;` is 80 (as of 2026-11), and the
   `promises` and `goal_vote_matches` counts match the pre-deploy numbers.

Rolling back only the database does not bring the code back; deploy the
previous image as well.

## Removed API

`GET /api/parties`, `/api/parties/{party}/goals`,
`/api/parties/{party}/goals/{goalId}/votes`, `/api/committees/{code}/goals` and
`/api/politicians/{id}/promises` answered `410 Gone` with a pointer to
`/api/loften` for one release. That shim (`backend/internal/api/retired.go`)
has been removed; the routes now answer plain 404. The daily chain no longer
runs `keyword-matcher` or `refresh-scorecards`.
