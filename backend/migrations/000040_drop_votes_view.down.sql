-- Restores the compatibility view exactly as 000039 defined it, so rolling
-- back to that migration lands on the schema it expects.
CREATE VIEW votes AS
SELECT
  vg.votering_id,
  b.politician_id,
  b.party,
  b.vote_result,
  vg.beteckning,
  vg.forslagspunkt,
  vg.session,
  vg.dok_id,
  vg.system_datum,
  vg.proposed_by_party,
  vg.proposal_type,
  vg.proposal_dok_id,
  vg.document_title,
  vg.origin_enriched,
  vg.created_at
FROM ballots b
JOIN voteringar vg ON vg.id = b.votering_ref;
