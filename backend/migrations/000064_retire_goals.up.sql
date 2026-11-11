-- Retire goals, keyword matching, scorecards and the old promises feature
-- (issue #164, spec #129). Löften (migration 000057, #156) replaces them.
--
-- DROPS, exactly:
--   materialized view  party_scorecards   (and its indexes idx_party_scorecards_pk, idx_party_scorecards_party)
--   tables             goal_vote_matches, promise_vote_matches, promises, party_goals
--   sequences          goal_vote_matches_id_seq, promise_vote_matches_id_seq,
--                      promises_id_seq, party_goals_id_seq   (owned by the tables, dropped with them)
--
-- Does NOT touch loften, valmanifest, kopplingar, veckobatch*, aendringslogg
-- or any other table. The 2022 goals already live on as archived Löften with
-- their ids kept (000057); the seven 2026 goals were deliberately not copied
-- there and are preserved only by the pg_dump taken before deploy.
--
-- Irreversible for data: the down migration recreates the empty schema; rows
-- come back from the pre-deploy pg_dump (see docs/runbooks/retire-goals.md).
--
-- Guard: refuse to drop if the archive is missing. Every party_goals row that
-- is not a 2026 goal must exist as an archived loften with the same id.

DO $$
DECLARE
  missing INT;
BEGIN
  SELECT count(*) INTO missing
  FROM party_goals g
  WHERE (g.election_cycle IS NULL OR g.election_cycle <> '2026')
    AND NOT EXISTS (
      SELECT 1 FROM loften l WHERE l.id = g.id AND l.visibility = 'archived'
    );
  IF missing > 0 THEN
    RAISE EXCEPTION 'retire goals: % party_goals rows have no archived loften (#156 archive missing); refusing to drop', missing;
  END IF;
END $$;

DROP MATERIALIZED VIEW IF EXISTS party_scorecards;
DROP TABLE IF EXISTS goal_vote_matches;
DROP TABLE IF EXISTS promise_vote_matches;
DROP TABLE IF EXISTS promises;
DROP TABLE IF EXISTS party_goals;
