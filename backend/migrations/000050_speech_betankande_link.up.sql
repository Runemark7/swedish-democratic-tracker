-- Link each anförande to the betänkande it debates, and keep the betänkande's
-- debate date for the title+date fallback. Additive only: nullable columns and
-- indexes. No existing row is rewritten; links are filled by the link-speeches
-- worker, which is idempotent and resumable.

ALTER TABLE speeches
    ADD COLUMN IF NOT EXISTS betankande_dok_id TEXT,
    -- The rule that made the link: 'rel_dok_id' or 'avsnittsrubrik_datum'.
    ADD COLUMN IF NOT EXISTS betankande_link_rule TEXT
        CHECK (betankande_link_rule IN ('rel_dok_id', 'avsnittsrubrik_datum')),
    ADD COLUMN IF NOT EXISTS betankande_linked_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_speeches_betankande ON speeches (upper(betankande_dok_id), anforande_nummer)
    WHERE betankande_dok_id IS NOT NULL;

ALTER TABLE betankanden
    ADD COLUMN IF NOT EXISTS debate_date DATE,
    -- Set when a read of the betänkande has looked for its debate date, so a
    -- betänkande stored before this column existed is read once more.
    ADD COLUMN IF NOT EXISTS debate_checked_at TIMESTAMPTZ;
