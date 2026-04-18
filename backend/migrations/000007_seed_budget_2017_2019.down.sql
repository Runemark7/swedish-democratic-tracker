DELETE FROM budget_allocations WHERE budget_year_id IN (SELECT id FROM budget_years WHERE year IN (2017, 2018, 2019));
DELETE FROM budget_years WHERE year IN (2017, 2018, 2019);
