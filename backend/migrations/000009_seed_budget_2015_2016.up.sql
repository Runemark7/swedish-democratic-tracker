-- Add 2015-2016 precise budget data.
-- Source: Official government Excel specifications from regeringen.se.
-- Amounts in KSEK (thousands of kronor).

INSERT INTO budget_years (year, status, total_ksek) VALUES
  (2015, 'decided', 890900386),
  (2016, 'decided', 926249927)
ON CONFLICT (year, status) DO NOTHING;

-- 2015 (Source: Prop. 2014/15:1)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  12413383),  ('UO2',  14589790),  ('UO3',  10580976),  ('UO4',  40757864),
  ('UO5',  1899062),   ('UO6',  48588647),  ('UO7',  30009056),  ('UO8',  17433309),
  ('UO9',  64441159),  ('UO10', 101016487), ('UO11', 38166036),  ('UO12', 83908698),
  ('UO13', 16807043),  ('UO14', 71846204),  ('UO15', 21177307),  ('UO16', 64152606),
  ('UO17', 13150941),  ('UO18', 1508973),   ('UO19', 2700801),   ('UO20', 6881418),
  ('UO21', 2495609),   ('UO22', 50078203),  ('UO23', 15787422),  ('UO24', 5366162),
  ('UO25', 94490852),  ('UO26', 20526200),  ('UO27', 40126178)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2015 AND by.status = 'decided'
ON CONFLICT DO NOTHING;

-- 2016 (Source: Prop. 2015/16:1)
INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  12717244),  ('UO2',  14812312),  ('UO3',  10782946),  ('UO4',  41573737),
  ('UO5',  1905203),   ('UO6',  48827432),  ('UO7',  32357474),  ('UO8',  19419719),
  ('UO9',  69237676),  ('UO10', 109868281), ('UO11', 36184250),  ('UO12', 87129050),
  ('UO13', 21070265),  ('UO14', 79681475),  ('UO15', 21707932),  ('UO16', 69452400),
  ('UO17', 13694535),  ('UO18', 7064024),   ('UO19', 3255721),   ('UO20', 7661756),
  ('UO21', 2812038),   ('UO22', 54122036),  ('UO23', 18919553),  ('UO24', 5998273),
  ('UO25', 93398252),  ('UO26', 10769176),  ('UO27', 31827167)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2016 AND by.status = 'decided'
ON CONFLICT DO NOTHING;
