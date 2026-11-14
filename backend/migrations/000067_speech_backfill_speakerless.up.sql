-- Speeches data.riksdagen.se lists with the placeholder speaker id
-- "0000000000000" (no person, e.g. a ceremony speech by an acting oldest
-- member) cannot be stored: speeches.politician_id is NOT NULL and references
-- politicians. The backfill counts them here instead, so a riksmöte can be
-- complete without hiding that some speeches were left out.
-- Additive: a constant default, so no table rewrite and no existing row touched.
ALTER TABLE speech_backfill ADD COLUMN IF NOT EXISTS speakerless_speeches INT NOT NULL DEFAULT 0;
