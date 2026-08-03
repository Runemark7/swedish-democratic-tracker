-- A mandate period bounds the record so votes from one parliament cannot be
-- silently attributed to another. Without this, ingestion after the September
-- 2026 election would flow into the outgoing parliament's record and misstate
-- it: a differently-composed chamber, presented as the same one.
--
-- Dates are the election days that open and close the period.
CREATE TABLE IF NOT EXISTS mandate_periods (
  code       TEXT PRIMARY KEY,
  label      TEXT NOT NULL,
  start_date DATE NOT NULL,
  end_date   DATE NOT NULL,
  riksmoten  TEXT[] NOT NULL
);

INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
VALUES ('2022-2026', 'Mandatperioden 2022–2026', '2022-09-11', '2026-09-13',
        ARRAY['2022/23', '2023/24', '2024/25', '2025/26'])
ON CONFLICT (code) DO NOTHING;
