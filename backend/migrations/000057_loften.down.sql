-- Removes only what 000057 created. party_goals and the other legacy tables
-- were never touched, so no data is lost.
DROP TABLE IF EXISTS loften;
DROP SEQUENCE IF EXISTS loften_id_seq;
DROP TABLE IF EXISTS valmanifest;
