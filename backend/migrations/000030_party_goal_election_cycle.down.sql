DROP INDEX IF EXISTS idx_party_goals_election_cycle;
ALTER TABLE party_goals DROP CONSTRAINT IF EXISTS party_goals_election_cycle_check;
ALTER TABLE party_goals DROP COLUMN IF EXISTS election_cycle;
