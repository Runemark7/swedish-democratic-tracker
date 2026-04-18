-- Budget tracking tables for national (statsbudgeten) budget data.
-- Expenditure areas (utgiftsområden) are the 27 fixed budget categories.
-- Amounts stored in KSEK (thousands of kronor) as BIGINT to avoid float drift.

CREATE TABLE expenditure_areas (
  id         SERIAL PRIMARY KEY,
  code       TEXT    NOT NULL UNIQUE,   -- "UO1" .. "UO27"
  name       TEXT    NOT NULL,
  description TEXT,
  sort_order INTEGER NOT NULL
);

CREATE TABLE budget_years (
  id         SERIAL PRIMARY KEY,
  year       INTEGER NOT NULL,
  status     TEXT    NOT NULL CHECK (status IN ('decided', 'proposed')),
  total_ksek BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (year, status)
);

CREATE TABLE budget_allocations (
  id                   SERIAL  PRIMARY KEY,
  budget_year_id       INTEGER NOT NULL REFERENCES budget_years (id) ON DELETE CASCADE,
  expenditure_area_id  INTEGER NOT NULL REFERENCES expenditure_areas (id) ON DELETE CASCADE,
  amount_ksek          BIGINT  NOT NULL,
  source               TEXT    NOT NULL DEFAULT 'government' CHECK (source IN ('government', 'party_alternative')),
  party                TEXT,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_budget_alloc_unique
  ON budget_allocations (budget_year_id, expenditure_area_id, source, COALESCE(party, ''));

CREATE INDEX idx_budget_alloc_year   ON budget_allocations (budget_year_id);
CREATE INDEX idx_budget_alloc_area   ON budget_allocations (expenditure_area_id);
CREATE INDEX idx_budget_years_year   ON budget_years (year);
