-- Restores the scored view from 000029; see that migration for its definition.
DROP MATERIALIZED VIEW IF EXISTS party_scorecards;

CREATE MATERIALIZED VIEW party_scorecards AS
WITH party_positions AS (
  SELECT
    beteckning,
    forslagspunkt,
    party,
    mode() WITHIN GROUP (ORDER BY vote_result) AS position
  FROM votes
  WHERE vote_result <> 'Frånvarande'
  GROUP BY beteckning, forslagspunkt, party
)
SELECT
  pg.party,
  pg.id                                                       AS goal_id,
  pg.goal_text,
  pg.topic,
  COUNT(DISTINCT gvm.beteckning || ':' || gvm.forslagspunkt)  AS relevant_votes,
  COUNT(*) FILTER (
    WHERE gvm.aligned_direction <> 'unclear'
      AND pp.position IS NOT NULL
  )                                                           AS scored_votes,
  COUNT(*) FILTER (
    WHERE pp.position = gvm.aligned_direction
  )                                                           AS aligned_votes,
  ROUND(
    COUNT(*) FILTER (WHERE pp.position = gvm.aligned_direction)::numeric /
    NULLIF(
      COUNT(*) FILTER (
        WHERE gvm.aligned_direction <> 'unclear'
          AND pp.position IS NOT NULL
      ), 0
    ) * 100,
    1
  )                                                           AS alignment_pct
FROM party_goals pg
LEFT JOIN goal_vote_matches gvm ON gvm.goal_id = pg.id
LEFT JOIN party_positions pp
  ON  pp.beteckning    = gvm.beteckning
  AND pp.forslagspunkt = gvm.forslagspunkt
  AND pp.party         = pg.party
GROUP BY pg.party, pg.id, pg.goal_text, pg.topic;

-- UNIQUE index is required for REFRESH MATERIALIZED VIEW CONCURRENTLY,
-- which RefreshScorecards uses.
CREATE UNIQUE INDEX IF NOT EXISTS idx_party_scorecards_pk
  ON party_scorecards (party, goal_id);
CREATE INDEX IF NOT EXISTS idx_party_scorecards_party
  ON party_scorecards (party);
