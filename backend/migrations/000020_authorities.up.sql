CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE authorities (
    org_number          TEXT PRIMARY KEY,
    slug                TEXT NOT NULL UNIQUE,
    name                TEXT NOT NULL,
    type                TEXT NOT NULL DEFAULT '',
    principal_body      TEXT NOT NULL DEFAULT '',
    department          TEXT NOT NULL DEFAULT '',
    under_government    BOOLEAN NOT NULL DEFAULT FALSE,
    website             TEXT NOT NULL DEFAULT '',
    sfs                 TEXT NOT NULL DEFAULT '',
    expenditure_mdkr    DOUBLE PRECISION,
    budget_mdkr         DOUBLE PRECISION,
    headcount_int       INTEGER,
    year                INTEGER NOT NULL DEFAULT 0,
    expenditure_history JSONB NOT NULL DEFAULT '[]',
    headcount_history   JSONB NOT NULL DEFAULT '[]',
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_authorities_under_government ON authorities (under_government);
CREATE INDEX idx_authorities_expenditure ON authorities (expenditure_mdkr DESC NULLS LAST);
CREATE INDEX idx_authorities_name_trgm ON authorities USING gin (name gin_trgm_ops);
