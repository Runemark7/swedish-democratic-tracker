-- Progress of the opt-in motions backfill (cmd/backfill-motions), per riksmöte
-- and source. New table only: nothing existing is read, rewritten or deleted.
--
-- completed_at is set when the command has finished the riksmöte for the
-- source: every motion the listing reports is held (source 'motioner'), or no
-- betänkande the riksmöte's motions point at is left to read ('betankanden').
-- A restart skips completed rows and carries on with the rest.
CREATE TABLE IF NOT EXISTS motion_backfill (
    riksmote     TEXT NOT NULL,
    source       TEXT NOT NULL CHECK (source IN ('motioner', 'betankanden')),
    expected     INT  NOT NULL DEFAULT 0,
    held         INT  NOT NULL DEFAULT 0,
    completed_at TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (riksmote, source)
);
