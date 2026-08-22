-- Removes only the five rows this migration added. 000033 owns the table and
-- the 2022-2026 row, so dropping either here would take down the period the
-- site currently reads.
--
-- Votes already ingested for these periods are left in place: the votes table
-- keys on session, not on a mandate_periods foreign key, so the rows simply
-- stop being reachable through a period-scoped query rather than being lost.
DELETE FROM mandate_periods
WHERE code IN ('2002-2006', '2006-2010', '2010-2014', '2014-2018', '2018-2022');
