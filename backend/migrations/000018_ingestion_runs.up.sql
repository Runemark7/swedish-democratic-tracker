CREATE TABLE ingestion_runs (
    id            BIGSERIAL   PRIMARY KEY,
    worker_name   TEXT        NOT NULL,
    started_at    TIMESTAMPTZ NOT NULL,
    finished_at   TIMESTAMPTZ,
    status        TEXT        NOT NULL CHECK (status IN ('running', 'success', 'error')),
    rows_affected INT,
    error_msg     TEXT
);

CREATE INDEX ON ingestion_runs (worker_name, started_at DESC);
