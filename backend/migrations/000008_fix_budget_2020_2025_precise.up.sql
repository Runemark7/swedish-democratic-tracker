-- Replace approximate 2020-2025 budget data with precise values from official government Excel files.
-- Source: regeringen.se "Specifikation av budgetens utgifter och inkomster" per year.
-- Amounts in KSEK (thousands of kronor).

-- ── Step 1: Delete approximate allocations ────────────────────────────

DELETE FROM budget_allocations
WHERE budget_year_id IN (
  SELECT id FROM budget_years
  WHERE year IN (2020, 2021, 2022, 2023, 2024, 2025) AND status = 'decided'
) AND source = 'government';

-- ── Step 2: Update year totals to precise values ──────────────────────

UPDATE budget_years SET total_ksek = 1062385804 WHERE year = 2020 AND status = 'decided';
UPDATE budget_years SET total_ksek = 1165822520 WHERE year = 2021 AND status = 'decided';
UPDATE budget_years SET total_ksek = 1204219603 WHERE year = 2022 AND status = 'decided';
UPDATE budget_years SET total_ksek = 1251870933 WHERE year = 2023 AND status = 'decided';
UPDATE budget_years SET total_ksek = 1346694314 WHERE year = 2024 AND status = 'decided';
UPDATE budget_years SET total_ksek = 1441595890 WHERE year = 2025 AND status = 'decided';

-- ── Step 3: Re-insert with precise allocations ────────────────────────

-- 2020 (Source: Prop. 2019/20:1)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  15131274),  ('UO2',  17390191),  ('UO3',  12056989),  ('UO4',  51731092),
  ('UO5',  2028315),   ('UO6',  64799753),  ('UO7',  45989153),  ('UO8',  11446213),
  ('UO9',  84167142),  ('UO10', 95705843),  ('UO11', 36542396),  ('UO12', 101430320),
  ('UO13', 10065663),  ('UO14', 77164287),  ('UO15', 25508400),  ('UO16', 83315865),
  ('UO17', 16061280),  ('UO18', 3726522),   ('UO19', 3672525),   ('UO20', 12571183),
  ('UO21', 3468932),   ('UO22', 61295946),  ('UO23', 19790236),  ('UO24', 7263863),
  ('UO25', 128417572), ('UO26', 29655200),  ('UO27', 41989649)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2020 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- 2021 (Source: Prop. 2020/21:1)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  15907564),  ('UO2',  17501844),  ('UO3',  12413921),  ('UO4',  56430577),
  ('UO5',  1999265),   ('UO6',  71153370),  ('UO7',  46828602),  ('UO8',  9268385),
  ('UO9',  101941904), ('UO10', 98410400),  ('UO11', 37715888),  ('UO12', 103578007),
  ('UO13', 7790741),   ('UO14', 105724319), ('UO15', 28313111),  ('UO16', 92077789),
  ('UO17', 17790875),  ('UO18', 5578679),   ('UO19', 4173425),   ('UO20', 16202213),
  ('UO21', 4514129),   ('UO22', 73916227),  ('UO23', 22581344),  ('UO24', 15355253),
  ('UO25', 153452049), ('UO26', 165200),    ('UO27', 45037439)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2021 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- 2022 (Source: Prop. 2021/22:1)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  17268338),  ('UO2',  17971183),  ('UO3',  12729734),  ('UO4',  61688986),
  ('UO5',  2235117),   ('UO6',  76525799),  ('UO7',  51939762),  ('UO8',  8456364),
  ('UO9',  112483613), ('UO10', 97721457),  ('UO11', 41786128),  ('UO12', 103040770),
  ('UO13', 6319015),   ('UO14', 94025726),  ('UO15', 27801485),  ('UO16', 94529445),
  ('UO17', 18454593),  ('UO18', 7025334),   ('UO19', 5241801),   ('UO20', 21851580),
  ('UO21', 4454924),   ('UO22', 76964552),  ('UO23', 22044754),  ('UO24', 9344901),
  ('UO25', 152322194), ('UO26', 12155200),  ('UO27', 47836848)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2022 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- 2023 (Source: Prop. 2022/23:1)
-- Note: UO1 total computed from sub-category sum (Excel file omits UO-level header row).
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  18270004),  ('UO2',  19021399),  ('UO3',  12923291),  ('UO4',  68305854),
  ('UO5',  2094669),   ('UO6',  93952930),  ('UO7',  47206227),  ('UO8',  16019745),
  ('UO9',  110429732), ('UO10', 106427304), ('UO11', 55394136),  ('UO12', 105159584),
  ('UO13', 5565206),   ('UO14', 90103686),  ('UO15', 27912395),  ('UO16', 94991971),
  ('UO17', 16667948),  ('UO18', 6099158),   ('UO19', 4534201),   ('UO20', 19542691),
  ('UO21', 4944849),   ('UO22', 78952417),  ('UO23', 19373352),  ('UO24', 11407773),
  ('UO25', 157545359), ('UO26', 13155200),  ('UO27', 45869852)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2023 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- 2024 (Source: Prop. 2023/24:1)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  19070363),  ('UO2',  20881058),  ('UO3',  16854357),  ('UO4',  76028013),
  ('UO5',  2273924),   ('UO6',  126059689), ('UO7',  48630129),  ('UO8',  13808274),
  ('UO9',  110258264), ('UO10', 116907208), ('UO11', 60310483),  ('UO12', 106070732),
  ('UO13', 4018388),   ('UO14', 92059399),  ('UO15', 30545170),  ('UO16', 99461047),
  ('UO17', 16646285),  ('UO18', 6081906),   ('UO19', 3922201),   ('UO20', 19307359),
  ('UO21', 5604320),   ('UO22', 82874980),  ('UO23', 23973181),  ('UO24', 9552305),
  ('UO25', 174280270), ('UO26', 20455200),  ('UO27', 40759809)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2024 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- 2025 (Source: Prop. 2024/25:1)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  20130981),  ('UO2',  21734078),  ('UO3',  14812398),  ('UO4',  86791679),
  ('UO5',  2278371),   ('UO6',  169680344), ('UO7',  44499844),  ('UO8',  11937532),
  ('UO9',  120254872), ('UO10', 123058927), ('UO11', 62890360),  ('UO12', 104531449),
  ('UO13', 6265767),   ('UO14', 93539399),  ('UO15', 33793466),  ('UO16', 103845141),
  ('UO17', 16861894),  ('UO18', 3240094),   ('UO19', 4294201),   ('UO20', 16438897),
  ('UO21', 6631297),   ('UO22', 94447432),  ('UO23', 21697804),  ('UO24', 8315298),
  ('UO25', 173107254), ('UO26', 28755200),  ('UO27', 47761911)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2025 AND by.status = 'decided'
ON CONFLICT DO NOTHING;
