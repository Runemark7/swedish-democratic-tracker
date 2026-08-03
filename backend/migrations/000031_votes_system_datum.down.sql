DROP INDEX IF EXISTS idx_votes_system_datum;
ALTER TABLE votes DROP COLUMN IF EXISTS system_datum;
