CREATE TABLE ministers (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name            TEXT NOT NULL,
  title           TEXT NOT NULL,
  department      TEXT NOT NULL,
  department_code TEXT NOT NULL,
  party           TEXT NOT NULL REFERENCES parties(code),
  politician_id   TEXT REFERENCES politicians(intressent_id) ON DELETE SET NULL,
  photo_url       TEXT,
  bio             TEXT,
  active          BOOLEAN NOT NULL DEFAULT true,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ministers_party ON ministers(party);
CREATE INDEX idx_ministers_dept_code ON ministers(department_code);
CREATE INDEX idx_ministers_active ON ministers(active);

INSERT INTO ministers (name, title, department, department_code, party, politician_id, active) VALUES
  ('Ulf Kristersson',         'Statsminister',                               'Statsrådsberedningen',                       'SB', 'M',  (SELECT intressent_id FROM politicians WHERE first_name='Ulf'       AND last_name='Kristersson'         LIMIT 1), true),
  ('Ebba Busch',              'Vice statsminister, Energi- och näringsdepartementet', 'Energi- och näringsdepartementet',   'N',  'KD', (SELECT intressent_id FROM politicians WHERE first_name='Ebba'      AND last_name='Busch'               LIMIT 1), true),
  ('Elisabeth Svantesson',    'Finansminister',                              'Finansdepartementet',                        'Fi', 'M',  (SELECT intressent_id FROM politicians WHERE first_name='Elisabeth'  AND last_name='Svantesson'          LIMIT 1), true),
  ('Gunnar Strömmer',         'Justitieminister',                            'Justitiedepartementet',                      'Ju', 'M',  (SELECT intressent_id FROM politicians WHERE first_name='Gunnar'     AND last_name='Strömmer'            LIMIT 1), true),
  ('Maria Malmer Stenergard', 'Migrationsminister',                          'Justitiedepartementet',                      'Ju', 'M',  (SELECT intressent_id FROM politicians WHERE first_name='Maria'      AND last_name='Malmer Stenergard'   LIMIT 1), true),
  ('Tobias Billström',        'Utrikesminister',                             'Utrikesdepartementet',                       'UD', 'M',  (SELECT intressent_id FROM politicians WHERE first_name='Tobias'     AND last_name='Billström'           LIMIT 1), true),
  ('Johan Forssell',          'Minister för bistånd och utrikeshandel',      'Utrikesdepartementet',                       'UD', 'M',  (SELECT intressent_id FROM politicians WHERE first_name='Johan'      AND last_name='Forssell'            LIMIT 1), true),
  ('Pål Jonson',              'Försvarsminister',                            'Försvarsdepartementet',                      'Fö', 'M',  (SELECT intressent_id FROM politicians WHERE first_name='Pål'        AND last_name='Jonson'              LIMIT 1), true),
  ('Carl-Oskar Bohlin',       'Minister för civilt försvar',                 'Försvarsdepartementet',                      'Fö', 'M',  (SELECT intressent_id FROM politicians WHERE first_name='Carl-Oskar' AND last_name='Bohlin'             LIMIT 1), true),
  ('Jakob Forssmed',          'Socialminister',                              'Socialdepartementet',                        'S',  'KD', (SELECT intressent_id FROM politicians WHERE first_name='Jakob'      AND last_name='Forssmed'            LIMIT 1), true),
  ('Acko Ankarberg Johansson','Sjukvårdsminister',                           'Socialdepartementet',                        'S',  'KD', (SELECT intressent_id FROM politicians WHERE first_name='Acko'       AND last_name='Ankarberg Johansson' LIMIT 1), true),
  ('Camilla Waltersson Grönvall','Socialtjänstminister',                     'Socialdepartementet',                        'S',  'M',  (SELECT intressent_id FROM politicians WHERE first_name='Camilla'    AND last_name='Waltersson Grönvall' LIMIT 1), true),
  ('Johan Pehrson',           'Arbetsmarknads- och integrationsminister',    'Arbetsmarknadsdepartementet',                'A',  'L',  (SELECT intressent_id FROM politicians WHERE first_name='Johan'      AND last_name='Pehrson'             LIMIT 1), true),
  ('Lotta Edholm',            'Skolminister',                                'Utbildningsdepartementet',                   'U',  'L',  (SELECT intressent_id FROM politicians WHERE first_name='Lotta'      AND last_name='Edholm'              LIMIT 1), true),
  ('Mats Persson',            'Minister för högre utbildning och forskning', 'Utbildningsdepartementet',                   'U',  'L',  (SELECT intressent_id FROM politicians WHERE first_name='Mats'       AND last_name='Persson'             LIMIT 1), true),
  ('Parisa Liljestrand',      'Kulturminister',                              'Kulturdepartementet',                        'Ku', 'M',  (SELECT intressent_id FROM politicians WHERE first_name='Parisa'     AND last_name='Liljestrand'         LIMIT 1), true),
  ('Romina Pourmokhtari',     'Klimat- och miljöminister',                   'Klimat- och näringslivsdepartementet',       'N',  'L',  (SELECT intressent_id FROM politicians WHERE first_name='Romina'     AND last_name='Pourmokhtari'        LIMIT 1), true),
  ('Peter Kullgren',          'Landsbygdsminister',                          'Landsbygds- och infrastrukturdepartementet', 'LI', 'KD', (SELECT intressent_id FROM politicians WHERE first_name='Peter'      AND last_name='Kullgren'            LIMIT 1), true),
  ('Andreas Carlson',         'Infrastruktur- och bostadsminister',          'Landsbygds- och infrastrukturdepartementet', 'LI', 'KD', (SELECT intressent_id FROM politicians WHERE first_name='Andreas'    AND last_name='Carlson'             LIMIT 1), true),
  ('Erik Slottner',           'Civilminister',                               'Finansdepartementet',                        'Fi', 'KD', (SELECT intressent_id FROM politicians WHERE first_name='Erik'       AND last_name='Slottner'            LIMIT 1), true);
