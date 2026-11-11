-- The parts of a betänkande that exist only in its HTML: reservationer,
-- särskilda yttranden and "Utskottets ställningstagande". Additive only:
-- nullable/defaulted columns and one new table. No existing row is rewritten.

ALTER TABLE betankanden
    ADD COLUMN IF NOT EXISTS html_attempted_at   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS html_fetched_at     TIMESTAMPTZ,
    -- What the parser recognised, so "no reservations" can be told from
    -- "could not read the reservations".
    ADD COLUMN IF NOT EXISTS html_has_reasoning    BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS html_has_reservations BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE forslagspunkter
    ADD COLUMN IF NOT EXISTS reasoning TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS betankande_statements (
    riksmote   TEXT NOT NULL,
    beteckning TEXT NOT NULL,
    kind       TEXT NOT NULL CHECK (kind IN ('reservation', 'sarskilt_yttrande')),
    nummer     INT  NOT NULL,
    rubrik     TEXT NOT NULL DEFAULT '',
    punkt      TEXT NOT NULL DEFAULT '',
    parties    TEXT[] NOT NULL DEFAULT '{}',
    signers    TEXT[] NOT NULL DEFAULT '{}',
    proposal   TEXT NOT NULL DEFAULT '',
    body       TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (riksmote, beteckning, kind, nummer),
    FOREIGN KEY (riksmote, beteckning) REFERENCES betankanden (riksmote, beteckning) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_betankande_statements_punkt ON betankande_statements (riksmote, beteckning, punkt);
