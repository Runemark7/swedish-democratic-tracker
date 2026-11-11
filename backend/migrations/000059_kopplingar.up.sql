-- Kopplingar between a Löfte and a votering, and the Veckobatch they are put
-- to the Redaktör in (issue #159, spec #129).
--
-- ADDITIVE ONLY. No existing table is touched and nothing is inserted.
--
-- Visibility: only 'bekraftad' kopplingar may ever be public. This issue adds
-- no public read of this table at all; föreslagen, avvisad and bortsorterad
-- rows are for the bot and the Redaktör only.

-- Created here so "new since the last Veckobatch" has something to refer to.
-- Creating and closing batches is #160.
CREATE TABLE veckobatch (
  id         SERIAL PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  closed_at  TIMESTAMPTZ
);

CREATE TABLE kopplingar (
  id            SERIAL PRIMARY KEY,
  loefte_id     INTEGER NOT NULL REFERENCES loften (id),
  -- voteringar.id, not beteckning + förslagspunkt: two voteringar can share those.
  votering_ref  INTEGER NOT NULL REFERENCES voteringar (id),
  -- Which ballot is in the Löfte's direction.
  direction     TEXT NOT NULL CHECK (direction IN ('Ja', 'Nej')),
  state         TEXT NOT NULL CHECK (state IN ('foreslagen', 'bekraftad', 'avvisad', 'bortsorterad')),
  confidence    NUMERIC(4,3) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
  threshold     NUMERIC(4,3) NOT NULL CHECK (threshold >= 0 AND threshold <= 1),
  reasoning     TEXT NOT NULL,
  origin        TEXT NOT NULL,
  veckobatch_id INTEGER REFERENCES veckobatch (id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  decided_at    TIMESTAMPTZ,
  UNIQUE (loefte_id, votering_ref)
);
CREATE INDEX idx_kopplingar_state ON kopplingar (state, confidence DESC);
CREATE INDEX idx_kopplingar_votering ON kopplingar (votering_ref);
