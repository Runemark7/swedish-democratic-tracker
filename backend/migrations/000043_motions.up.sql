-- Motions from data.riksdagen.se (doktyp=mot). Keyed on dok_id; (riksmote,
-- beteckning) is the public identity. subtyp is stored as Riksdagen gave it;
-- the kind (Flerpartimotion included) is derived on read.
CREATE TABLE IF NOT EXISTS motions (
    dok_id        TEXT PRIMARY KEY,
    riksmote      TEXT NOT NULL,
    beteckning    TEXT NOT NULL,
    title         TEXT NOT NULL DEFAULT '',
    subtyp        TEXT NOT NULL DEFAULT '',
    motion_date   DATE,
    organ         TEXT NOT NULL DEFAULT '',
    tilldelat     TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT '',
    systemdatum   TIMESTAMP NOT NULL,
    behandlas_i   TEXT NOT NULL DEFAULT '',
    fetched_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (riksmote, beteckning)
);

CREATE TABLE IF NOT EXISTS motion_authors (
    dok_id        TEXT NOT NULL REFERENCES motions(dok_id) ON DELETE CASCADE,
    ordning       INT  NOT NULL,
    intressent_id TEXT NOT NULL DEFAULT '',
    name          TEXT NOT NULL DEFAULT '',
    party         TEXT NOT NULL DEFAULT '',
    role          TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (dok_id, ordning)
);

CREATE TABLE IF NOT EXISTS motion_yrkanden (
    dok_id            TEXT NOT NULL REFERENCES motions(dok_id) ON DELETE CASCADE,
    nummer            INT  NOT NULL,
    wording           TEXT NOT NULL DEFAULT '',
    utskottet         TEXT NOT NULL DEFAULT '',
    kammaren          TEXT NOT NULL DEFAULT '',
    behandlas_i       TEXT NOT NULL DEFAULT '',
    behandlas_i_punkt TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (dok_id, nummer)
);

-- What Riksdagen's listing says exists, per riksmöte and source. The number we
-- hold is counted from motions at read time; it is not stored here.
CREATE TABLE IF NOT EXISTS motion_coverage (
    riksmote           TEXT NOT NULL,
    source             TEXT NOT NULL,
    expected           INT  NOT NULL,
    last_enumerated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (riksmote, source)
);
