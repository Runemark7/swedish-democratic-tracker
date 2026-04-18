-- Historical budget data 2017-2019.
-- Source: Budgetpropositionen for each year, official Excel from regeringen.se.
-- Amounts in KSEK (tusental kronor).

-- ── Budget Years ──────────────────────────────────────────────────────

INSERT INTO budget_years (year, status, total_ksek) VALUES
  (2017, 'decided', 971169281),
  (2018, 'decided', 1000514994),
  (2019, 'decided', 1026634185)
ON CONFLICT (year, status) DO NOTHING;

-- ── 2017 ──────────────────────────────────────────────────────────────

INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  13268476),
  ('UO2',  15258562),
  ('UO3',  10985564),
  ('UO4',  42466641),
  ('UO5',  1913291),
  ('UO6',  50254364),
  ('UO7',  34990005),
  ('UO8',  32580319),
  ('UO9',  68496164),
  ('UO10', 107051805),
  ('UO11', 34774451),
  ('UO12', 89491398),
  ('UO13', 32635380),
  ('UO14', 75656514),
  ('UO15', 22433999),
  ('UO16', 72381263),
  ('UO17', 14521254),
  ('UO18', 6764560),
  ('UO19', 3595481),
  ('UO20', 8403539),
  ('UO21', 2876898),
  ('UO22', 55114917),
  ('UO23', 17189401),
  ('UO24', 6456020),
  ('UO25', 105554920),
  ('UO26', 16467790),
  ('UO27', 29586305)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2017 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- ── 2018 ───────��────────────────────────────────────────────────────���─

INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  14531335),
  ('UO2',  16285547),
  ('UO3',  11399384),
  ('UO4',  45776783),
  ('UO5',  2009032),
  ('UO6',  53835520),
  ('UO7',  42985209),
  ('UO8',  15748230),
  ('UO9',  77696041),
  ('UO10', 102615229),
  ('UO11', 34635669),
  ('UO12', 94586418),
  ('UO13', 25600209),
  ('UO14', 74083840),
  ('UO15', 24352692),
  ('UO16', 77965638),
  ('UO17', 15879752),
  ('UO18', 6951648),
  ('UO19', 3921525),
  ('UO20', 10772788),
  ('UO21', 3588354),
  ('UO22', 56418505),
  ('UO23', 19253779),
  ('UO24', 7370651),
  ('UO25', 111385385),
  ('UO26', 11355200),
  ('UO27', 39510631)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2018 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- ── 2019 ───────────────────────────────────────────────���──────────────

INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  14901645),
  ('UO2',  17315466),
  ('UO3',  11402533),
  ('UO4',  46342859),
  ('UO5',  2362277),
  ('UO6',  56190704),
  ('UO7',  44945140),
  ('UO8',  13573430),
  ('UO9',  78122185),
  ('UO10', 97657012),
  ('UO11', 34450476),
  ('UO12', 97332203),
  ('UO13', 18486483),
  ('UO14', 76814772),
  ('UO15', 25022956),
  ('UO16', 81347211),
  ('UO17', 15994531),
  ('UO18', 6972123),
  ('UO19', 3496525),
  ('UO20', 11836248),
  ('UO21', 3594947),
  ('UO22', 59456655),
  ('UO23', 19787281),
  ('UO24', 7266717),
  ('UO25', 115892772),
  ('UO26', 25155200),
  ('UO27', 40913834)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2019 AND by.status = 'decided'
ON CONFLICT DO NOTHING;
