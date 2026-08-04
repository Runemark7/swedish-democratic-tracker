-- Retire the computed alignment percentage.
--
-- The figure systematically disadvantaged opposition parties: 68.6% for
-- government parties against 43.2% for the opposition, measured across the
-- 2022-2026 record. That gap measures nothing about promise-keeping.
--
-- The mechanism: an opposition party files a motion, the committee proposes
-- rejecting it, the vote is on the committee's proposal -- so the party votes
-- Nej, and inferDirection, expecting Ja because the party filed the motion,
-- records it as breaking its own promise. It fails the other way too: an
-- opposition party voting with the government was also scored as not aligned.
--
-- 88% of scored matches rested on the single near-tautological rule "the goal's
-- own party proposed it, so voting Ja aligns".
--
-- This cannot be tuned. Direction depends on what a betänkande actually
-- proposes -- bifall or avslag, and to what -- which no heuristic over the
-- proposer's identity can recover.
--
-- relevant_votes survives: "this goal was matched to N vote points" is a fact
-- about our matching, not a judgement about the party. The promise and the
-- votes are both shown, both sourced, and the reader connects them.

DROP MATERIALIZED VIEW IF EXISTS party_scorecards;

CREATE MATERIALIZED VIEW party_scorecards AS
SELECT
  pg.party,
  pg.id                                                       AS goal_id,
  pg.goal_text,
  pg.topic,
  COUNT(DISTINCT gvm.beteckning || ':' || gvm.forslagspunkt)  AS relevant_votes
FROM party_goals pg
LEFT JOIN goal_vote_matches gvm ON gvm.goal_id = pg.id
GROUP BY pg.party, pg.id, pg.goal_text, pg.topic;

CREATE UNIQUE INDEX IF NOT EXISTS idx_party_scorecards_pk
  ON party_scorecards (party, goal_id);
CREATE INDEX IF NOT EXISTS idx_party_scorecards_party
  ON party_scorecards (party);
