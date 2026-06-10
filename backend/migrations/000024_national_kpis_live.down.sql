ALTER TABLE riksdag_kpis DROP CONSTRAINT IF EXISTS riksdag_kpis_label_year_key;

ALTER TABLE riksdag_kpis ADD COLUMN target NUMERIC NOT NULL DEFAULT 0;

-- The hardcoded seed rows from 000012 are intentionally not restored: they
-- were unsourced snapshots. After rolling back, the KPI strip is empty until
-- re-seeded manually.
