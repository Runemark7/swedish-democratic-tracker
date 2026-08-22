-- The five mandate periods before 2022-2026.
--
-- 000033 created the table with one row, because one period was all the record
-- held. The unit was settled separately: a mandate period is a boundary the
-- constitution defines -- elections set it, we do not -- so any other window
-- ("since 2010", "last ten years") would be our editorial pick rather than
-- Sweden's. All six available periods are therefore held, none is promoted,
-- and the reader chooses which to read.
--
-- Six, not more: Riksdagen's votering data begins at 2002/03. Riksmöten from
-- 2001/02 back return zero, so 2002-2006 is the earliest period there is a
-- record to hold.
--
-- start_date and end_date are the election days that open and close each
-- period, matching 000033's convention -- so one period's end_date is the next
-- one's start_date. Swedish general elections fell on the third Sunday of
-- September through 2010 and the second Sunday from 2014.
--
-- Inserting the rows does not ingest the votes. It declares which riksmöten
-- belong to which period, which is what cmd/backfill reads to know what to
-- fetch and what the coverage queries scope against.
INSERT INTO mandate_periods (code, label, start_date, end_date, riksmoten)
VALUES
  ('2002-2006', 'Mandatperioden 2002–2006', '2002-09-15', '2006-09-17',
   ARRAY['2002/03', '2003/04', '2004/05', '2005/06']),
  ('2006-2010', 'Mandatperioden 2006–2010', '2006-09-17', '2010-09-19',
   ARRAY['2006/07', '2007/08', '2008/09', '2009/10']),
  ('2010-2014', 'Mandatperioden 2010–2014', '2010-09-19', '2014-09-14',
   ARRAY['2010/11', '2011/12', '2012/13', '2013/14']),
  ('2014-2018', 'Mandatperioden 2014–2018', '2014-09-14', '2018-09-09',
   ARRAY['2014/15', '2015/16', '2016/17', '2017/18']),
  ('2018-2022', 'Mandatperioden 2018–2022', '2018-09-09', '2022-09-11',
   ARRAY['2018/19', '2019/20', '2020/21', '2021/22'])
ON CONFLICT (code) DO NOTHING;
