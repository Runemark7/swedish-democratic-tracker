-- Enforce the no-orphan half of the partition that 000035 and
-- docs/data-sources/utskott-utgiftsomrade.md both already assert.
--
-- The allocation in the Bilaga to riksdagsordningen is keyed on UO *code*, and
-- nothing checked that the code exists. A typo'd uo_code ('UO31', 'U013') would
-- simply fail to join expenditure_areas, so a committee page would quietly show
-- one utgiftsområde fewer than the law gives it — a wrong page with no error
-- anywhere. The claim is only worth making if the database refuses to hold a
-- row that breaks it.
--
-- Added as a separate migration rather than by editing 000035, which is already
-- applied and would therefore never re-run.
ALTER TABLE utskott_utgiftsomrade
  ADD CONSTRAINT utskott_utgiftsomrade_uo_code_fkey
  FOREIGN KEY (uo_code) REFERENCES expenditure_areas (code);
