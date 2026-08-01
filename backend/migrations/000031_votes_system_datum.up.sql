-- Riksdagen's own timestamp for each vote record.
--
-- Until now the votes table held only created_at, our insert time. That says
-- nothing about when a decision was taken, so any "senaste beslut" built from
-- it would present an ingest date as a decision date -- exactly the kind of
-- quiet inaccuracy this work exists to remove.
--
-- It is also the only honest basis for the ingestion cursor and for the
-- last_vote_date recorded in ingestion_coverage.

ALTER TABLE votes ADD COLUMN IF NOT EXISTS system_datum TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_votes_system_datum ON votes (system_datum);
