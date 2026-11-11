-- Drops the backfill's own progress table, index and column. Speech rows stay.
DROP TABLE IF EXISTS speech_backfill;
DROP INDEX IF EXISTS idx_speeches_riksmote;
ALTER TABLE speeches DROP COLUMN IF EXISTS riksmote;
