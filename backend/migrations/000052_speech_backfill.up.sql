-- Progress and storage growth of the opt-in speech backfill (cmd/backfill-speeches),
-- per riksmöte. Additive only: a nullable column without a default (no table
-- rewrite, no existing row touched), one index and one new table.
--
-- speeches.riksmote is set by the backfill and by the speeches worker's upsert
-- (from dok_rm). Rows stored before this migration stay NULL until a backfill or
-- the worker meets them again; nothing is rewritten here.
ALTER TABLE speeches ADD COLUMN IF NOT EXISTS riksmote TEXT;

CREATE INDEX IF NOT EXISTS idx_speeches_riksmote ON speeches (riksmote) WHERE riksmote IS NOT NULL;

-- completed_at is set when the command has finished the riksmöte: every speech
-- the lists report is stored, every one has its text and no speaker is missing.
-- The *_added columns accumulate over the runs that worked on the riksmöte, so
-- an interrupted riksmöte still adds up to what it cost.
CREATE TABLE IF NOT EXISTS speech_backfill (
    riksmote            TEXT PRIMARY KEY,
    expected            INT    NOT NULL DEFAULT 0,
    text_bytes          BIGINT NOT NULL DEFAULT 0,
    rows_added          INT    NOT NULL DEFAULT 0,
    table_bytes_added   BIGINT NOT NULL DEFAULT 0,
    index_bytes_added   BIGINT NOT NULL DEFAULT 0,
    politicians_created INT    NOT NULL DEFAULT 0,
    unresolved_speakers INT    NOT NULL DEFAULT 0,
    completed_at        TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
