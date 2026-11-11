-- Riksdagen's own listing row for each betänkande (decided or planned), kept
-- for the Beslut list. Separate from `betankanden` on purpose: that table means
-- "förslagspunkter read" and the fetch-betankanden worker keys on it.
-- status, beslutsdag and beslutad are Riksdagen's, refreshed on every sync;
-- whether a row is decided is computed from them by RiksdagDocument.Decided,
-- never stored. New table only; nothing existing is touched.
CREATE TABLE IF NOT EXISTS beslut_listing (
    riksmote    TEXT NOT NULL,
    beteckning  TEXT NOT NULL,
    dok_id      TEXT NOT NULL,
    title       TEXT NOT NULL DEFAULT '',
    organ       TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT '',
    beslutsdag  DATE,
    beslutad    BOOLEAN,
    -- beslutsdag when Riksdagen gives one, else the publication date.
    sort_date   DATE NOT NULL,
    synced_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (riksmote, beteckning)
);
CREATE INDEX IF NOT EXISTS idx_beslut_listing_sort ON beslut_listing (sort_date DESC, riksmote DESC, beteckning DESC);
