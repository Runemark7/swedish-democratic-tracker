ALTER TABLE betankanden
    DROP COLUMN IF EXISTS debate_checked_at,
    DROP COLUMN IF EXISTS debate_date;

DROP INDEX IF EXISTS idx_speeches_betankande;

ALTER TABLE speeches
    DROP COLUMN IF EXISTS betankande_linked_at,
    DROP COLUMN IF EXISTS betankande_link_rule,
    DROP COLUMN IF EXISTS betankande_dok_id;
