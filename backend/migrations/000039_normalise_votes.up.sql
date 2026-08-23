-- Split votes into the votering and the ballots cast in it.
--
-- One votering is decided by ~350 members, and the old table repeated every
-- votering-level field on all 350 rows: document_title, votering_id, session,
-- beteckning, dok_id, the four proposal_* columns. Measured on the full
-- six-period record (5 526 415 ballots, 15 835 voteringar), that duplication
-- is ~620 MB of the 800 MB heap. Only politician_id, party and vote_result
-- actually vary per ballot -- about 106 MB.
--
-- votering_id is the identity, not (beteckning, forslagspunkt): 18
-- förslagspunkter in 2022-2026 were decided by two separate voteringar sharing
-- a beteckning and punkt. Verified on the full record -- votering_id and
-- (votering_id, beteckning, forslagspunkt, session) both count 15 835, so
-- votering_id alone is a sound key.
--
-- votes survives as a view over the join, so every existing read keeps working
-- while the queries are moved across one at a time.

CREATE TABLE voteringar (
  id                SERIAL PRIMARY KEY,
  votering_id       TEXT NOT NULL UNIQUE,
  beteckning        TEXT NOT NULL,
  forslagspunkt     TEXT NOT NULL DEFAULT '',
  session           TEXT NOT NULL,
  dok_id            TEXT,
  system_datum      TIMESTAMPTZ,
  -- Enriched proposal origin. Filled per vote point by enrich-vote-origins,
  -- which is why it belongs here rather than on every ballot.
  proposed_by_party TEXT,
  proposal_type     TEXT CHECK (proposal_type IN ('prop', 'mot', 'bet')),
  proposal_dok_id   TEXT,
  document_title    TEXT,
  origin_enriched   BOOLEAN NOT NULL DEFAULT false,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ballots (
  votering_ref  INTEGER NOT NULL REFERENCES voteringar (id) ON DELETE CASCADE,
  politician_id TEXT    NOT NULL REFERENCES politicians (intressent_id) ON DELETE CASCADE,
  party         TEXT    NOT NULL,
  vote_result   TEXT    NOT NULL CHECK (vote_result IN ('Ja', 'Nej', 'Avstår', 'Frånvarande')),
  -- Same grain the old UNIQUE (votering_id, politician_id) enforced: one
  -- member votes once per votering. Doubles as the lookup index, so no
  -- separate surrogate key is needed.
  PRIMARY KEY (votering_ref, politician_id)
);

-- Aggregated rather than DISTINCT ON: enrich-vote-origins updates a vote point
-- as a unit, so every ballot of a votering should already agree on these. Using
-- max()/bool_or() means that if any row was ever left behind, the enriched
-- value wins instead of an arbitrary one.
INSERT INTO voteringar (
  votering_id, beteckning, forslagspunkt, session, dok_id, system_datum,
  proposed_by_party, proposal_type, proposal_dok_id, document_title,
  origin_enriched, created_at)
SELECT
  votering_id,
  max(beteckning),
  max(forslagspunkt),
  max(session),
  max(dok_id),
  max(system_datum),
  max(proposed_by_party),
  max(proposal_type),
  max(proposal_dok_id),
  max(document_title),
  bool_or(origin_enriched),
  min(created_at)
FROM votes
GROUP BY votering_id;

INSERT INTO ballots (votering_ref, politician_id, party, vote_result)
SELECT vg.id, v.politician_id, v.party, v.vote_result
FROM votes v
JOIN voteringar vg ON vg.votering_id = v.votering_id;

-- promise_vote_matches.vote_id referenced votes(id), a per-ballot surrogate --
-- so a promise was matched to one member's ballot rather than to the votering.
-- The table and promises are both empty, so nothing is lost by dropping the
-- column now; a promise should point at a votering when that feature returns.
ALTER TABLE promise_vote_matches DROP COLUMN IF EXISTS vote_id;

DROP TABLE votes;

-- Compatibility view: the shape the old table had, minus `id`.
--
-- `id` was a per-ballot SERIAL that no caller ever used -- not in any query
-- filter, not in the frontend. Reproducing it here would mean either an
-- expensive row_number() over 5.5M rows or quietly redefining it as the
-- votering's id, so it is dropped from the contract instead.
CREATE VIEW votes AS
SELECT
  vg.votering_id,
  b.politician_id,
  b.party,
  b.vote_result,
  vg.beteckning,
  vg.forslagspunkt,
  vg.session,
  vg.dok_id,
  vg.system_datum,
  vg.proposed_by_party,
  vg.proposal_type,
  vg.proposal_dok_id,
  vg.document_title,
  vg.origin_enriched,
  vg.created_at
FROM ballots b
JOIN voteringar vg ON vg.id = b.votering_ref;

CREATE INDEX idx_voteringar_session     ON voteringar (session);
CREATE INDEX idx_voteringar_beteckning  ON voteringar (beteckning, forslagspunkt);
CREATE INDEX idx_voteringar_unenriched  ON voteringar (id) WHERE origin_enriched = false;
CREATE INDEX idx_voteringar_system_datum ON voteringar (system_datum);
-- The ballots PK already covers (votering_ref, politician_id). This is the
-- other direction: one member's whole record, which is what the politician
-- page reads.
CREATE INDEX idx_ballots_politician     ON ballots (politician_id);
