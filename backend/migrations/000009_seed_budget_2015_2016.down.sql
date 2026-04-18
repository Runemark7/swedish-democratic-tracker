-- Remove 2015-2016 budget data.

DELETE FROM budget_allocations
WHERE budget_year_id IN (
  SELECT id FROM budget_years
  WHERE year IN (2015, 2016) AND status = 'decided'
);

DELETE FROM budget_years WHERE year IN (2015, 2016);
