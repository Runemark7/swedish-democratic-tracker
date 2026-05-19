CREATE TABLE region_budget_snapshots (
    region_code TEXT        NOT NULL,
    area_name   TEXT        NOT NULL,
    year        INT         NOT NULL,
    value_mnkr  NUMERIC     NOT NULL,
    total_mnkr  NUMERIC     NOT NULL,
    pct         NUMERIC     NOT NULL,
    fetched_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (region_code, area_name, year)
);

CREATE INDEX ON region_budget_snapshots (region_code);
CREATE INDEX ON region_budget_snapshots (area_name, year);
