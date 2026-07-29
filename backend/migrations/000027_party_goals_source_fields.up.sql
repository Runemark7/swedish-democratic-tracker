-- Add verifiable source provenance to party goals.
-- Both nullable so existing rows are unaffected.
ALTER TABLE party_goals
  ADD COLUMN source_url   TEXT,
  ADD COLUMN source_quote TEXT;
