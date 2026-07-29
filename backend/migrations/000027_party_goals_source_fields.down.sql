ALTER TABLE party_goals
  DROP COLUMN IF EXISTS source_quote,
  DROP COLUMN IF EXISTS source_url;
