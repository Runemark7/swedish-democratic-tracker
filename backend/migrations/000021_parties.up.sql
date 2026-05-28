CREATE TABLE parties (
  code         TEXT PRIMARY KEY,
  name         TEXT NOT NULL,
  founded_year INT,
  ideology     TEXT,
  color_hex    TEXT NOT NULL,
  text_hex     TEXT NOT NULL,
  active       BOOLEAN NOT NULL DEFAULT true,
  website_url  TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO parties (code, name, founded_year, ideology, color_hex, text_hex, active, website_url) VALUES
  ('S',  'Socialdemokraterna',   1889, 'Socialdemokrati',                      '#E8192C', '#ffffff', true, 'https://www.socialdemokraterna.se'),
  ('M',  'Moderaterna',          1904, 'Liberal konservatism',                 '#52BDEC', '#ffffff', true, 'https://moderaterna.se'),
  ('SD', 'Sverigedemokraterna',  1988, 'Nationell konservatism',               '#DDDD00', '#1a1a1a', true, 'https://sd.se'),
  ('C',  'Centerpartiet',        1913, 'Agrarliberalism',                      '#009933', '#ffffff', true, 'https://www.centerpartiet.se'),
  ('V',  'Vänsterpartiet',       1921, 'Demokratisk socialism',                '#DA291C', '#ffffff', true, 'https://www.vansterpartiet.se'),
  ('KD', 'Kristdemokraterna',    1964, 'Kristdemokrati',                       '#005EA1', '#ffffff', true, 'https://www.kristdemokraterna.se'),
  ('L',  'Liberalerna',          1902, 'Liberalism',                           '#006AB3', '#ffffff', true, 'https://www.liberalerna.se'),
  ('MP', 'Miljöpartiet',         1981, 'Grön politik',                         '#83CF39', '#1a1a1a', true, 'https://www.mp.se');

ALTER TABLE party_goals
  ADD CONSTRAINT party_goals_party_fk FOREIGN KEY (party) REFERENCES parties(code);
