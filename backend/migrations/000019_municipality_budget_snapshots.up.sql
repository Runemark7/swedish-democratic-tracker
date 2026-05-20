CREATE TABLE municipality_budget_snapshots (
    mun_code    TEXT        NOT NULL,
    area_name   TEXT        NOT NULL,
    year        INT         NOT NULL,
    value_mnkr  NUMERIC     NOT NULL,
    total_mnkr  NUMERIC     NOT NULL,
    pct         NUMERIC     NOT NULL,
    fetched_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (mun_code, area_name, year)
);

CREATE INDEX ON municipality_budget_snapshots (mun_code);
CREATE INDEX ON municipality_budget_snapshots (area_name, year);
