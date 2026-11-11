-- Skriftliga frågor (fr) with their svar (frs) and interpellationer (ip), plus
-- the motion links the rules allow. Additive only: new tables, no existing row
-- is touched. Rows are filled by the fetch-questions worker, which is
-- idempotent (keyed on dok_id and Riksdagen's systemdatum).

CREATE TABLE IF NOT EXISTS questions (
    dok_id             TEXT PRIMARY KEY,
    kind               TEXT NOT NULL CHECK (kind IN ('fraga', 'interpellation')),
    riksmote           TEXT NOT NULL,
    beteckning         TEXT NOT NULL,
    title              TEXT NOT NULL DEFAULT '',
    asked_date         DATE,
    status             TEXT NOT NULL DEFAULT '',
    systemdatum        TIMESTAMP NOT NULL,
    -- The member who asked (dokintressent roll undertecknare/fragestallare).
    asker_id           TEXT NOT NULL DEFAULT '',
    asker_name         TEXT NOT NULL DEFAULT '',
    asker_party        TEXT NOT NULL DEFAULT '',
    -- The minister asked (roll stalldtill).
    addressed_id       TEXT NOT NULL DEFAULT '',
    addressed_name     TEXT NOT NULL DEFAULT '',
    addressed_party    TEXT NOT NULL DEFAULT '',
    -- The minister who answered (roll besvaradav); empty until answered.
    answerer_id        TEXT NOT NULL DEFAULT '',
    answerer_name      TEXT NOT NULL DEFAULT '',
    answerer_party     TEXT NOT NULL DEFAULT '',
    -- The svar document (frs) for a fråga; empty for an interpellation.
    answer_dok_id      TEXT NOT NULL DEFAULT '',
    answer_date        DATE,
    -- Plain text of the question, kept so links can be recomputed without
    -- another request.
    body_text          TEXT NOT NULL DEFAULT '',
    -- Motions Riksdagen's own references point at, as 'riksmöte:beteckning'.
    motion_refs        TEXT[] NOT NULL DEFAULT '{}',
    fetched_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (kind, riksmote, beteckning)
);

CREATE INDEX IF NOT EXISTS idx_questions_asker ON questions (asker_id) WHERE asker_id <> '';
CREATE INDEX IF NOT EXISTS idx_questions_answerer ON questions (answerer_id) WHERE answerer_id <> '';

-- Only links that pass a rule are stored. The rule is stored with the link.
CREATE TABLE IF NOT EXISTS question_motion_links (
    question_dok_id TEXT NOT NULL REFERENCES questions (dok_id) ON DELETE CASCADE,
    motion_dok_id   TEXT NOT NULL REFERENCES motions (dok_id) ON DELETE CASCADE,
    rule            TEXT NOT NULL CHECK (rule IN ('riksdagen-reference', 'cited-beteckning')),
    PRIMARY KEY (question_dok_id, motion_dok_id)
);

CREATE INDEX IF NOT EXISTS idx_question_links_motion ON question_motion_links (motion_dok_id);

-- What Riksdagen's listing says exists, per riksmöte and kind; held is counted.
CREATE TABLE IF NOT EXISTS question_coverage (
    riksmote           TEXT NOT NULL,
    kind               TEXT NOT NULL CHECK (kind IN ('fraga', 'interpellation')),
    expected           INT NOT NULL,
    last_enumerated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (riksmote, kind)
);
