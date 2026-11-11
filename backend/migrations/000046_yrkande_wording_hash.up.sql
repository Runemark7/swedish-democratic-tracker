-- Återkommande krav, Phase 1: yrkanden whose normalised wording is identical
-- share a hash. The hash is written on ingest. Rows stored before this
-- migration stay NULL until they are read or the opt-in backfill-wording-hash
-- command fills them; nothing here rewrites existing data.
ALTER TABLE motion_yrkanden ADD COLUMN IF NOT EXISTS wording_hash TEXT;

CREATE INDEX IF NOT EXISTS motion_yrkanden_wording_hash_idx
    ON motion_yrkanden (wording_hash) WHERE wording_hash IS NOT NULL;
