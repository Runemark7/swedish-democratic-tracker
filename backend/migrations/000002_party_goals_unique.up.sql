ALTER TABLE party_goals ADD CONSTRAINT uq_party_goals_party_text UNIQUE (party, goal_text);
