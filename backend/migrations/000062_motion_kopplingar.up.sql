-- Curated motion links through the Veckobatch (issue #163, spec #129).
--
-- A motion koppling is either
--   kind 'fraga': a fråga/interpellation that cites a motion only in prose, or
--   kind 'krav':  two yrkanden whose wording is not identical but that are the
--                 same Återkommande krav.
-- The bot suggests them; they become public only when the Redaktör confirms
-- ('bekraftad'). 'foreslagen', 'avvisad' and 'bortsorterad' rows are never
-- public. The rule recorded with the link is one of the two rules below.
--
-- ADDITIVE ONLY. No existing table loses data; nothing is inserted.
CREATE TABLE motion_kopplingar (
  id                SERIAL PRIMARY KEY,
  kind              TEXT NOT NULL CHECK (kind IN ('fraga', 'krav')),
  rule              TEXT NOT NULL CHECK (rule IN ('prose-citation', 'similar-wording')),
  state             TEXT NOT NULL CHECK (state IN ('foreslagen', 'bekraftad', 'avvisad', 'bortsorterad')),
  -- kind 'fraga': the question and the motion it cites.
  question_dok_id   TEXT REFERENCES questions (dok_id),
  motion_dok_id     TEXT NOT NULL REFERENCES motions (dok_id),
  -- kind 'krav': yrkande (motion_dok_id, yrkande_nr) and yrkande (other_*).
  -- No FK on the yrkande rows: a re-read of a motion replaces them.
  yrkande_nr        INTEGER,
  other_motion_dok_id TEXT REFERENCES motions (dok_id),
  other_yrkande_nr  INTEGER,
  confidence        NUMERIC(4,3) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
  threshold         NUMERIC(4,3) NOT NULL CHECK (threshold >= 0 AND threshold <= 1),
  reasoning         TEXT NOT NULL,
  origin            TEXT NOT NULL,
  veckobatch_id     INTEGER REFERENCES veckobatch (id),
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  decided_at        TIMESTAMPTZ,
  CONSTRAINT motion_kopplingar_shape CHECK (
    (kind = 'fraga' AND rule = 'prose-citation' AND question_dok_id IS NOT NULL
       AND yrkande_nr IS NULL AND other_motion_dok_id IS NULL AND other_yrkande_nr IS NULL)
    OR
    (kind = 'krav' AND rule = 'similar-wording' AND question_dok_id IS NULL
       AND yrkande_nr IS NOT NULL AND other_motion_dok_id IS NOT NULL AND other_yrkande_nr IS NOT NULL)
  )
);
CREATE UNIQUE INDEX motion_kopplingar_fraga_uniq
  ON motion_kopplingar (question_dok_id, motion_dok_id) WHERE kind = 'fraga';
CREATE UNIQUE INDEX motion_kopplingar_krav_uniq
  ON motion_kopplingar (motion_dok_id, yrkande_nr, other_motion_dok_id, other_yrkande_nr) WHERE kind = 'krav';
CREATE INDEX idx_motion_kopplingar_state ON motion_kopplingar (state, confidence DESC);
CREATE INDEX idx_motion_kopplingar_public ON motion_kopplingar (motion_dok_id) WHERE state = 'bekraftad';

-- A Veckobatch item is a Löfte koppling OR a motion koppling.
ALTER TABLE veckobatch_items ALTER COLUMN koppling_id DROP NOT NULL;
ALTER TABLE veckobatch_items ADD COLUMN motion_koppling_id INTEGER REFERENCES motion_kopplingar (id);
ALTER TABLE veckobatch_items ADD CONSTRAINT veckobatch_items_one_koppling
  CHECK ((koppling_id IS NOT NULL) <> (motion_koppling_id IS NOT NULL));
CREATE UNIQUE INDEX veckobatch_items_motion_uniq ON veckobatch_items (veckobatch_id, motion_koppling_id)
  WHERE motion_koppling_id IS NOT NULL;
CREATE INDEX idx_veckobatch_items_motion_koppling ON veckobatch_items (motion_koppling_id);
