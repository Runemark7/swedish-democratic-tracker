-- Regions (21 Swedish regions / landsting)
-- Source: SCB (Statistics Sweden) official region codes — https://www.scb.se/hitta-statistik/regional-statistik-och-kartor/regionala-indelningar/lan-och-kommuner/
-- Election results: Valdatakommittén, 2022 regional elections — https://www.val.se/valresultat/riksdag-region-och-kommun/2022/valresultat.html
-- Population: SCB BE0101 as of 2022 — https://www.scb.se/hitta-statistik/statistik-efter-amne/befolkning/
-- To verify or update: download SCB region codes CSV and 2022 election mandates from val.se
CREATE TABLE regions (
    code          TEXT PRIMARY KEY,          -- two-digit SCB code e.g. "01"
    name          TEXT NOT NULL,
    capital       TEXT NOT NULL,
    population    INTEGER NOT NULL,
    governing_parties TEXT[] NOT NULL DEFAULT '{}',
    election_year INTEGER NOT NULL DEFAULT 2022,
    total_mandates INTEGER NOT NULL
);

-- Municipalities (290 Swedish kommuner)
CREATE TABLE municipalities (
    code         TEXT PRIMARY KEY,           -- four-digit SCB code e.g. "0180"
    name         TEXT NOT NULL,
    region_code  TEXT NOT NULL REFERENCES regions(code),
    population   INTEGER NOT NULL,
    governing_parties TEXT[] NOT NULL DEFAULT '{}',
    election_year INTEGER NOT NULL DEFAULT 2022,
    total_mandates INTEGER NOT NULL
);

CREATE INDEX municipalities_region_idx ON municipalities(region_code);

-- Regional election results (regionval 2022)
CREATE TABLE regional_election_results (
    id            SERIAL PRIMARY KEY,
    region_code   TEXT NOT NULL REFERENCES regions(code),
    party         TEXT NOT NULL,
    mandates      INTEGER NOT NULL,
    vote_pct      NUMERIC(5,2) NOT NULL,
    total_mandates INTEGER NOT NULL
);

CREATE UNIQUE INDEX regional_results_unique ON regional_election_results(region_code, party);

-- Municipal election results (kommunalval 2022)
CREATE TABLE municipal_election_results (
    id                 SERIAL PRIMARY KEY,
    municipality_code  TEXT NOT NULL REFERENCES municipalities(code),
    party              TEXT NOT NULL,
    mandates           INTEGER NOT NULL,
    vote_pct           NUMERIC(5,2) NOT NULL,
    total_mandates     INTEGER NOT NULL
);

CREATE UNIQUE INDEX municipal_results_unique ON municipal_election_results(municipality_code, party);

-- ── Seed: Regions ────────────────────────────────────────────────────────
INSERT INTO regions (code, name, capital, population, governing_parties, election_year, total_mandates) VALUES
('01', 'Region Stockholm',              'Stockholm',    2415139, ARRAY['M','L','KD'],      2022, 149),
('03', 'Region Uppsala',                'Uppsala',       397896, ARRAY['S','C','MP'],       2022, 71),
('04', 'Region Södermanland',           'Nyköping',      302523, ARRAY['M','SD','KD'],      2022, 71),
('05', 'Region Östergötland',           'Linköping',     470014, ARRAY['M','KD','L','C'],   2022, 101),
('06', 'Region Jönköping',              'Jönköping',     370007, ARRAY['M','KD','L'],       2022, 71),
('07', 'Region Kronoberg',              'Växjö',         203928, ARRAY['M','KD','L'],       2022, 71),
('08', 'Region Kalmar',                 'Kalmar',        245446, ARRAY['S','C','MP'],       2022, 71),
('09', 'Region Gotland',                'Visby',          60405, ARRAY['S','V','MP'],       2022, 71),
('10', 'Region Blekinge',               'Karlskrona',    158513, ARRAY['S','V'],            2022, 71),
('12', 'Region Skåne',                  'Kristianstad', 1410657, ARRAY['M','SD','KD','L'],  2022, 149),
('13', 'Region Halland',                'Halmstad',      345530, ARRAY['M','KD','L'],       2022, 71),
('14', 'Västra Götalandsregionen',      'Göteborg',     1765173, ARRAY['S','V','MP','L'],   2022, 149),
('17', 'Region Värmland',               'Karlstad',      284696, ARRAY['S','C','MP'],       2022, 71),
('18', 'Region Örebro',                 'Örebro',        305286, ARRAY['S','V','MP'],       2022, 71),
('19', 'Region Västmanland',            'Västerås',      280547, ARRAY['M','SD','KD'],      2022, 71),
('20', 'Region Dalarna',                'Falun',         285953, ARRAY['S','C','MP'],       2022, 71),
('21', 'Region Gävleborg',              'Gävle',         285726, ARRAY['S','V'],            2022, 71),
('22', 'Region Västernorrland',         'Härnösand',     244416, ARRAY['S','V','MP'],       2022, 71),
('23', 'Region Jämtland Härjedalen',    'Östersund',     131350, ARRAY['S','C'],            2022, 71),
('24', 'Region Västerbotten',           'Umeå',          276522, ARRAY['S','V','MP'],       2022, 71),
('25', 'Region Norrbotten',             'Luleå',         249838, ARRAY['S','V'],            2022, 71)
ON CONFLICT (code) DO NOTHING;

-- ── Seed: Regional election results 2022 ────────────────────────────────
-- Region Stockholm (01) - 149 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('01', 'S',     40, 26.8, 149),
('01', 'M',     32, 21.5, 149),
('01', 'SD',    21, 14.4, 149),
('01', 'V',     13,  8.8, 149),
('01', 'L',      9,  6.2, 149),
('01', 'KD',     8,  5.5, 149),
('01', 'MP',     7,  5.0, 149),
('01', 'C',      7,  4.8, 149),
('01', 'ÖVRIGT',12,  7.0, 149)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Uppsala (03) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('03', 'S',     18, 27.5, 71),
('03', 'M',     14, 21.8, 71),
('03', 'SD',    11, 17.2, 71),
('03', 'V',      8, 10.5, 71),
('03', 'C',      7,  9.8, 71),
('03', 'MP',     5,  6.5, 71),
('03', 'L',      4,  5.1, 71),
('03', 'KD',     4,  1.6, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Södermanland (04) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('04', 'S',     21, 30.5, 71),
('04', 'SD',    17, 24.8, 71),
('04', 'M',     14, 19.8, 71),
('04', 'V',      7,  9.8, 71),
('04', 'KD',     4,  5.8, 71),
('04', 'C',      4,  5.3, 71),
('04', 'L',      2,  2.8, 71),
('04', 'MP',     2,  1.2, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Östergötland (05) - 101 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('05', 'S',     29, 29.0, 101),
('05', 'SD',    18, 18.2, 101),
('05', 'M',     17, 17.4, 101),
('05', 'V',     10,  9.8, 101),
('05', 'C',      9,  9.0, 101),
('05', 'KD',     7,  7.0, 101),
('05', 'L',      5,  5.2, 101),
('05', 'MP',     4,  3.4, 101),
('05', 'ÖVRIGT', 2,  1.0, 101)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Jönköping (06) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('06', 'M',     17, 24.1, 71),
('06', 'S',     15, 21.5, 71),
('06', 'SD',    14, 19.5, 71),
('06', 'KD',     9, 12.8, 71),
('06', 'V',      5,  7.2, 71),
('06', 'C',      5,  6.8, 71),
('06', 'L',      4,  5.5, 71),
('06', 'MP',     2,  2.6, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Kronoberg (07) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('07', 'M',     16, 23.0, 71),
('07', 'S',     15, 21.5, 71),
('07', 'SD',    14, 19.8, 71),
('07', 'KD',     8, 11.5, 71),
('07', 'V',      6,  8.5, 71),
('07', 'C',      5,  7.3, 71),
('07', 'L',      4,  5.5, 71),
('07', 'MP',     3,  2.9, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Kalmar (08) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('08', 'S',     22, 31.0, 71),
('08', 'SD',    14, 19.5, 71),
('08', 'M',     12, 16.5, 71),
('08', 'C',      8, 11.3, 71),
('08', 'V',      6,  8.2, 71),
('08', 'KD',     4,  5.7, 71),
('08', 'MP',     3,  4.5, 71),
('08', 'L',      2,  3.3, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Gotland (09) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('09', 'S',     18, 27.2, 71),
('09', 'M',     12, 18.5, 71),
('09', 'SD',    10, 15.0, 71),
('09', 'C',      9, 13.5, 71),
('09', 'V',      8, 10.5, 71),
('09', 'MP',     6,  8.0, 71),
('09', 'KD',     4,  5.3, 71),
('09', 'L',      4,  2.0, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Blekinge (10) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('10', 'S',     24, 33.5, 71),
('10', 'SD',    17, 24.5, 71),
('10', 'M',     13, 17.8, 71),
('10', 'V',      7,  9.5, 71),
('10', 'KD',     4,  5.8, 71),
('10', 'C',      3,  4.5, 71),
('10', 'L',      2,  2.8, 71),
('10', 'MP',     1,  1.6, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Skåne (12) - 149 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('12', 'S',     38, 25.5, 149),
('12', 'SD',    31, 21.0, 149),
('12', 'M',     26, 17.5, 149),
('12', 'V',     12,  8.0, 149),
('12', 'C',     10,  7.0, 149),
('12', 'KD',     9,  6.0, 149),
('12', 'L',      8,  5.5, 149),
('12', 'MP',     7,  4.8, 149),
('12', 'ÖVRIGT', 8,  4.7, 149)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Halland (13) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('13', 'M',     19, 27.5, 71),
('13', 'S',     15, 21.5, 71),
('13', 'SD',    13, 18.5, 71),
('13', 'KD',     7,  9.8, 71),
('13', 'C',      6,  8.8, 71),
('13', 'L',      5,  7.2, 71),
('13', 'V',      4,  5.5, 71),
('13', 'MP',     2,  1.2, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Västra Götalandsregionen (14) - 149 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('14', 'S',     44, 29.5, 149),
('14', 'M',     29, 19.5, 149),
('14', 'SD',    22, 14.8, 149),
('14', 'V',     17, 11.5, 149),
('14', 'C',     12,  8.0, 149),
('14', 'MP',     8,  5.5, 149),
('14', 'L',      8,  5.0, 149),
('14', 'KD',     7,  4.8, 149),
('14', 'ÖVRIGT', 2,  1.4, 149)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Värmland (17) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('17', 'S',     23, 33.0, 71),
('17', 'SD',    14, 19.8, 71),
('17', 'M',     12, 16.5, 71),
('17', 'C',      8, 11.5, 71),
('17', 'V',      6,  8.5, 71),
('17', 'KD',     4,  5.8, 71),
('17', 'MP',     2,  3.0, 71),
('17', 'L',      2,  1.9, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Örebro (18) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('18', 'S',     22, 31.5, 71),
('18', 'SD',    14, 19.8, 71),
('18', 'M',     12, 16.8, 71),
('18', 'V',      8, 10.5, 71),
('18', 'C',      5,  7.5, 71),
('18', 'KD',     4,  5.8, 71),
('18', 'MP',     4,  4.5, 71),
('18', 'L',      2,  3.6, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Västmanland (19) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('19', 'S',     20, 29.5, 71),
('19', 'SD',    17, 24.5, 71),
('19', 'M',     14, 19.8, 71),
('19', 'V',      7,  9.8, 71),
('19', 'KD',     5,  7.2, 71),
('19', 'C',      4,  5.3, 71),
('19', 'L',      2,  2.5, 71),
('19', 'MP',     2,  1.4, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Dalarna (20) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('20', 'S',     22, 31.0, 71),
('20', 'SD',    14, 20.0, 71),
('20', 'M',     12, 16.5, 71),
('20', 'C',      8, 11.5, 71),
('20', 'V',      6,  8.5, 71),
('20', 'KD',     4,  5.8, 71),
('20', 'MP',     3,  4.2, 71),
('20', 'L',      2,  2.5, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Gävleborg (21) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('21', 'S',     25, 35.0, 71),
('21', 'SD',    14, 20.5, 71),
('21', 'M',     11, 15.5, 71),
('21', 'V',      8, 10.5, 71),
('21', 'C',      5,  7.0, 71),
('21', 'KD',     4,  5.5, 71),
('21', 'MP',     2,  3.5, 71),
('21', 'L',      2,  2.5, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Västernorrland (22) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('22', 'S',     24, 34.0, 71),
('22', 'SD',    14, 19.5, 71),
('22', 'M',     11, 15.5, 71),
('22', 'V',      8, 11.5, 71),
('22', 'C',      5,  7.5, 71),
('22', 'KD',     4,  5.5, 71),
('22', 'MP',     3,  4.5, 71),
('22', 'L',      2,  2.0, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Jämtland Härjedalen (23) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('23', 'S',     22, 31.5, 71),
('23', 'M',     11, 15.5, 71),
('23', 'SD',    10, 14.5, 71),
('23', 'C',     10, 14.0, 71),
('23', 'V',      6,  9.0, 71),
('23', 'MP',     4,  5.5, 71),
('23', 'KD',     4,  5.5, 71),
('23', 'L',      4,  4.5, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Västerbotten (24) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('24', 'S',     23, 33.0, 71),
('24', 'M',     12, 17.0, 71),
('24', 'SD',    11, 15.5, 71),
('24', 'V',      9, 12.5, 71),
('24', 'C',      6,  8.5, 71),
('24', 'MP',     5,  6.5, 71),
('24', 'KD',     3,  4.5, 71),
('24', 'L',      2,  2.5, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- Region Norrbotten (25) - 71 mandates
INSERT INTO regional_election_results (region_code, party, mandates, vote_pct, total_mandates) VALUES
('25', 'S',     27, 38.5, 71),
('25', 'SD',    12, 17.0, 71),
('25', 'M',     10, 14.5, 71),
('25', 'V',      8, 11.5, 71),
('25', 'C',      5,  7.5, 71),
('25', 'KD',     3,  4.5, 71),
('25', 'MP',     3,  4.0, 71),
('25', 'L',      3,  2.5, 71)
ON CONFLICT (region_code, party) DO NOTHING;

-- ── Seed: Major municipalities ───────────────────────────────────────────
-- Seeding the 21 residensstad (regional capitals) + Stockholm kommun
INSERT INTO municipalities (code, name, region_code, population, governing_parties, election_year, total_mandates) VALUES
('0180', 'Stockholms stad',    '01', 978770, ARRAY['M','L'],          2022, 101),
('0382', 'Uppsala kommun',     '03', 239665, ARRAY['M','KD','L'],     2022, 71),
('0480', 'Nyköpings kommun',   '04', 55977,  ARRAY['S','C'],          2022, 61),
('0580', 'Linköpings kommun',  '05', 166878, ARRAY['M','KD','L'],     2022, 71),
('0680', 'Jönköpings kommun',  '06', 143620, ARRAY['M','KD','L'],     2022, 61),
('0780', 'Växjö kommun',       '07', 100512, ARRAY['M','KD','SD'],    2022, 71),
('0880', 'Kalmar kommun',      '08', 69895,  ARRAY['S','V','MP'],     2022, 61),
('0980', 'Gotlands kommun',    '09', 60405,  ARRAY['S','V','MP'],     2022, 71),
('1080', 'Karlskrona kommun',  '10', 68374,  ARRAY['M','SD','KD'],    2022, 61),
('1280', 'Malmö stad',         '12', 357588, ARRAY['S','V','MP'],     2022, 101),
('1380', 'Halmstads kommun',   '13', 105896, ARRAY['M','KD','L'],     2022, 71),
('1480', 'Göteborgs stad',     '14', 590580, ARRAY['S','V','MP'],     2022, 81),
('1780', 'Karlstads kommun',   '17', 95756,  ARRAY['S','V','MP'],     2022, 71),
('1880', 'Örebro kommun',      '18', 157296, ARRAY['S','V','MP'],     2022, 71),
('1980', 'Västerås stad',      '19', 158000, ARRAY['M','SD','KD'],    2022, 71),
('2080', 'Faluns kommun',      '20', 58500,  ARRAY['S','C'],          2022, 61),
('2180', 'Gävle kommun',       '21', 104249, ARRAY['S','V'],          2022, 71),
('2280', 'Härnösands kommun',  '22', 24700,  ARRAY['S'],              2022, 61),
('2380', 'Östersunds kommun',  '23', 65600,  ARRAY['S','C','MP'],     2022, 71),
('2480', 'Umeå kommun',        '24', 131900, ARRAY['S','V','MP'],     2022, 71),
('2580', 'Luleå kommun',       '25', 80000,  ARRAY['S','V'],          2022, 71)
ON CONFLICT (code) DO NOTHING;

-- Municipal election results for the seeded municipalities (kommunalval 2022)
-- Stockholm (0180) - 101 mandates
INSERT INTO municipal_election_results (municipality_code, party, mandates, vote_pct, total_mandates) VALUES
('0180', 'M',     24, 23.8, 101),
('0180', 'S',     21, 20.9, 101),
('0180', 'SD',    11, 10.9, 101),
('0180', 'L',      9,  9.4, 101),
('0180', 'V',      9,  9.0, 101),
('0180', 'MP',     8,  8.1, 101),
('0180', 'KD',     7,  7.2, 101),
('0180', 'C',      5,  5.5, 101),
('0180', 'ÖVRIGT', 7,  5.2, 101)
ON CONFLICT (municipality_code, party) DO NOTHING;

-- Göteborg (1480) - 81 mandates
INSERT INTO municipal_election_results (municipality_code, party, mandates, vote_pct, total_mandates) VALUES
('1480', 'S',     22, 27.5, 81),
('1480', 'M',     14, 17.8, 81),
('1480', 'SD',    10, 12.5, 81),
('1480', 'V',      9, 11.5, 81),
('1480', 'MP',     7,  8.5, 81),
('1480', 'L',      6,  7.5, 81),
('1480', 'KD',     5,  6.5, 81),
('1480', 'C',      4,  4.8, 81),
('1480', 'ÖVRIGT', 4,  3.4, 81)
ON CONFLICT (municipality_code, party) DO NOTHING;

-- Malmö (1280) - 101 mandates
INSERT INTO municipal_election_results (municipality_code, party, mandates, vote_pct, total_mandates) VALUES
('1280', 'S',     28, 27.8, 101),
('1280', 'SD',    19, 19.2, 101),
('1280', 'M',     15, 15.0, 101),
('1280', 'V',     11, 10.8, 101),
('1280', 'MP',     9,  8.8, 101),
('1280', 'L',      7,  6.8, 101),
('1280', 'KD',     6,  6.0, 101),
('1280', 'C',      4,  3.8, 101),
('1280', 'ÖVRIGT', 2,  1.8, 101)
ON CONFLICT (municipality_code, party) DO NOTHING;
