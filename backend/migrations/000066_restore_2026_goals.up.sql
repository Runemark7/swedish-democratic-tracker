-- Restore the seven 2026 party goals dropped by 000064 (#164) as PROPOSED Löften.
--
-- 000057 archived the 2022 goals but never copied the 2026 ones, so 000064
-- left them only in the pre-deploy pg_dump (docs/runbooks/retire-goals.md).
--
-- State: visibility 'proposed', approved_at NULL. These goals never went
-- through the approval mechanism (#158), so they are not archived 2022
-- material and must not appear as approved: 'proposed' is invisible to every
-- public endpoint until the editor approves each one. Nothing here approves.
--
-- Ids 45-51 and created_at are kept. Mapping: text = goal_text, quote =
-- source_quote, source_url, source_document, topic, specificity, utskott =
-- relevant_committees. Keywords have no column in loften and are dropped.
-- mandate_period is '2026-2030' (the default of the write API for new Löften).
-- valmanifest_id is looked up (S, 2026, platform url); NULL if absent.
--
-- Idempotent: ON CONFLICT (id) DO NOTHING; the ändringslogg entry is written
-- only for rows this run inserted, and its request_key is unique.
WITH src (id, party, text, quote, source_url, source_document, topic, specificity, utskott, created_at) AS (
VALUES
  (45, 'S', 'Höja pensionerna för dem som slitit ett helt arbetsliv', 'Pensionen ska bli bättre för den som slitit ett helt arbetsliv.', 'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026', 'Valplattform 2026', 'pension', 'directional', ARRAY['SfU']::text[], '2026-07-29 21:22:46.382462+02'::timestamptz),
  (46, 'S', 'Återinföra kostnadsfri tandvård för unga upp till 23 år', 'Vi vill återinföra kostnadsfri tandvård för unga upp till 23 år (istället för dagens 19 år) och det förstärkta stödet till unga upp till 29 år, som regeringen valt att ta bort.', 'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026', 'Valplattform 2026', 'sjukvard', 'concrete', ARRAY['SoU']::text[], '2026-07-29 21:22:46.382462+02'::timestamptz),
  (47, 'S', 'Slopa karensavdraget för sjuka arbetstagare', 'Vi vill slopa det orättvisa karensavdraget som gör att hårt arbetande svenskar som sliter i vår välfärd och andra samhällsbärande yrken kan förlora tusenlappar vid en förkylning.', 'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026', 'Valplattform 2026', 'arbete', 'directional', ARRAY['AU','SoU']::text[], '2026-07-29 21:22:46.382462+02'::timestamptz),
  (48, 'S', 'Höja barn- och studiebidraget och sänka kostnader för hushållen', 'Vi ska ge tillbaka det du förlorat med ett plånbokslöfte: höjt barn- och studiebidrag, billigare mediciner, avgiftsfri tandvård för unga, hårdare press på banker och matkedjor och starkare skydd mot höga elpriser för villaägare.', 'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026', 'Valplattform 2026', 'valfard', 'directional', ARRAY['FiU','SfU']::text[], '2026-07-29 21:22:46.382462+02'::timestamptz),
  (49, 'S', 'Återinföra statlig byggstimulans för bostadsbyggande', 'Vi socialdemokrater föreslår därför att det införs en statlig byggstimulans i form av ett investeringsstöd, riktat till mindre orter med stora industrietableringar samt till ungdoms- och studentbostäder.', 'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026', 'Valplattform 2026', 'bostad', 'directional', ARRAY['CU']::text[], '2026-07-29 21:22:46.382462+02'::timestamptz),
  (50, 'S', 'Kriminalisera gängverksamhet med ny maffialag', 'En maffialag ska fängsla gängtopparna.', 'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026', 'Valplattform 2026', 'brott', 'directional', ARRAY['JuU']::text[], '2026-07-29 21:22:46.382462+02'::timestamptz),
  (51, 'S', 'Anställa 3 000 nya lärare och sätta tak på max tolv elever per lärare i lågstadiet', '3 000 nya lärare i skolan och max tolv elever per lärare i lågstadiet', 'https://www.socialdemokraterna.se/nyheter/nyheter/2026-05-01-s-gar-till-val-pa-smaskolereform-med-sma-klasser-for-sma-barn', 'Valplattform 2026', 'skola', 'concrete', ARRAY['UbU']::text[], '2026-07-29 21:22:46.382462+02'::timestamptz)
), ins AS (
  INSERT INTO loften (id, party, valmanifest_id, text, quote, source_url, source_document,
                      topic, specificity, utskott, mandate_period, visibility, created_at, approved_at)
  SELECT s.id, s.party, (SELECT v.id FROM valmanifest v WHERE v.party = 'S' AND v.election_year = 2026 AND v.url = 'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026' ORDER BY v.id LIMIT 1), s.text, s.quote, s.source_url, s.source_document,
         s.topic, s.specificity, s.utskott, '2026-2030', 'proposed', s.created_at, NULL
  FROM src s
  ON CONFLICT (id) DO NOTHING
  RETURNING *
)
INSERT INTO aendringslogg (actor, on_behalf_of, chat_ref, action, entity, entity_id, old_value, new_value, request_key, request_hash)
SELECT 'migration', 'restore-2026-goals', 'migration 000066', 'upsert', 'loefte', i.id, NULL,
       jsonb_build_object('id', i.id, 'party', i.party, 'valmanifestId', coalesce(i.valmanifest_id, 0),
         'text', i.text, 'quote', coalesce(i.quote, ''), 'locator', '', 'sourceUrl', coalesce(i.source_url, ''),
         'topic', coalesce(i.topic, ''), 'specificity', coalesce(i.specificity, ''), 'utskott', to_jsonb(i.utskott),
         'mandatePeriod', i.mandate_period, 'visibility', i.visibility),
       'restore-2026-goals-000066-' || i.id, 'migration-000066'
FROM ins i
ON CONFLICT (request_key) DO NOTHING;

SELECT setval('loften_id_seq', GREATEST((SELECT COALESCE(max(id), 0) FROM loften), 1));
