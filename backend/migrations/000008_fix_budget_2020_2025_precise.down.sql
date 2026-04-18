-- Revert to approximate 2020-2025 values from migration 000005.

DELETE FROM budget_allocations
WHERE budget_year_id IN (
  SELECT id FROM budget_years
  WHERE year IN (2020, 2021, 2022, 2023, 2024, 2025) AND status = 'decided'
) AND source = 'government';

UPDATE budget_years SET total_ksek = 1026700000 WHERE year = 2020 AND status = 'decided';
UPDATE budget_years SET total_ksek = 1108800000 WHERE year = 2021 AND status = 'decided';
UPDATE budget_years SET total_ksek = 1105300000 WHERE year = 2022 AND status = 'decided';
UPDATE budget_years SET total_ksek = 1191400000 WHERE year = 2023 AND status = 'decided';
UPDATE budget_years SET total_ksek = 1244600000 WHERE year = 2024 AND status = 'decided';
UPDATE budget_years SET total_ksek = 1297300000 WHERE year = 2025 AND status = 'decided';

-- 2020 (approximate)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  16450000), ('UO2',  1820000),  ('UO3',  12450000), ('UO4',  52400000),
  ('UO5',  2090000),  ('UO6',  62200000), ('UO7',  46100000), ('UO8',  12800000),
  ('UO9',  88700000), ('UO10', 101500000),('UO11', 38100000), ('UO12', 97800000),
  ('UO13', 7840000),  ('UO14', 77300000), ('UO15', 24800000), ('UO16', 83500000),
  ('UO17', 16300000), ('UO18', 5200000),  ('UO19', 4180000),  ('UO20', 12800000),
  ('UO21', 4350000),  ('UO22', 63300000), ('UO23', 20700000), ('UO24', 7600000),
  ('UO25', 137500000),('UO26', 18300000), ('UO27', 32600000)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2020 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- 2021 (approximate)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  17100000), ('UO2',  1950000),  ('UO3',  12800000), ('UO4',  55200000),
  ('UO5',  2150000),  ('UO6',  69500000), ('UO7',  48000000), ('UO8',  11400000),
  ('UO9',  97800000), ('UO10', 107200000),('UO11', 37600000), ('UO12', 100500000),
  ('UO13', 9300000),  ('UO14', 95100000), ('UO15', 28700000), ('UO16', 87200000),
  ('UO17', 17100000), ('UO18', 7800000),  ('UO19', 4500000),  ('UO20', 16200000),
  ('UO21', 4900000),  ('UO22', 68400000), ('UO23', 21300000), ('UO24', 9500000),
  ('UO25', 146200000),('UO26', 16800000), ('UO27', 36600000)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2021 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- 2022 (approximate)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  17400000), ('UO2',  2000000),  ('UO3',  13000000), ('UO4',  58500000),
  ('UO5',  2200000),  ('UO6',  78600000), ('UO7',  43800000), ('UO8',  10500000),
  ('UO9',  97000000), ('UO10', 103400000),('UO11', 36800000), ('UO12', 101200000),
  ('UO13', 8400000),  ('UO14', 82700000), ('UO15', 27100000), ('UO16', 89400000),
  ('UO17', 17500000), ('UO18', 5900000),  ('UO19', 4300000),  ('UO20', 14700000),
  ('UO21', 5100000),  ('UO22', 71200000), ('UO23', 22100000), ('UO24', 8300000),
  ('UO25', 149100000),('UO26', 22100000), ('UO27', 43900000)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2022 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- 2023 (approximate)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  18100000), ('UO2',  2100000),  ('UO3',  13400000), ('UO4',  65800000),
  ('UO5',  2300000),  ('UO6',  105600000),('UO7',  42900000), ('UO8',  11200000),
  ('UO9',  101400000),('UO10', 110800000),('UO11', 36200000), ('UO12', 104300000),
  ('UO13', 9100000),  ('UO14', 80200000), ('UO15', 26800000), ('UO16', 92100000),
  ('UO17', 17800000), ('UO18', 4600000),  ('UO19', 4400000),  ('UO20', 12600000),
  ('UO21', 5500000),  ('UO22', 76900000), ('UO23', 23600000), ('UO24', 8700000),
  ('UO25', 153400000),('UO26', 33500000), ('UO27', 48200000)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2023 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- 2024 (approximate)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  18800000), ('UO2',  2200000),  ('UO3',  13800000), ('UO4',  72300000),
  ('UO5',  2400000),  ('UO6',  126400000),('UO7',  41500000), ('UO8',  12000000),
  ('UO9',  105800000),('UO10', 116100000),('UO11', 35800000), ('UO12', 107200000),
  ('UO13', 9600000),  ('UO14', 85200000), ('UO15', 27800000), ('UO16', 95600000),
  ('UO17', 18200000), ('UO18', 4200000),  ('UO19', 4500000),  ('UO20', 11900000),
  ('UO21', 5800000),  ('UO22', 80500000), ('UO23', 24300000), ('UO24', 9100000),
  ('UO25', 158300000),('UO26', 28200000), ('UO27', 53800000)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2024 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- 2025 (approximate)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  19200000), ('UO2',  2300000),  ('UO3',  14200000), ('UO4',  78100000),
  ('UO5',  2500000),  ('UO6',  145200000),('UO7',  39800000), ('UO8',  11800000),
  ('UO9',  110300000),('UO10', 120500000),('UO11', 35400000), ('UO12', 110100000),
  ('UO13', 10100000), ('UO14', 79100000), ('UO15', 28500000), ('UO16', 98200000),
  ('UO17', 18600000), ('UO18', 4400000),  ('UO19', 4600000),  ('UO20', 13200000),
  ('UO21', 6100000),  ('UO22', 84600000), ('UO23', 25100000), ('UO24', 9500000),
  ('UO25', 163400000),('UO26', 21500000), ('UO27', 47000000)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2025 AND by.status = 'decided'
ON CONFLICT DO NOTHING;
