# Runbook: dirty database migration

## Symptom

The backend pod crash-loops or serves a stale schema, and the logs show:

```
ERROR migration failed error="Dirty database version N. Fix and force version."
```

golang-migrate sets `dirty = true` in the `schema_migrations` table when a
migration starts but does not finish cleanly. Once dirty, the migrator refuses
to run anything — including the migration that failed — until a human resolves
the state. Restarting the pod does not help; every boot logs the same error.

## Why this happens

Each file in `backend/migrations/` runs as one transaction, so a failed
migration normally rolls back completely and only the flag is left set. The
flag can also be left behind by a crash mid-migration (OOM-kill, node
eviction) or by concurrent migrators fighting over partially-shared state
(should not occur in prod: `replicas: 1` and golang-migrate takes a Postgres
advisory lock, but local dev with repeated `docker compose up --build` churn
can get there).

## Fix procedure

Work against the affected database (`kubectl exec` into the postgres pod, or
`docker exec` locally). `N` below is the version from the error message.

1. **Read the failed migration** — open
   `backend/migrations/00000N_*.up.sql` and note every statement it runs.

2. **Inspect what actually applied.** For each statement, check the live
   schema/data:

   ```sql
   \d <table>                 -- columns, constraints, indexes
   SELECT ... LIMIT 5;        -- affected rows
   ```

3. **Restore the pre-migration state.** Manually revert any statement that
   did apply, so the database matches what version `N-1` left behind. Typical
   reversals: drop a constraint/index the migration added, re-add a column it
   dropped (the `*.down.sql` file shows the inverse operations — but apply
   them selectively, only for the parts that actually ran).

4. **Clear the dirty flag and step the version back:**

   ```sql
   UPDATE schema_migrations SET version = N - 1, dirty = false;
   ```

5. **Restart the backend pod.** Migration `N` re-runs from the top against a
   clean pre-`N` state and should apply in one pass:

   ```bash
   kubectl rollout restart deployment/<backend>   # prod
   docker restart <backend-container>             # local dev
   ```

6. **Verify:** logs show `migrations applied`, and
   `SELECT version, dirty FROM schema_migrations;` reads `N | f`.

## What NOT to do

- **Do not** just flip `dirty = false` while leaving `version = N`. That
  tells the migrator migration `N` completed — if it only partially applied,
  the schema is silently wrong and every later migration builds on the wrong
  base.
- **Do not** edit an already-released migration file to "make it pass".
  Released migrations are immutable; databases that already ran the original
  will diverge.
- **Do not** run statements from the migration by hand and then force the
  version forward, unless you have verified every statement, in order —
  re-running the file from a restored state (steps 3–5) is safer.

## Real example (2026-06-10, dev)

Migration `000024_national_kpis_live` wedged on a dev database: the unique
constraint already existed from earlier container churn, so the re-run failed
with `relation "riksdag_kpis_label_year_key" already exists` and left
`version = 24, dirty = true`. Resolution per this runbook:

```sql
ALTER TABLE riksdag_kpis DROP CONSTRAINT IF EXISTS riksdag_kpis_label_year_key;
UPDATE schema_migrations SET version = 23, dirty = false;
```

then restarted the backend — migration 24 applied cleanly in one pass.
