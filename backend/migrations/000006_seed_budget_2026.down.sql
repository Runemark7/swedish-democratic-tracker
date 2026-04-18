DELETE FROM budget_allocations WHERE budget_year_id IN (SELECT id FROM budget_years WHERE year = 2026);
DELETE FROM budget_years WHERE year = 2026;
