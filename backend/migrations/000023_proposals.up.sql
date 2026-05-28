CREATE TABLE proposals (
  id              TEXT PRIMARY KEY,
  title           TEXT NOT NULL,
  department_code TEXT NOT NULL,
  riksdag_year    TEXT NOT NULL,
  published_at    TIMESTAMPTZ,
  url             TEXT NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_proposals_dept_year ON proposals(department_code, riksdag_year);
