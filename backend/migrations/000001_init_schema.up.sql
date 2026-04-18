-- Core entities

CREATE TABLE politicians (
  intressent_id TEXT PRIMARY KEY,
  first_name    TEXT NOT NULL,
  last_name     TEXT NOT NULL,
  party         TEXT NOT NULL,
  constituency  TEXT,
  image_url     TEXT,
  is_active     BOOLEAN NOT NULL DEFAULT true,
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_politicians_party ON politicians (party);

CREATE TABLE speeches (
  id               SERIAL PRIMARY KEY,
  dok_id           TEXT NOT NULL,
  anforande_nummer TEXT,
  politician_id    TEXT NOT NULL REFERENCES politicians (intressent_id) ON DELETE CASCADE,
  party            TEXT NOT NULL,
  date             DATE NOT NULL,
  topic_heading    TEXT,
  speech_text      TEXT,
  related_dok_id   TEXT,
  ai_processed     BOOLEAN NOT NULL DEFAULT false,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (dok_id, anforande_nummer)
);

CREATE INDEX idx_speeches_politician_id ON speeches (politician_id);
CREATE INDEX idx_speeches_date         ON speeches (date DESC);
CREATE INDEX idx_speeches_unprocessed  ON speeches (id) WHERE ai_processed = false;

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
  -- Enriched proposal origin
  proposed_by_party TEXT,
  proposal_type     TEXT CHECK (proposal_type IN ('prop', 'mot', 'bet')),
  proposal_dok_id   TEXT,
  document_title    TEXT,
  origin_enriched   BOOLEAN NOT NULL DEFAULT false,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (votering_id, politician_id)
);

CREATE INDEX idx_votes_politician_id  ON votes (politician_id);
CREATE INDEX idx_votes_beteckning     ON votes (beteckning, forslagspunkt);
CREATE INDEX idx_votes_session        ON votes (session);
CREATE INDEX idx_votes_party          ON votes (party);
CREATE INDEX idx_votes_unenriched     ON votes (id) WHERE origin_enriched = false;

-- AI-extracted data

CREATE TABLE party_goals (
  id                 SERIAL PRIMARY KEY,
  party              TEXT NOT NULL,
  goal_text          TEXT NOT NULL,
  topic              TEXT NOT NULL,
  specificity        TEXT NOT NULL CHECK (specificity IN ('concrete', 'directional', 'rhetorical')),
  source_document    TEXT NOT NULL,
  keywords           TEXT[] NOT NULL DEFAULT '{}',
  relevant_committees TEXT[] NOT NULL DEFAULT '{}',
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_party_goals_party ON party_goals (party);
CREATE INDEX idx_party_goals_topic ON party_goals (topic);

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
CREATE INDEX idx_promises_topic         ON promises (topic);

-- Cross-reference results

CREATE TABLE goal_vote_matches (
  id               SERIAL PRIMARY KEY,
  goal_id          INTEGER NOT NULL REFERENCES party_goals (id) ON DELETE CASCADE,
  beteckning       TEXT NOT NULL,
  forslagspunkt    TEXT NOT NULL DEFAULT '',
  relevance_score  FLOAT NOT NULL CHECK (relevance_score BETWEEN 0 AND 1),
  aligned_direction TEXT NOT NULL CHECK (aligned_direction IN ('Ja', 'Nej', 'Avstår', 'unclear')),
  explanation      TEXT,
  proposed_by_party TEXT,
  proposal_type    TEXT CHECK (proposal_type IN ('prop', 'mot', 'bet')),
  context_note     TEXT,
  document_title   TEXT,
  verified         BOOLEAN NOT NULL DEFAULT false,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (goal_id, beteckning, forslagspunkt)
);

CREATE INDEX idx_goal_vote_matches_goal_id     ON goal_vote_matches (goal_id);
CREATE INDEX idx_goal_vote_matches_beteckning  ON goal_vote_matches (beteckning, forslagspunkt);

CREATE TABLE promise_vote_matches (
  id               SERIAL PRIMARY KEY,
  promise_id       INTEGER NOT NULL REFERENCES promises (id) ON DELETE CASCADE,
  vote_id          INTEGER NOT NULL REFERENCES votes (id) ON DELETE CASCADE,
  relevance_score  FLOAT NOT NULL CHECK (relevance_score BETWEEN 0 AND 1),
  alignment        TEXT NOT NULL CHECK (alignment IN ('supports', 'contradicts', 'unclear')),
  explanation      TEXT,
  proposed_by_party TEXT,
  proposal_type    TEXT CHECK (proposal_type IN ('prop', 'mot', 'bet')),
  verified         BOOLEAN NOT NULL DEFAULT false,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (promise_id, vote_id)
);

CREATE INDEX idx_promise_vote_matches_promise_id ON promise_vote_matches (promise_id);
CREATE INDEX idx_promise_vote_matches_vote_id    ON promise_vote_matches (vote_id);

-- Ingestion cursor table (tracks last-fetched position per data type)
CREATE TABLE ingestion_cursors (
  data_type   TEXT PRIMARY KEY,  -- 'speeches', 'votes', 'politicians'
  last_date   DATE,
  last_id     TEXT,
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Materialized views for scorecards

CREATE MATERIALIZED VIEW party_scorecards AS
SELECT
  pg.party,
  pg.id                                                         AS goal_id,
  pg.goal_text,
  pg.topic,
  COUNT(DISTINCT gvm.beteckning || ':' || gvm.forslagspunkt)   AS relevant_votes,
  ROUND(
    COUNT(CASE WHEN v.vote_result = gvm.aligned_direction THEN 1 END)::numeric /
    NULLIF(COUNT(v.id), 0) * 100,
    1
  )                                                             AS alignment_pct
FROM party_goals pg
LEFT JOIN goal_vote_matches gvm ON gvm.goal_id = pg.id
LEFT JOIN votes v
  ON  v.beteckning     = gvm.beteckning
  AND v.forslagspunkt  = gvm.forslagspunkt
  AND v.party          = pg.party
GROUP BY pg.party, pg.id, pg.goal_text, pg.topic;

CREATE UNIQUE INDEX idx_party_scorecards_pk ON party_scorecards (party, goal_id);
CREATE INDEX idx_party_scorecards_party     ON party_scorecards (party);
