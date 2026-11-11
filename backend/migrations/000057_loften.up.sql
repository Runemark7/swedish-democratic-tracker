-- Valmanifest and Löfte as public, read-only data (issue #156, spec #129).
--
-- ADDITIVE ONLY. party_goals, goal_vote_matches and promises are left exactly
-- as they are; #164 retires them later. The 2022 goals are copied into loften
-- with their ids kept, so an old goal URL maps one-to-one onto a Löfte.
--
-- Visibility is the only gate on what is public:
--   proposed / struck  never public
--   approved           public (nothing in this migration creates one)
--   archived           public, shown without a Löftesstatus
-- No approval is created here.

CREATE TABLE valmanifest (
  id            SERIAL PRIMARY KEY,
  party         TEXT NOT NULL REFERENCES parties (code),
  election_year INT  NOT NULL,
  title         TEXT NOT NULL,
  published_on  DATE,                       -- not known for the seeded catalogue
  url           TEXT NOT NULL,
  pdf_url       TEXT,
  state         TEXT NOT NULL CHECK (state IN ('draft', 'final')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (party, election_year, title)
);
CREATE INDEX idx_valmanifest_party ON valmanifest (party, election_year DESC);

-- Seeded from the catalogue that was hard-coded in the frontend
-- (ManifestosPage.tsx). "published" there is "final" here; "draft" is the
-- party's own marking (utkast, preliminär).
INSERT INTO valmanifest (party, election_year, title, url, pdf_url, state) VALUES
  ('C', 2026, 'Sverige kan mer – Centerpartiets valmanifest 2026', 'https://val2026.centerpartiet.se/', NULL, 'final'),
  ('L', 2026, 'För din frihet – Liberalernas valmanifest 2026', 'https://www.liberalerna.se/pressmeddelanden/liberalerna-presenterar-valmanifest-infor-valet-2026-for-din-frihet', NULL, 'final'),
  ('MP', 2026, 'Miljöpartiets valmanifest 2026 (utkast)', 'https://www.mp.se/valmanifest-2026-utkast/', NULL, 'draft'),
  ('S', 2026, 'Plan för Sverige – Socialdemokraternas valplattform 2026', 'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026', NULL, 'final'),
  ('SD', 2026, 'Valplattform 2026', 'https://www.sd.se/wp-content/uploads/2026/07/valplattform-2026.pdf', 'https://www.sd.se/wp-content/uploads/2026/07/valplattform-2026.pdf', 'final'),
  ('V', 2026, 'Preliminär valplattform 2026, efter beslut på kongressen', 'https://www.vansterpartiet.se/wp-content/uploads/2026/04/Preliminar-Valplattform-efter-beslut-pa-kongressen-2026.pdf', 'https://www.vansterpartiet.se/wp-content/uploads/2026/04/Preliminar-Valplattform-efter-beslut-pa-kongressen-2026.pdf', 'draft'),
  ('S', 2022, 'Vårt Sverige kan bättre – Socialdemokraternas valmanifest 2022', 'https://www.socialdemokraterna.se/5.36cb33a81817fcdd3ec4a8d.html', 'https://snd.gu.se/sv/vivill/party/s/v/2022', 'final'),
  ('M', 2022, 'Så får vi ordning på Sverige – Moderaternas valmanifest 2022', 'https://moderaterna.se/nyhet/moderaternasvalmanifest/', 'https://moderaterna.se/app/uploads/2022/08/moderaternas_valmanifest_2022_webversion.pdf', 'final'),
  ('SD', 2022, 'Sverigedemokraternas valmanifest 2022', 'https://sd.se/wp-content/uploads/2022/08/valmanifest.pdf', 'https://sd.se/wp-content/uploads/2022/08/valmanifest.pdf', 'final'),
  ('C', 2022, 'Centerpartiets valmanifest 2022', 'https://snd.gu.se/sv/vivill/party/c/v/2022', 'https://snd.gu.se/sv/vivill/party/c/v/2022', 'final'),
  ('V', 2022, 'Vänsterpartiets valplattform 2022', 'https://snd.gu.se/sv/vivill/party/v/v/2022', 'https://snd.gu.se/sv/vivill/party/v/v/2022', 'final'),
  ('KD', 2022, 'Kristdemokraternas valmanifest 2022', 'https://snd.gu.se/sv/vivill/party/kd/v/2022', 'https://snd.gu.se/sv/vivill/party/kd/v/2022', 'final'),
  ('L', 2022, 'Liberalernas valmanifest 2022', 'https://snd.gu.se/sv/vivill/party/l/v/2022', 'https://snd.gu.se/sv/vivill/party/l/v/2022', 'final'),
  ('MP', 2022, 'Alla ska med när Sverige ställer om – Miljöpartiets valmanifest 2022', 'https://snd.gu.se/sv/vivill/party/mp/v/2022', 'https://snd.gu.se/sv/vivill/party/mp/v/2022', 'final'),
  ('S', 2018, 'Det största trygghetsprogrammet i modern tid – Valmanifest 2018', 'https://snd.gu.se/sv/vivill/party/s/v/2018', 'https://snd.gu.se/sv/vivill/party/s/v/2018', 'final'),
  ('M', 2018, 'Moderaternas valmanifest 2018', 'https://snd.gu.se/sv/vivill/party/m/v/2018', 'https://snd.gu.se/sv/vivill/party/m/v/2018', 'final'),
  ('SD', 2018, 'Sverigedemokraternas valmanifest 2018', 'https://snd.gu.se/sv/vivill/party/sd/v/2018', 'https://snd.gu.se/sv/vivill/party/sd/v/2018', 'final'),
  ('C', 2018, 'Centerpartiets valmanifest 2018', 'https://snd.gu.se/sv/vivill/party/c/v/2018', 'https://snd.gu.se/sv/vivill/party/c/v/2018', 'final'),
  ('V', 2018, 'Vänsterpartiets valplattform 2018', 'https://snd.gu.se/sv/vivill/party/v/v/2018', 'https://snd.gu.se/sv/vivill/party/v/v/2018', 'final'),
  ('KD', 2018, 'Kristdemokraternas valmanifest 2018', 'https://snd.gu.se/sv/vivill/party/kd/v/2018', 'https://snd.gu.se/sv/vivill/party/kd/v/2018', 'final'),
  ('L', 2018, 'Liberalernas valmanifest 2018', 'https://snd.gu.se/sv/vivill/party/l/v/2018', 'https://snd.gu.se/sv/vivill/party/l/v/2018', 'final'),
  ('MP', 2018, 'Miljöpartiets valmanifest 2018', 'https://snd.gu.se/sv/vivill/party/mp/v/2018', 'https://snd.gu.se/sv/vivill/party/mp/v/2018', 'final'),
  ('S', 2014, 'Socialdemokraternas valmanifest 2014', 'https://snd.gu.se/sv/vivill/party/s/v/2014', NULL, 'final'),
  ('M', 2014, 'Moderaternas valmanifest 2014', 'https://snd.gu.se/sv/vivill/party/m/v/2014', NULL, 'final'),
  ('SD', 2014, 'Sverigedemokraternas valmanifest 2014', 'https://snd.gu.se/sv/vivill/party/sd/v/2014', NULL, 'final'),
  ('C', 2014, 'Centerpartiets valmanifest 2014', 'https://snd.gu.se/sv/vivill/party/c/v/2014', NULL, 'final'),
  ('V', 2014, 'Vänsterpartiets valplattform 2014', 'https://snd.gu.se/sv/vivill/party/v/v/2014', NULL, 'final'),
  ('KD', 2014, 'Kristdemokraternas valmanifest 2014', 'https://snd.gu.se/sv/vivill/party/kd/v/2014', NULL, 'final'),
  ('L', 2014, 'Folkpartiets valmanifest 2014', 'https://snd.gu.se/sv/vivill/party/l/v/2014', NULL, 'final'),
  ('MP', 2014, 'Miljöpartiets valmanifest 2014', 'https://snd.gu.se/sv/vivill/party/mp/v/2014', NULL, 'final')
ON CONFLICT DO NOTHING;

CREATE TABLE loften (
  id              INTEGER PRIMARY KEY,
  party           TEXT NOT NULL REFERENCES parties (code),
  valmanifest_id  INTEGER REFERENCES valmanifest (id),
  text            TEXT NOT NULL,            -- the testable claim, close to the source
  quote           TEXT,                     -- exact quote; required once approved
  locator         TEXT,                     -- page number or anchor
  source_url      TEXT,
  source_document TEXT,                     -- legacy label, kept for archived rows
  topic           TEXT,
  specificity     TEXT CHECK (specificity IN ('concrete', 'directional', 'rhetorical')),
  utskott         TEXT[] NOT NULL DEFAULT '{}',
  mandate_period  TEXT NOT NULL,            -- mandate_periods.code; no FK, that table is rewritten by tests and tooling
  visibility      TEXT NOT NULL CHECK (visibility IN ('proposed', 'approved', 'struck', 'archived')),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  approved_at     TIMESTAMPTZ,
  CONSTRAINT loften_approved_is_checkable CHECK (
    visibility <> 'approved'
    OR (coalesce(quote, '') <> '' AND coalesce(source_url, '') <> '')
  )
);
CREATE INDEX idx_loften_party_visibility ON loften (party, visibility);

CREATE SEQUENCE loften_id_seq OWNED BY loften.id;
ALTER TABLE loften ALTER COLUMN id SET DEFAULT nextval('loften_id_seq');

-- Archive the 2022 goals, ids kept. The same rule as the record view
-- (election_cycle NULL or not 2026): 2026 campaign material is not a promise
-- of the 2022-2026 period and is not copied.
INSERT INTO loften (id, party, valmanifest_id, text, quote, source_url, source_document,
                    topic, specificity, utskott, mandate_period, visibility, created_at)
SELECT g.id, g.party,
       (SELECT v.id FROM valmanifest v
         WHERE v.party = g.party AND v.election_year = 2022 AND g.election_cycle = '2022'
         ORDER BY v.id LIMIT 1),
       g.goal_text, g.source_quote, g.source_url, g.source_document,
       g.topic, g.specificity, g.relevant_committees, '2022-2026', 'archived', g.created_at
FROM party_goals g
WHERE g.election_cycle IS NULL OR g.election_cycle <> '2026'
ON CONFLICT (id) DO NOTHING;

SELECT setval('loften_id_seq', GREATEST((SELECT COALESCE(max(id), 0) FROM loften),
                                        (SELECT COALESCE(max(id), 0) FROM party_goals), 1));
