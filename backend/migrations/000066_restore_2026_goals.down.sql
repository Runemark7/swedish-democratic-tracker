-- Remove exactly the Löften restored by 000066, and only while each is still
-- un-approved and unchanged (proposed, approved_at NULL, same text and quote)
-- and no koppling points at it. Other rows are never touched. The append-only
-- ändringslogg entries stay, as that table cannot be edited.
DELETE FROM loften l
WHERE l.id IN (45,46,47,48,49,50,51)
  AND l.visibility = 'proposed'
  AND l.approved_at IS NULL
  AND l.party = 'S'
  AND l.mandate_period = '2026-2030'
  AND l.source_document = 'Valplattform 2026'
  AND NOT EXISTS (SELECT 1 FROM kopplingar k WHERE k.loefte_id = l.id)
  AND (l.id, l.text, l.quote) IN (
    (45, 'Höja pensionerna för dem som slitit ett helt arbetsliv', 'Pensionen ska bli bättre för den som slitit ett helt arbetsliv.'),
    (46, 'Återinföra kostnadsfri tandvård för unga upp till 23 år', 'Vi vill återinföra kostnadsfri tandvård för unga upp till 23 år (istället för dagens 19 år) och det förstärkta stödet till unga upp till 29 år, som regeringen valt att ta bort.'),
    (47, 'Slopa karensavdraget för sjuka arbetstagare', 'Vi vill slopa det orättvisa karensavdraget som gör att hårt arbetande svenskar som sliter i vår välfärd och andra samhällsbärande yrken kan förlora tusenlappar vid en förkylning.'),
    (48, 'Höja barn- och studiebidraget och sänka kostnader för hushållen', 'Vi ska ge tillbaka det du förlorat med ett plånbokslöfte: höjt barn- och studiebidrag, billigare mediciner, avgiftsfri tandvård för unga, hårdare press på banker och matkedjor och starkare skydd mot höga elpriser för villaägare.'),
    (49, 'Återinföra statlig byggstimulans för bostadsbyggande', 'Vi socialdemokrater föreslår därför att det införs en statlig byggstimulans i form av ett investeringsstöd, riktat till mindre orter med stora industrietableringar samt till ungdoms- och studentbostäder.'),
    (50, 'Kriminalisera gängverksamhet med ny maffialag', 'En maffialag ska fängsla gängtopparna.'),
    (51, 'Anställa 3 000 nya lärare och sätta tak på max tolv elever per lärare i lågstadiet', '3 000 nya lärare i skolan och max tolv elever per lärare i lågstadiet')
  );
