-- Curated macro-economic KPIs displayed on the riksdag overview page.
-- Values updated manually when official statistics are published.
CREATE TABLE riksdag_kpis (
  id           SERIAL  PRIMARY KEY,
  label        TEXT    NOT NULL,
  description  TEXT    NOT NULL DEFAULT '',
  raw          NUMERIC NOT NULL,
  target       NUMERIC NOT NULL,
  worse_higher BOOLEAN NOT NULL DEFAULT false,
  unit         TEXT    NOT NULL DEFAULT '',
  trend        TEXT    NOT NULL CHECK (trend IN ('up', 'down', 'flat')),
  delta        TEXT    NOT NULL DEFAULT '',
  note         TEXT    NOT NULL DEFAULT '',
  source_url   TEXT,
  year         INTEGER NOT NULL,
  sort_order   INTEGER NOT NULL DEFAULT 0,
  UNIQUE (label, year)
);

CREATE INDEX idx_riksdag_kpis_year ON riksdag_kpis (year);

INSERT INTO riksdag_kpis (label, description, raw, target, worse_higher, unit, trend, delta, note, source_url, year, sort_order) VALUES
(
  'Statsskuld/BNP',
  'Statens samlade skuld i förhållande till BNP. Hög skuld begränsar statens möjlighet att investera i välfärd och krisberedskap utan att höja skatterna.',
  38, 35, true, '%', 'down', '−1,2 pp',
  'Källa: Riksgälden',
  'https://www.riksgalden.se/sv/var-verksamhet/statsskulden/',
  2025, 1
),
(
  'Arbetslöshet',
  'Andel av arbetskraften som är arbetslösa. Hög arbetslöshet minskar skatteintäkterna och ökar kostnaderna för socialförsäkringssystemet.',
  8.5, 5, true, '%', 'up', '+0,3 pp',
  'Källa: SCB',
  'https://www.scb.se/hitta-statistik/statistik-efter-amne/arbetsmarknad/',
  2025, 2
),
(
  'Inflation (KPI)',
  'Konsumentprisindex visar hur snabbt priserna stiger. Riksbankens mål är 2 %. För hög inflation urholkar hushållens köpkraft.',
  1.8, 2, false, '%', 'down', '−0,6 pp',
  'Källa: SCB',
  'https://www.scb.se/hitta-statistik/statistik-efter-amne/priser-och-konsumtion/konsumentprisindex/konsumentprisindex-kpi/',
  2025, 3
),
(
  'BNP-tillväxt',
  'Hur snabbt ekonomin växer. Positiv tillväxt innebär fler jobb och ökade skatteintäkter som kan finansiera välfärden.',
  1.2, 2, false, '%', 'up', '+0,8 pp',
  'Källa: Konjunkturinstitutet',
  'https://www.konj.se/statistik/prognoser.html',
  2025, 4
);
