-- Restores the columns, not the five curated rows: those were our paraphrase of
-- selected points and are deliberately not reconstructed.
ALTER TABLE riksdag_agenda ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE riksdag_agenda ADD COLUMN IF NOT EXISTS status      TEXT NOT NULL DEFAULT 'active';

ALTER TABLE riksdag_agenda DROP COLUMN IF EXISTS published;
ALTER TABLE riksdag_agenda DROP COLUMN IF EXISTS issuer;
ALTER TABLE riksdag_agenda DROP COLUMN IF EXISTS url;
