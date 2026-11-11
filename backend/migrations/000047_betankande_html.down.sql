DROP TABLE IF EXISTS betankande_statements;
ALTER TABLE forslagspunkter DROP COLUMN IF EXISTS reasoning;
ALTER TABLE betankanden
    DROP COLUMN IF EXISTS html_has_reservations,
    DROP COLUMN IF EXISTS html_has_reasoning,
    DROP COLUMN IF EXISTS html_fetched_at,
    DROP COLUMN IF EXISTS html_attempted_at;
