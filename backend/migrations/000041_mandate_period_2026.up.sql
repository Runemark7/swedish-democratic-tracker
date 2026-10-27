-- The 2026–2030 parliament.
--
-- 000033 and 000038 stop at the election that closed 2022–2026. Without this
-- row, votes and speeches from riksmöte 2026/27 join no period: the chamber
-- that opened after the 13 September 2026 election is stored, then read as if
-- it were still the previous one.
--
-- start_date and end_date are the election days that open and close the
-- period, the same convention as 000033. General elections are held on the
-- second Sunday of September (Regeringsformen, vallagen), which is
-- 13 September 2026 and 8 September 2030.
--
-- The four riksmöten are the labels Riksdagen files documents under. Inserting
-- them does not fetch anything; ingestion reads this table to decide which
-- sessions belong to the period. A session that has not opened yet is simply
-- empty upstream.
INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
VALUES ('2026-2030', 'Mandatperioden 2026–2030', '2026-09-13', '2030-09-08',
        ARRAY['2026/27', '2027/28', '2028/29', '2029/30'])
ON CONFLICT (code) DO NOTHING;
