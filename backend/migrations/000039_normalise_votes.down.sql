-- Rebuilds the denormalised votes table from voteringar + ballots.
--
-- The reconstruction is faithful except for `id`: the original per-ballot
-- SERIAL values are gone, so rows get fresh ones. Nothing referenced them --
-- promise_vote_matches.vote_id was the only foreign key and its table was
-- empty -- so the renumbering is not observable, but it does mean this is a
-- rebuild rather than a restore.
DROP VIEW IF EXISTS votes;

CREATE TABLE votes (
  id               SERIAL PRIMARY KEY,
  votering_id      TEXT NOT NULL,
  politician_id    TEXT NOT NULL REFERENCES politicians (intressent_id) ON DELETE CASCADE,
  party            TEXT NOT NULL,
  vote_result      TEXT NOT NULL CHECK (vote_result IN ('Ja', 'Nej', 'Avstår', 'Frånvarande')),
  beteckning       TEXT NOT NULL,
  forslagspunkt    TEXT NOT NULL DEFAULT '',
  session          TEXT NOT NULL,
  dok_id           TEXT,
  system_datum     TIMESTAMPTZ,
  proposed_by_party TEXT,
  proposal_type     TEXT CHECK (proposal_type IN ('prop', 'mot', 'bet')),
  proposal_dok_id   TEXT,
  document_title    TEXT,
  origin_enriched   BOOLEAN NOT NULL DEFAULT false,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (votering_id, politician_id)
);

INSERT INTO votes (
  votering_id, politician_id, party, vote_result, beteckning, forslagspunkt,
  session, dok_id, system_datum, proposed_by_party, proposal_type,
  proposal_dok_id, document_title, origin_enriched, created_at)
SELECT
  vg.votering_id, b.politician_id, b.party, b.vote_result, vg.beteckning,
  vg.forslagspunkt, vg.session, vg.dok_id, vg.system_datum,
  vg.proposed_by_party, vg.proposal_type, vg.proposal_dok_id,
  vg.document_title, vg.origin_enriched, vg.created_at
FROM ballots b
JOIN voteringar vg ON vg.id = b.votering_ref;

CREATE INDEX idx_votes_politician_id  ON votes (politician_id);
CREATE INDEX idx_votes_beteckning     ON votes (beteckning, forslagspunkt);
CREATE INDEX idx_votes_session        ON votes (session);
CREATE INDEX idx_votes_party          ON votes (party);
CREATE INDEX idx_votes_unenriched     ON votes (id) WHERE origin_enriched = false;
CREATE INDEX idx_votes_system_datum   ON votes (system_datum);

ALTER TABLE promise_vote_matches
  ADD COLUMN IF NOT EXISTS vote_id INTEGER REFERENCES votes (id) ON DELETE CASCADE;

DROP TABLE ballots;
DROP TABLE voteringar;
