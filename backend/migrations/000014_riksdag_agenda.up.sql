-- Policy agenda items displayed on the riksdag overview page.
-- Source: Tidöavtalet 2022 — https://www.regeringen.se/rattsliga-dokument/overenskommelse/2022/10/tidoavtalet/
-- Additional sources: Budgetpropositionen per year — https://www.regeringen.se/sveriges-regering/finansdepartementet/statsbudgeten/
-- UPDATE REQUIRED: manually verify each item against current government policy documents.
CREATE TABLE riksdag_agenda (
  id          SERIAL PRIMARY KEY,
  title       TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  source      TEXT NOT NULL DEFAULT '',
  status      TEXT NOT NULL CHECK (status IN ('active', 'in_progress', 'completed', 'cancelled')),
  sort_order  INTEGER NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO riksdag_agenda (title, description, source, status, sort_order) VALUES
(
  'Reformerad arbetslöshetsförsäkring',
  'Skärpta krav på aktivitet och tydligare matchning mot lediga tjänster. Ersättningsnivåer kopplas till tidigare inkomst med ett nytt trappstegssystem.',
  'Tidöavtalet 2022', 'in_progress', 1
),
(
  'Ny migrationslagstiftning',
  'Permanenta uppehållstillstånd ersätts av tidsbegränsade med krav på självförsörjning. Asylreglerna anpassas till EU:s miniminivå.',
  'Tidöavtalet 2022', 'active', 2
),
(
  'Skattesänkningar för arbete',
  'Jobbskatteavdraget utökas i tre steg under mandatperioden med fokus på låg- och medelinkomsttagare för att öka sysselsättningsgraden.',
  'Budgetpropositionen 2024', 'in_progress', 3
),
(
  'Höjd pensionsålder utreds',
  'En parlamentarisk kommission ska senast 2025 lägga fram förslag om gradvis höjd riktålder för pension i linje med ökad medellivslängd.',
  'Pensionsgruppens direktiv 2023', 'active', 4
),
(
  'Stärkt försvar mot 2030',
  'Försvarsanslaget når 2,5 % av BNP 2030. Värnplikt utökas och tre nya regementen återaktiveras för att möta förändrat säkerhetsläge.',
  'Försvarspropositionen 2024–2030', 'in_progress', 5
);
