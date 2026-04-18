CREATE TABLE economic_indicators (
  id         SERIAL PRIMARY KEY,
  series_id  TEXT NOT NULL,
  topic      TEXT NOT NULL,
  region     TEXT,
  period     TEXT NOT NULL,
  value      NUMERIC NOT NULL,
  unit       TEXT NOT NULL,
  source     TEXT NOT NULL,
  fetched_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (series_id, region, period)
);
CREATE INDEX idx_indicators_topic  ON economic_indicators(topic);
CREATE INDEX idx_indicators_series ON economic_indicators(series_id, period);
