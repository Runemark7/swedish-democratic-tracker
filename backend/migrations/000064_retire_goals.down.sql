-- Recreates the schema dropped by 000064_retire_goals.up.sql: tables, indexes,
-- constraints, sequences and the party_scorecards materialized view, as they
-- stood after migration 000034. Rows are NOT restored here; restore
-- them from the pre-deploy pg_dump (docs/runbooks/retire-goals.md).

CREATE TABLE party_goals (
  id                  SERIAL PRIMARY KEY,
  party               TEXT NOT NULL,
  goal_text           TEXT NOT NULL,
  topic               TEXT NOT NULL,
  specificity         TEXT NOT NULL CHECK (specificity IN ('concrete', 'directional', 'rhetorical')),
  source_document     TEXT NOT NULL,
  keywords            TEXT[] NOT NULL DEFAULT '{}',
  relevant_committees TEXT[] NOT NULL DEFAULT '{}',
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  source_url          TEXT,
  source_quote        TEXT,
  election_cycle      TEXT,
  CONSTRAINT party_goals_party_fk FOREIGN KEY (party) REFERENCES parties (code),
  CONSTRAINT uq_party_goals_party_text UNIQUE (party, goal_text),
  CONSTRAINT party_goals_election_cycle_check CHECK (election_cycle IS NULL OR election_cycle IN ('2022', '2026'))
);
CREATE INDEX idx_party_goals_party ON party_goals (party);
CREATE INDEX idx_party_goals_topic ON party_goals (topic);
CREATE INDEX idx_party_goals_election_cycle ON party_goals (election_cycle);

CREATE TABLE goal_vote_matches (
  id                SERIAL PRIMARY KEY,
  goal_id           INTEGER NOT NULL REFERENCES party_goals (id) ON DELETE CASCADE,
  beteckning        TEXT NOT NULL,
  forslagspunkt     TEXT NOT NULL DEFAULT '',
  relevance_score   DOUBLE PRECISION NOT NULL CHECK (relevance_score >= 0 AND relevance_score <= 1),
  aligned_direction TEXT NOT NULL CHECK (aligned_direction IN ('Ja', 'Nej', 'Avstår', 'unclear')),
  explanation       TEXT,
  proposed_by_party TEXT,
  proposal_type     TEXT CHECK (proposal_type IN ('prop', 'mot', 'bet')),
  context_note      TEXT,
  document_title    TEXT,
  verified          BOOLEAN NOT NULL DEFAULT false,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (goal_id, beteckning, forslagspunkt)
);
CREATE INDEX idx_goal_vote_matches_goal_id ON goal_vote_matches (goal_id);
CREATE INDEX idx_goal_vote_matches_beteckning ON goal_vote_matches (beteckning, forslagspunkt);

CREATE TABLE promises (
  id            SERIAL PRIMARY KEY,
  politician_id TEXT NOT NULL REFERENCES politicians (intressent_id) ON DELETE CASCADE,
  speech_id     INTEGER NOT NULL REFERENCES speeches (id) ON DELETE CASCADE,
  promise_text  TEXT NOT NULL,
  topic         TEXT NOT NULL,
  specificity   TEXT NOT NULL CHECK (specificity IN ('concrete', 'directional', 'rhetorical')),
  keywords      TEXT[] NOT NULL DEFAULT '{}',
  extracted_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_promises_politician_id ON promises (politician_id);
CREATE INDEX idx_promises_topic ON promises (topic);

CREATE TABLE promise_vote_matches (
  id                SERIAL PRIMARY KEY,
  promise_id        INTEGER NOT NULL REFERENCES promises (id) ON DELETE CASCADE,
  relevance_score   DOUBLE PRECISION NOT NULL CHECK (relevance_score >= 0 AND relevance_score <= 1),
  alignment         TEXT NOT NULL CHECK (alignment IN ('supports', 'contradicts', 'unclear')),
  explanation       TEXT,
  proposed_by_party TEXT,
  proposal_type     TEXT CHECK (proposal_type IN ('prop', 'mot', 'bet')),
  verified          BOOLEAN NOT NULL DEFAULT false,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_promise_vote_matches_promise_id ON promise_vote_matches (promise_id);

CREATE MATERIALIZED VIEW party_scorecards AS
SELECT pg.party,
       pg.id AS goal_id,
       pg.goal_text,
       pg.topic,
       count(DISTINCT (gvm.beteckning || ':' || gvm.forslagspunkt)) AS relevant_votes
FROM party_goals pg
LEFT JOIN goal_vote_matches gvm ON gvm.goal_id = pg.id
GROUP BY pg.party, pg.id, pg.goal_text, pg.topic;
CREATE UNIQUE INDEX idx_party_scorecards_pk ON party_scorecards (party, goal_id);
CREATE INDEX idx_party_scorecards_party ON party_scorecards (party);
