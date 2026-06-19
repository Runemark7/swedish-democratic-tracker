-- Remove ghost municipality rows.
--
-- SCB's Kfmandat code list still includes defunct municipalities — notably
-- 1229 Bara, merged into Svedala kommun in 1977. The seeder's len(code)==4
-- filter let them through, inserting rows with 0 mandates and 0 population
-- that render as empty kommuner in the UI (and as "code not in SCB set"
-- errors in the value-verification audit). The seeder now skips codes with no
-- 2022 mandates; this migration deletes the rows that were already seeded.
--
-- Delete dependent election-result rows first (FK is ON DELETE NO ACTION),
-- then the municipalities themselves. A current kommun always has 2022
-- mandate data, so total_mandates = 0 AND population = 0 uniquely identifies
-- the ghosts.

DELETE FROM municipal_election_results
WHERE municipality_code IN (
    SELECT code FROM municipalities WHERE total_mandates = 0 AND population = 0
);

DELETE FROM municipalities
WHERE total_mandates = 0 AND population = 0;
