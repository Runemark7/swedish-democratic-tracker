-- Betänkanden the motions are handled in, with their förslagspunkter, and the
-- link from each yrkande to the punkt(er) that decided it. New tables only.

-- One row per (riksmöte, beteckning) we have looked up. found=false records a
-- lookup that returned nothing, so the worker does not retry it every run.
CREATE TABLE IF NOT EXISTS betankanden (
    riksmote    TEXT NOT NULL,
    beteckning  TEXT NOT NULL,
    dok_id      TEXT NOT NULL DEFAULT '',
    found       BOOLEAN NOT NULL DEFAULT false,
    title       TEXT NOT NULL DEFAULT '',
    organ       TEXT NOT NULL DEFAULT '',
    systemdatum TIMESTAMP,
    fetched_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (riksmote, beteckning)
);

CREATE TABLE IF NOT EXISTS forslagspunkter (
    riksmote       TEXT NOT NULL,
    beteckning     TEXT NOT NULL,
    punkt          TEXT NOT NULL,
    rubrik         TEXT NOT NULL DEFAULT '',
    forslag_text   TEXT NOT NULL DEFAULT '',
    beslutstyp     TEXT NOT NULL DEFAULT '',
    votering_id    TEXT NOT NULL DEFAULT '',
    -- "Utskottets förslag mot reservation 12 (M)": the alternative put to the vote.
    alternativ     TEXT NOT NULL DEFAULT '',
    vinnare        TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (riksmote, beteckning, punkt),
    FOREIGN KEY (riksmote, beteckning) REFERENCES betankanden (riksmote, beteckning) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_forslagspunkter_votering ON forslagspunkter (upper(votering_id)) WHERE votering_id <> '';

-- A yrkande is identified by (motion_dok_id, nummer). No foreign key to
-- motion_yrkanden: a motion re-read replaces its yrkande rows, and the link
-- must survive that.
CREATE TABLE IF NOT EXISTS yrkande_forslagspunkt (
    motion_dok_id   TEXT NOT NULL,
    yrkande_nummer  INT  NOT NULL,
    riksmote        TEXT NOT NULL,
    beteckning      TEXT NOT NULL,
    punkt           TEXT NOT NULL,
    linked_by       TEXT NOT NULL CHECK (linked_by IN ('behandlas_i_punkt', 'forslag_text')),
    decision        TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (motion_dok_id, yrkande_nummer, riksmote, beteckning, punkt),
    FOREIGN KEY (riksmote, beteckning, punkt) REFERENCES forslagspunkter (riksmote, beteckning, punkt) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_yfp_motion ON yrkande_forslagspunkt (motion_dok_id);
