-- Ändringslogg: the append-only audit log of the write API (issue #157, ADR-003).
--
-- ADDITIVE ONLY. No existing table is touched. Rows can never be updated,
-- deleted or truncated (trigger below); an undo is a new row that points at
-- the entry it reverses.
CREATE TABLE aendringslogg (
  id           BIGSERIAL PRIMARY KEY,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  actor        TEXT NOT NULL,                 -- who wrote: the token holder (the bot)
  on_behalf_of TEXT NOT NULL,                 -- the Redaktör the bot acted for
  chat_ref     TEXT NOT NULL,                 -- the chat answer the change came from
  action       TEXT NOT NULL CHECK (action IN ('upsert', 'undo')),
  entity       TEXT NOT NULL,                 -- e.g. 'valmanifest'
  entity_id    BIGINT NOT NULL,
  old_value    JSONB,                         -- NULL: did not exist before
  new_value    JSONB,                         -- NULL: does not exist after
  undoes       BIGINT REFERENCES aendringslogg (id),
  request_key  TEXT NOT NULL UNIQUE,          -- Idempotency-Key; never served by the API
  request_hash TEXT NOT NULL                  -- hash of the request, to catch key reuse
);
CREATE INDEX idx_aendringslogg_entity ON aendringslogg (entity, entity_id, id DESC);

CREATE FUNCTION aendringslogg_append_only() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'aendringslogg is append-only';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER aendringslogg_no_change
  BEFORE UPDATE OR DELETE ON aendringslogg
  FOR EACH ROW EXECUTE FUNCTION aendringslogg_append_only();
CREATE TRIGGER aendringslogg_no_truncate
  BEFORE TRUNCATE ON aendringslogg
  FOR EACH STATEMENT EXECUTE FUNCTION aendringslogg_append_only();
