-- Add 2026 decided budget data.
-- Source: Budgetpropositionen 2025/26:1, decided by Riksdagen 2025-11-26.
-- Amounts in KSEK from official government Excel specification.

INSERT INTO budget_years (year, status, total_ksek) VALUES
  (2026, 'decided', 1543317501)
ON CONFLICT (year, status) DO NOTHING;

INSERT INTO budget_allocations (budget_year_id, expenditure_area_id, amount_ksek, source)
SELECT by.id, ea.id, v.amount, 'government'
FROM (VALUES
  ('UO1',  21614324),
  ('UO2',  22038701),
  ('UO3',  15511898),
  ('UO4',  94743980),
  ('UO5',  2547401),
  ('UO6',  225022485),
  ('UO7',  43652932),
  ('UO8',  13524360),
  ('UO9',  127707301),
  ('UO10', 121994196),
  ('UO11', 59321791),
  ('UO12', 104228329),
  ('UO13', 5866226),
  ('UO14', 91162652),
  ('UO15', 35787594),
  ('UO16', 106936696),
  ('UO17', 17337204),
  ('UO18', 1775128),
  ('UO19', 4874201),
  ('UO20', 19625341),
  ('UO21', 7939815),
  ('UO22', 106704772),
  ('UO23', 22566654),
  ('UO24', 8955115),
  ('UO25', 180723354),
  ('UO26', 26955200),
  ('UO27', 54199851)
) AS v(code, amount)
JOIN expenditure_areas ea ON ea.code = v.code
CROSS JOIN budget_years by
WHERE by.year = 2026 AND by.status = 'decided'
ON CONFLICT DO NOTHING;
