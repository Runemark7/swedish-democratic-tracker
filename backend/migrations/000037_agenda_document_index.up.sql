-- The agenda becomes an index of whole documents.
--
-- It held five hand-picked points from documents like Tidöavtalet, each with a
-- `description` we paraphrased and a `status` we assigned ('active',
-- 'in_progress'). Three problems, decided in issue #102:
--
--   * Selecting five points out of a programme containing hundreds is an
--     editorial act. Selecting whole documents is far less bias-prone: the
--     reader gets the primary source and extracts from it themselves.
--   * The description was our wording, not the document's, with no verbatim
--     quote and no URL to check it against.
--   * `status` was a claim about the present -- how the sitting government is
--     progressing -- that was hand-set, never sourced, and had not been touched
--     since the rows were seeded on 2026-05-19.
--
-- What replaces it: title, issuer, publication date, link. Nothing else.
ALTER TABLE riksdag_agenda ADD COLUMN IF NOT EXISTS url       TEXT NOT NULL DEFAULT '';
ALTER TABLE riksdag_agenda ADD COLUMN IF NOT EXISTS issuer    TEXT NOT NULL DEFAULT '';
ALTER TABLE riksdag_agenda ADD COLUMN IF NOT EXISTS published DATE;

ALTER TABLE riksdag_agenda DROP CONSTRAINT IF EXISTS riksdag_agenda_status_check;
ALTER TABLE riksdag_agenda DROP COLUMN IF EXISTS status;
ALTER TABLE riksdag_agenda DROP COLUMN IF EXISTS description;

-- The five curated points are not documents and cannot be salvaged into them.
DELETE FROM riksdag_agenda;

-- Both URLs were checked to return 200 on 2026-08-11. The two URLs originally
-- written into the implementation plan both 404'd, which is why they are
-- verified here rather than trusted.
INSERT INTO riksdag_agenda (title, source, issuer, url, published, sort_order) VALUES
  ('Tidöavtalet – Överenskommelse för Sverige',
   'Tidöavtalet 2022',
   'Moderaterna, Kristdemokraterna, Liberalerna och Sverigedemokraterna',
   'https://www.liberalerna.se/wp-content/uploads/tidoavtalet-overenskommelse-for-sverige-slutlig.pdf',
   '2022-10-14', 1),
  ('Budgetpropositionen för 2026',
   'Prop. 2025/26:1',
   'Regeringen (Finansdepartementet)',
   'https://www.regeringen.se/rattsliga-dokument/proposition/2025/09/2025261/',
   '2025-09-22', 2);
