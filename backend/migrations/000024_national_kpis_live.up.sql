-- National riksdag KPIs become live-sourced from SCB via the ingestion
-- scheduler (national-kpis worker) instead of hardcoded snapshots.
--
-- 1. Remove the four hand-entered rows from migration 000012: their raw
--    values, deltas and trends were static guesses presented next to source
--    links, and two of them (Statsskuld/BNP, BNP-tillväxt) have no clean
--    machine-readable source at all.
-- 2. Drop the target column: the thresholds were editorial, not sourced
--    (same rationale as removing region/kommun KPI targets).
-- 3. Add a uniqueness key so the worker can upsert one row per (label, year).

DELETE FROM riksdag_kpis;

ALTER TABLE riksdag_kpis DROP COLUMN IF EXISTS target;

-- Idempotent: migration 000012 already creates an inline UNIQUE (label, year)
-- whose auto-generated name is riksdag_kpis_label_year_key, so adding it again
-- collides on a fresh database. Drop any existing one first, then (re)add.
ALTER TABLE riksdag_kpis DROP CONSTRAINT IF EXISTS riksdag_kpis_label_year_key;

ALTER TABLE riksdag_kpis
  ADD CONSTRAINT riksdag_kpis_label_year_key UNIQUE (label, year);
