-- Remove only rows added by migration 000027.
-- Keyed on exact (party, goal_text) pairs — never touches rows from 000003.
DELETE FROM party_goals WHERE (party, goal_text) IN (
  -- S (Socialdemokraterna)
  ('S', 'Höja pensionerna för dem som slitit ett helt arbetsliv'),
  ('S', 'Återinföra kostnadsfri tandvård för unga upp till 23 år'),
  ('S', 'Slopa karensavdraget för sjuka arbetstagare'),
  ('S', 'Höja barn- och studiebidraget och sänka kostnader för hushållen'),
  ('S', 'Återinföra statlig byggstimulans för bostadsbyggande'),
  ('S', 'Kriminalisera gängverksamhet med ny maffialag'),
  ('S', 'Anställa 3 000 nya lärare och sätta tak på max tolv elever per lärare i lågstadiet'),
  -- M (Moderaterna)
  ('M', 'Införa bidragstak för att göra det alltid lönsamt att arbeta'),
  ('M', 'Avskaffa begränsningarna för ny kärnkraft och tillåta fler reaktorer'),
  ('M', 'Anpassa asylmottagandet till EU:s miniminivå'),
  ('M', 'Förbereda skolan för yrkesliv och stärka studiero'),
  -- SD (Sverigedemokraterna)
  ('SD', 'Höja pensionerna för alla och mest för dem som arbetat'),
  ('SD', 'Sänka skatten på pension och öka pensionsavsättningarna'),
  ('SD', 'Förstatliga A-kassan och stärka skyddet vid arbetslöshet'),
  ('SD', 'Höja polisernas löner och införa betald polisutbildning'),
  -- C (Centerpartiet)
  ('C', 'Avskaffa amorteringskravet för att underlätta bostadsmarknaden'),
  ('C', 'Sänka det totala skattetrycket och skapa stabila statsfinanser'),
  ('C', 'Skapa bättre förutsättningar med lägre skatt och mindre regelkrångel'),
  ('C', 'Frigöra landsbygdens enorma potential'),
  -- V (Vänsterpartiet)
  ('V', 'Införa tandvårdsförsäkring med högkostnadsskydd'),
  ('V', 'Avgiftsfri förskola för familjer med inkomster upp till 20 000 kronor'),
  ('V', 'Återinföra förmögenhetsskatt på förmögenheter över 5 miljoner kronor'),
  -- KD (Kristdemokraterna)
  ('KD', 'Inför fast läkarkontakt med namngiven läkare och tak på antal patienter'),
  ('KD', 'Höja garantipensionen och bostadstillägget för pensionärer'),
  ('KD', 'Bryta äldres ofrivilliga ensamhet med äldresamtal i kommunerna'),
  ('KD', 'Införa en flexibel föräldraförsäkring för att stödja familjerna'),
  ('KD', 'Sätta tak för barngruppsstorlek i förskolan och tillämpa närhetsprincipen'),
  -- L (Liberalerna)
  ('L', 'Prioritera kunskap i skolan för att hindra brott och skapa jobb'),
  ('L', 'Möta klimathotet med teknik och framtidstro istället för förbud'),
  ('L', 'Bygga nya kärnkraftverk och skapa ett elnät för all grön el'),
  ('L', 'Avlasta lärare och ge dem ordentlig lön och betald fortbildning'),
  -- MP (Miljöpartiet)
  ('MP', 'Satsa 100 miljarder per år på den gröna omställningen'),
  ('MP', 'Skydda 30 procent av Sveriges natur och återställa våtmarker'),
  ('MP', 'Göra tåget billigare än flyget och stärka nattågen i hela Sverige'),
  ('MP', 'Kräva att nyproducerade bilar är helt fossilfria från 2025'),
  ('MP', 'Finansiera klimatåtgärder med progressiv kapitalbeskattning')
);
