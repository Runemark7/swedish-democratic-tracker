-- Veckobatch items: the stable numbering of a Veckobatch and the Redaktör's
-- answer to each numbered item (issue #160, spec #129).
--
-- ADDITIVE ONLY. No existing table is touched and nothing is inserted.
-- A koppling that is not answered stays 'foreslagen' and is simply put in the
-- next batch with a new number there (rollover); the old row keeps its number.
CREATE TABLE veckobatch_items (
  veckobatch_id INTEGER NOT NULL REFERENCES veckobatch (id),
  nr            INTEGER NOT NULL CHECK (nr > 0),
  koppling_id   INTEGER NOT NULL REFERENCES kopplingar (id),
  -- NULL = not answered yet. ja-omvaent = confirmed with the direction reversed.
  answer        TEXT CHECK (answer IN ('ja', 'nej', 'ja-omvaent')),
  answered_at   TIMESTAMPTZ,
  PRIMARY KEY (veckobatch_id, nr),
  UNIQUE (veckobatch_id, koppling_id)
);
CREATE INDEX idx_veckobatch_items_koppling ON veckobatch_items (koppling_id);
