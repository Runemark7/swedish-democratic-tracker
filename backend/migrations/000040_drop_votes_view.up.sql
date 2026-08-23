-- Drops the votes compatibility view.
--
-- 000039 kept it so the split could land without rewriting every caller in the
-- same change. Every query has since moved to voteringar and ballots directly,
-- so the view now only offers a way to reintroduce the join by accident --
-- writing `FROM votes` and getting 5.5M denormalised rows back is exactly the
-- shape this work removed.
DROP VIEW IF EXISTS votes;
