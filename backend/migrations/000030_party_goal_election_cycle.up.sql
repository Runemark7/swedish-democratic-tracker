-- Promise vintage.
--
-- The 2022-2026 voting record can only speak to promises made at or before the
-- 2022 election. 2026 campaign material answers a different question -- "what do
-- they promise now" -- and placing it beside a four-year voting record invites
-- the reading "promised this, didn't do it" about a promise six weeks old.
--
-- It is also a selection asymmetry as things stand: S is currently the only
-- party with 2026 material on the site, six weeks before the election, while
-- every other party shows 2022-era documents. Bias enters through selection.
--
-- NULL means the goal is not tied to an election cycle (a standing party
-- programme), and such goals remain in the record view.

ALTER TABLE party_goals ADD COLUMN IF NOT EXISTS election_cycle TEXT;

ALTER TABLE party_goals DROP CONSTRAINT IF EXISTS party_goals_election_cycle_check;
ALTER TABLE party_goals
  ADD CONSTRAINT party_goals_election_cycle_check
  CHECK (election_cycle IS NULL OR election_cycle IN ('2022', '2026'));

UPDATE party_goals SET election_cycle = '2026'
  WHERE source_document ILIKE '%2026%'
    AND election_cycle IS DISTINCT FROM '2026';

UPDATE party_goals SET election_cycle = '2022'
  WHERE source_document ILIKE '%2022%'
    AND election_cycle IS NULL;

CREATE INDEX IF NOT EXISTS idx_party_goals_election_cycle
  ON party_goals (election_cycle);
