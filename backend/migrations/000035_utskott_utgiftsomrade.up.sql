-- Which utskott bereder which utgiftsområde.
--
-- Statutory, not our judgement: the allocation is the Bilaga (tilläggsbestämmelse
-- 7.5.1) to riksdagsordningen (2014:801). It is a strict partition — every
-- utgiftsområde has exactly one utskott and every utskott has at least one, with
-- no orphans either way.
--
-- Keyed on UO *code*, never on name: UO13 and UO20 have been renamed without the
-- allocation changing, so matching on name would silently drop rows.
--
-- Versioned by in-force date rather than per riksmöte. Attribution of a decision
-- is permanent (the beteckning encodes it), so a past vote never needs
-- re-mapping; only the description of what an utskott handles today changes.
CREATE TABLE IF NOT EXISTS utskott_utgiftsomrade (
  utskott_code  TEXT NOT NULL,
  uo_code       TEXT NOT NULL,
  in_force_from DATE NOT NULL,
  PRIMARY KEY (uo_code, in_force_from)
);

INSERT INTO utskott_utgiftsomrade (utskott_code, uo_code, in_force_from) VALUES
  ('KU',  'UO1',  '2014-09-01'),
  ('FiU', 'UO2',  '2014-09-01'),
  ('SkU', 'UO3',  '2014-09-01'),
  ('JuU', 'UO4',  '2014-09-01'),
  ('UU',  'UO5',  '2014-09-01'),
  ('FöU', 'UO6',  '2014-09-01'),
  ('UU',  'UO7',  '2014-09-01'),
  ('SfU', 'UO8',  '2014-09-01'),
  ('SoU', 'UO9',  '2014-09-01'),
  ('SfU', 'UO10', '2014-09-01'),
  ('SfU', 'UO11', '2014-09-01'),
  ('SfU', 'UO12', '2014-09-01'),
  ('AU',  'UO13', '2014-09-01'),
  ('AU',  'UO14', '2014-09-01'),
  ('UbU', 'UO15', '2014-09-01'),
  ('UbU', 'UO16', '2014-09-01'),
  ('KrU', 'UO17', '2014-09-01'),
  ('CU',  'UO18', '2014-09-01'),
  ('NU',  'UO19', '2014-09-01'),
  ('MJU', 'UO20', '2014-09-01'),
  ('NU',  'UO21', '2014-09-01'),
  ('TU',  'UO22', '2014-09-01'),
  ('MJU', 'UO23', '2014-09-01'),
  ('NU',  'UO24', '2014-09-01'),
  ('FiU', 'UO25', '2014-09-01'),
  ('FiU', 'UO26', '2014-09-01'),
  ('FiU', 'UO27', '2014-09-01')
ON CONFLICT (uo_code, in_force_from) DO NOTHING;
