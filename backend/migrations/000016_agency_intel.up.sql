CREATE TABLE agency_regleringsbrev (
    id          SERIAL PRIMARY KEY,
    agency_slug TEXT NOT NULL,
    year        INT  NOT NULL,
    dok_id      TEXT NOT NULL,
    date        DATE NOT NULL,
    title       TEXT NOT NULL,
    summary     TEXT NOT NULL DEFAULT '',
    url         TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (agency_slug, year)
);

CREATE TABLE agency_decisions (
    id          SERIAL PRIMARY KEY,
    agency_slug TEXT NOT NULL,
    dok_id      TEXT NOT NULL,
    date        DATE NOT NULL,
    title       TEXT NOT NULL,
    doc_type    TEXT NOT NULL,
    summary     TEXT NOT NULL DEFAULT '',
    url         TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (agency_slug, dok_id)
);

CREATE INDEX ON agency_regleringsbrev (agency_slug);
CREATE INDEX ON agency_decisions (agency_slug);
