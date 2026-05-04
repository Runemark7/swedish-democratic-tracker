-- Current Riksdag government composition.
-- One row in riksdag_government is the active coalition (valid_to IS NULL).
-- Source: Riksdagen.se — https://www.riksdagen.se/sv/riksdagen/regeringen/
-- Seats: 2022 election results from Valmyndigheten — https://www.val.se/valresultat/riksdag-region-och-kommun/2022/valresultat.html
-- Party colors: Official party visual identities.
-- UPDATE REQUIRED after each general election (held every 4 years, next 2026).
CREATE TABLE riksdag_government (
  id          SERIAL PRIMARY KEY,
  type_label  TEXT NOT NULL,
  valid_from  DATE NOT NULL,
  valid_to    DATE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE riksdag_government_parties (
  id            SERIAL PRIMARY KEY,
  government_id INTEGER NOT NULL REFERENCES riksdag_government(id) ON DELETE CASCADE,
  name          TEXT    NOT NULL,
  short         TEXT    NOT NULL,
  seats         INTEGER NOT NULL,
  color         TEXT    NOT NULL,
  role          TEXT    NOT NULL CHECK (role IN ('governing', 'support', 'opposition')),
  sort_order    INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_riksdag_gov_parties ON riksdag_government_parties (government_id, role);

-- Seed: Swedish government 2022– (Tidöavtalet)
INSERT INTO riksdag_government (type_label, valid_from) VALUES
  ('Regeringskoalition (stöd av SD)', '2022-10-18');

INSERT INTO riksdag_government_parties (government_id, name, short, seats, color, role, sort_order)
SELECT g.id, v.name, v.short, v.seats, v.color, v.role, v.sort_order
FROM (SELECT id FROM riksdag_government ORDER BY id DESC LIMIT 1) g
CROSS JOIN (VALUES
  ('Moderaterna',          'M',  68,  '#1E88E5', 'governing',  1),
  ('Kristdemokraterna',    'KD', 19,  '#00558E', 'governing',  2),
  ('Liberalerna',          'L',  16,  '#006AB3', 'governing',  3),
  ('Sverigedemokraterna',  'SD', 73,  '#DDD014', 'support',    4),
  ('Socialdemokraterna',   'S',  107, '#E8112D', 'opposition', 5),
  ('Vänsterpartiet',       'V',  24,  '#AF0000', 'opposition', 6),
  ('Centerpartiet',        'C',  24,  '#009933', 'opposition', 7),
  ('Miljöpartiet',         'MP', 18,  '#83CF39', 'opposition', 8)
) AS v(name, short, seats, color, role, sort_order);
