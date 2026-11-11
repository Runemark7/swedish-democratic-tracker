-- Removes only what 000062 added: motion kopplingar and the batch items that
-- point at them. Löfte kopplingar and their batch items are untouched.
DELETE FROM veckobatch_items WHERE motion_koppling_id IS NOT NULL;
DROP INDEX IF EXISTS veckobatch_items_motion_uniq;
DROP INDEX IF EXISTS idx_veckobatch_items_motion_koppling;
ALTER TABLE veckobatch_items DROP CONSTRAINT veckobatch_items_one_koppling;
ALTER TABLE veckobatch_items DROP COLUMN motion_koppling_id;
ALTER TABLE veckobatch_items ALTER COLUMN koppling_id SET NOT NULL;
DROP TABLE motion_kopplingar;
