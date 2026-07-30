-- Restore the original (buggy) view definition from 000001.

DROP MATERIALIZED VIEW IF EXISTS party_scorecards;

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

CREATE UNIQUE INDEX IF NOT EXISTS idx_party_scorecards_pk
  ON party_scorecards (party, goal_id);
CREATE INDEX IF NOT EXISTS idx_party_scorecards_party
  ON party_scorecards (party);
