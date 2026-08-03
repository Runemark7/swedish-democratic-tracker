-- What we actually hold, per riksmöte, so the site can state its own coverage
-- rather than implying completeness.
--
-- expected_voteringar is Riksdagen's own count (@traffar from /dokumentlista),
-- not ours: the denominator of a coverage claim should come from the source,
-- not from the party making the claim.
CREATE TABLE IF NOT EXISTS ingestion_coverage (
  riksmote            TEXT PRIMARY KEY,
  expected_voteringar INTEGER,
  ingested_voteringar INTEGER NOT NULL DEFAULT 0,
  -- Voteringar the betänkande-partitioned fetch cannot reach because the
  -- document carries no beteckning (a votering on a motion rather than on a
  -- committee report). Recorded so a shortfall is explained rather than
  -- mysterious: absence must never read as an unexplained gap.
  unreachable_voteringar INTEGER NOT NULL DEFAULT 0,
  last_vote_date      DATE,
  checked_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Added after the table shipped in an earlier local run: CREATE TABLE IF NOT
-- EXISTS is a no-op on an existing table, so the column needs its own
-- idempotent statement for databases that already have the table.
ALTER TABLE ingestion_coverage
  ADD COLUMN IF NOT EXISTS unreachable_voteringar INTEGER NOT NULL DEFAULT 0;
