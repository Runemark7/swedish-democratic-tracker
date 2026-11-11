DROP INDEX IF EXISTS motion_yrkanden_wording_hash_idx;
ALTER TABLE motion_yrkanden DROP COLUMN IF EXISTS wording_hash;
