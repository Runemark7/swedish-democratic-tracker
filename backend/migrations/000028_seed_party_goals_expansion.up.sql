-- Expansion of curated party goals (Promise Tracking A).
-- Idempotent: ON CONFLICT (party, goal_text) DO NOTHING (constraint from 000002).
-- Every row carries source_url + verbatim source_quote. Quotes are
-- verbatim from each party's published document — never fabricated.
-- Existing rows from 000003 are untouched; this migration only adds new goals.

-- === S (Socialdemokraterna) — 7 new goals, 13 total ===
-- Source: Valplattform 2026 "Plan för Sverige", published 2026-02-05
INSERT INTO party_goals
  (party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees) VALUES
  ('S', 'Höja pensionerna för dem som slitit ett helt arbetsliv',
   'pension', 'directional', 'Valplattform 2026',
   'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026',
   'Pensionen ska bli bättre för den som slitit ett helt arbetsliv.',
   ARRAY['pension','pensionär','garantipension','pensionssystem','ålderspension'],
   ARRAY['SfU']),
  ('S', 'Återinföra kostnadsfri tandvård för unga upp till 23 år',
   'sjukvard', 'concrete', 'Valplattform 2026',
   'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026',
   'Vi vill återinföra kostnadsfri tandvård för unga upp till 23 år (istället för dagens 19 år) och det förstärkta stödet till unga upp till 29 år, som regeringen valt att ta bort.',
   ARRAY['tandvård','tandläkare','unga','kostnadsfri','sjukvård'],
   ARRAY['SoU']),
  ('S', 'Slopa karensavdraget för sjuka arbetstagare',
   'arbete', 'directional', 'Valplattform 2026',
   'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026',
   'Vi vill slopa det orättvisa karensavdraget som gör att hårt arbetande svenskar som sliter i vår välfärd och andra samhällsbärande yrken kan förlora tusenlappar vid en förkylning.',
   ARRAY['karensavdrag','karens','sjukskrivning','sjukpenning','arbete'],
   ARRAY['AU','SoU']),
  ('S', 'Höja barn- och studiebidraget och sänka kostnader för hushållen',
   'valfard', 'directional', 'Valplattform 2026',
   'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026',
   'Vi ska ge tillbaka det du förlorat med ett plånbokslöfte: höjt barn- och studiebidrag, billigare mediciner, avgiftsfri tandvård för unga, hårdare press på banker och matkedjor och starkare skydd mot höga elpriser för villaägare.',
   ARRAY['barnbidrag','studiebidrag','mediciner','elpriser','hushåll','plånbokslöfte'],
   ARRAY['FiU','SfU']),
  ('S', 'Återinföra statlig byggstimulans för bostadsbyggande',
   'bostad', 'directional', 'Valplattform 2026',
   'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026',
   'Vi socialdemokrater föreslår därför att det införs en statlig byggstimulans i form av ett investeringsstöd, riktat till mindre orter med stora industrietableringar samt till ungdoms- och studentbostäder.',
   ARRAY['bostad','bostadsbyggande','investeringsstöd','hyresrätt','studentbostad'],
   ARRAY['CU']),
  ('S', 'Kriminalisera gängverksamhet med ny maffialag',
   'brott', 'directional', 'Valplattform 2026',
   'https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026',
   'En maffialag ska fängsla gängtopparna.',
   ARRAY['maffialag','gäng','kriminalitet','organiserad brottslighet','fängelse','straff'],
   ARRAY['JuU']),
  ('S', 'Anställa 3 000 nya lärare och sätta tak på max tolv elever per lärare i lågstadiet',
   'skola', 'concrete', 'Valplattform 2026',
   'https://www.socialdemokraterna.se/nyheter/nyheter/2026-05-01-s-gar-till-val-pa-smaskolereform-med-sma-klasser-for-sma-barn',
   '3 000 nya lärare i skolan och max tolv elever per lärare i lågstadiet',
   ARRAY['lärare','elever','lågstadiet','småskolereform','skola','klassrum'],
   ARRAY['UbU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- === M (Moderaterna) — 4 new goals, 10 total ===
-- Source: Valmanifest 2022 "Så får vi ordning på Sverige"
INSERT INTO party_goals
  (party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees) VALUES
  ('M', 'Införa bidragstak för att göra det alltid lönsamt att arbeta',
   'arbete', 'directional', 'Valmanifest 2022',
   'https://moderaterna.se/app/uploads/2022/08/moderaternas_valmanifest_2022_webversion.pdf',
   'Vi vill införa ett bidragstak så att det alltid lönar sig bättre att ta ett jobb än att stapla olika bidrag på varandra',
   ARRAY['bidragstak','bidrag','arbetsincitament','sysselsättning','lön','försörjningsstöd'],
   ARRAY['AU','FiU']),
  ('M', 'Avskaffa begränsningarna för ny kärnkraft och tillåta fler reaktorer',
   'energi', 'concrete', 'Valmanifest 2022',
   'https://moderaterna.se/app/uploads/2022/08/moderaternas_valmanifest_2022_webversion.pdf',
   'Ta bort förbudet mot att ha fler än tio kärnkraftsreaktorer i drift och tillåt att kärnkraftsreaktorer byggs på fler platser än de befintliga.',
   ARRAY['kärnkraft','reaktor','energi','el','elproduktion','atomkraft','deregulering'],
   ARRAY['NU']),
  ('M', 'Anpassa asylmottagandet till EU:s miniminivå',
   'invandring', 'directional', 'Valmanifest 2022',
   'https://moderaterna.se/app/uploads/2022/08/moderaternas_valmanifest_2022_webversion.pdf',
   'För att asylmottagandet ska motsvara Sveriges integrationsförmåga krävs en migrationspolitik utifrån EU:s miniminivå och internationella konventionsåtaganden',
   ARRAY['asyl','migration','invandring','eu','miniminivå','uppehållstillstånd','integration'],
   ARRAY['SfU']),
  ('M', 'Förbereda skolan för yrkesliv och stärka studiero',
   'skola', 'directional', 'Valmanifest 2022',
   'https://moderaterna.se/app/uploads/2022/08/moderaternas_valmanifest_2022_webversion.pdf',
   'Skolan ska förbereda för ett yrkesliv – och forskningen stå fri',
   ARRAY['skola','utbildning','studiero','ordning','yrkesliv','betyg','lärare'],
   ARRAY['UbU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- === SD (Sverigedemokraterna) — 4 new goals, 10 total ===
-- Source: Valmanifest 2022
INSERT INTO party_goals
  (party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees) VALUES
  ('SD', 'Höja pensionerna för alla och mest för dem som arbetat',
   'pension', 'directional', 'Valmanifest 2022',
   'https://www.sd.se/sverigedemokraternas-valmanifest-2022/',
   'Sverigedemokraterna vill höja pensionerna för alla och mest för dem som arbetat.',
   ARRAY['pension','pensionär','garantipension','inkomstpension','ålderspension'],
   ARRAY['SfU']),
  ('SD', 'Sänka skatten på pension och öka pensionsavsättningarna',
   'pension', 'directional', 'Valmanifest 2022',
   'https://www.sd.se/sverigedemokraternas-valmanifest-2022/',
   'Skatten på pension ska sänkas och avsättningarna inom den allmänna pensionen ses över i syfte att kännbart öka pensionerna.',
   ARRAY['pensionsskatt','pension','skattelättnad','avsättning','garantipension'],
   ARRAY['SkU','SfU']),
  ('SD', 'Förstatliga A-kassan och stärka skyddet vid arbetslöshet',
   'arbete', 'directional', 'Valmanifest 2022',
   'https://www.sd.se/sverigedemokraternas-valmanifest-2022/',
   'Tillfällig arbetslöshet ska inte behöva bli en privatekonomisk kris. Därför ska tryggheten för den som drabbas av arbetslöshet öka och fler kunna få ersättning vid arbetslöshet genom att A-kassan fullt ut inryms i det gemensamma välfärdsåtagandet.',
   ARRAY['a-kassa','arbetslöshet','arbetslöshetsersättning','förstatliga','trygghet','arbete'],
   ARRAY['AU','SfU']),
  ('SD', 'Höja polisernas löner och införa betald polisutbildning',
   'brott', 'directional', 'Valmanifest 2022',
   'https://www.sd.se/sverigedemokraternas-valmanifest-2022/',
   'Sverigedemokraterna vill höja polisernas löner, införa en betald polisutbildning, förbättra arbetsvillkor och arbetsmiljö, avlasta poliserna med civilanställda, ge polisen förutsättningarna för att bekämpa brottslighet, och skapa trygghet och säkerhet för poliser och deras familjer.',
   ARRAY['polis','polislön','polisutbildning','trygghet','brottslighet','rekrytering'],
   ARRAY['JuU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- === C (Centerpartiet) — 4 new goals, 9 total ===
-- Source: Valmanifest 2022 "För Sveriges bästa"
INSERT INTO party_goals
  (party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees) VALUES
  ('C', 'Avskaffa amorteringskravet för att underlätta bostadsmarknaden',
   'bostad', 'directional', 'Valmanifest 2022',
   'https://www.centerpartiet.se/download/18.5a7401c51829f250987e4f/1660832711101/Valmanifest%202022%20Utskrift%20ny.pdf',
   'Amorteringskravet – ett påtvingat sparande för dem med lägst inkomster – bör avskaffas när ränteläget nu normaliseras.',
   ARRAY['amorteringskrav','bostad','bostadsmarknad','bolån','ränta','bostadsköp'],
   ARRAY['CU','FiU']),
  ('C', 'Sänka det totala skattetrycket och skapa stabila statsfinanser',
   'skatt', 'directional', 'Valmanifest 2022',
   'https://www.centerpartiet.se/download/18.5a7401c51829f250987e4f/1660832711101/Valmanifest%202022%20Utskrift%20ny.pdf',
   'Vi ska sänka det totala skattetrycket och skapa stabilitet i statsfinanserna genom att hushålla med skattebetalarnas pengar.',
   ARRAY['skatt','skattetryck','statsfinans','skattepolitik','skattelättnad'],
   ARRAY['FiU','SkU']),
  ('C', 'Skapa bättre förutsättningar med lägre skatt och mindre regelkrångel',
   'foretagande', 'directional', 'Valmanifest 2022',
   'https://www.centerpartiet.se/download/18.5a7401c51829f250987e4f/1660832711101/Valmanifest%202022%20Utskrift%20ny.pdf',
   'Vi ska skapa bättre förutsättningar för dem att lyckas, med lägre skatt, mindre regelkrångel och en starkare arbetslinje.',
   ARRAY['regelkrångel','företag','skatt','arbetsmarknad','näringsliv','tillväxt'],
   ARRAY['NU','SkU']),
  ('C', 'Frigöra landsbygdens enorma potential',
   'landsbygd', 'directional', 'Valmanifest 2022',
   'https://www.centerpartiet.se/download/18.5a7401c51829f250987e4f/1660832711101/Valmanifest%202022%20Utskrift%20ny.pdf',
   'Det är dags att frigöra landsbygdens enorma potential.',
   ARRAY['landsbygd','glesbygd','regional','infrastruktur','bredband','lantbruk'],
   ARRAY['CU','NU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- === V (Vänsterpartiet) — 3 new goals, 9 total ===
-- Source: Valplattform 2022
INSERT INTO party_goals
  (party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees) VALUES
  ('V', 'Införa tandvårdsförsäkring med högkostnadsskydd',
   'sjukvard', 'directional', 'Valplattform 2022',
   'https://www.vansterpartiet.se/wp-content/uploads/2022/07/valplattform-2022.pdf',
   'Vänsterpartiet vill därför införa en tandvårdsförsäkring liknande den som finns inom övrig hälso- och sjukvård med ett högkostnadsskydd som gör att vuxna människor med behov inte ska behöva tacka nej till god tandvård.',
   ARRAY['tandvård','tandläkare','högkostnadsskydd','sjukvård','försäkring'],
   ARRAY['SoU']),
  ('V', 'Avgiftsfri förskola för familjer med inkomster upp till 20 000 kronor',
   'skola', 'concrete', 'Valplattform 2022',
   'https://www.vansterpartiet.se/wp-content/uploads/2022/07/valplattform-2022.pdf',
   'Avgiftsfri förskola för hushåll med inkomster på upp till 20 000 kronor i månaden och sänk avgiften för hushåll som tjänar mindre än 72 400 kronor per månad',
   ARRAY['förskola','barnomsorg','avgiftsfri','barnfamilj','maxtaxa'],
   ARRAY['UbU','SfU']),
  ('V', 'Återinföra förmögenhetsskatt på förmögenheter över 5 miljoner kronor',
   'skatt', 'concrete', 'Valplattform 2022',
   'https://www.vansterpartiet.se/wp-content/uploads/2022/07/valplattform-2022.pdf',
   'en återinförd förmögenhetsskatt på förmögenheter över 5 miljoner kronor',
   ARRAY['förmögenhetsskatt','kapitalskatt','skatt','ojämlikhet','förmögenhet'],
   ARRAY['SkU','FiU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- === KD (Kristdemokraterna) — 5 new goals, 10 total ===
-- Source: Valmanifest 2022 "Redo för en ny regering"
INSERT INTO party_goals
  (party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees) VALUES
  ('KD', 'Inför fast läkarkontakt med namngiven läkare och tak på antal patienter',
   'sjukvard', 'directional', 'Valmanifest 2022',
   'https://kristdemokraterna.se/download/18.2e2b5b3e18180017d1be0d/1660223521596/Valmanifest2022-KD.pdf',
   'Inför fast läkarkontakt med en namngiven läkare och tak för antal patienter för att möjliggöra relationsbyggande mellan patient och läkare',
   ARRAY['läkarkontakt','primärvård','vårdcentral','patient','husläkare','vård'],
   ARRAY['SoU']),
  ('KD', 'Höja garantipensionen och bostadstillägget för pensionärer',
   'pension', 'directional', 'Valmanifest 2022',
   'https://kristdemokraterna.se/download/18.2e2b5b3e18180017d1be0d/1660223521596/Valmanifest2022-KD.pdf',
   'Höjd garantipension och höjt bostadstillägg för pensionärer',
   ARRAY['garantipension','bostadstillägg','pension','pensionär','äldre'],
   ARRAY['SfU']),
  ('KD', 'Bryta äldres ofrivilliga ensamhet med äldresamtal i kommunerna',
   'aldre', 'directional', 'Valmanifest 2022',
   'https://kristdemokraterna.se/download/18.2e2b5b3e18180017d1be0d/1660223521596/Valmanifest2022-KD.pdf',
   'Bryt äldres ofrivilliga ensamhet och isolering genom äldresamtal i alla kommuner, gemenskapsfrämjande insatser i omsorgen samt digitala hjälpmedel',
   ARRAY['äldre','ensamhet','isolering','äldreomsorg','hemtjänst','gemenskap'],
   ARRAY['SoU']),
  ('KD', 'Införa en flexibel föräldraförsäkring för att stödja familjerna',
   'familj', 'directional', 'Valmanifest 2022',
   'https://kristdemokraterna.se/download/18.2e2b5b3e18180017d1be0d/1660223521596/Valmanifest2022-KD.pdf',
   'Vi vill införa en flexibel föräldraförsäkring – för att stödja familjerna.',
   ARRAY['föräldraförsäkring','familj','föräldrar','barn','föräldraledighet'],
   ARRAY['SfU']),
  ('KD', 'Sätta tak för barngruppsstorlek i förskolan och tillämpa närhetsprincipen',
   'skola', 'directional', 'Valmanifest 2022',
   'https://kristdemokraterna.se/download/18.2e2b5b3e18180017d1be0d/1660223521596/Valmanifest2022-KD.pdf',
   'Vi vill införa ett tak för hur stora barngrupper det ska få finnas på förskolan - och att skolplatser ska erbjudas utifrån närhetsprincipen',
   ARRAY['förskola','barngrupp','barnomsorg','skola','närhetsprincip','barn'],
   ARRAY['UbU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- === L (Liberalerna) — 4 new goals, 9 total ===
-- Source: Valmanifest 2022 "Maktskifte för ett nytt Sverige"
INSERT INTO party_goals
  (party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees) VALUES
  ('L', 'Prioritera kunskap i skolan för att hindra brott och skapa jobb',
   'skola', 'directional', 'Valmanifest 2022',
   'https://www.liberalerna.se/wp-content/uploads/valmanifest-2022.pdf',
   'Vi är partiet som vet att mer kunskap i skolan i dag hindrar brotten och skapar jobben i morgon.',
   ARRAY['skola','kunskap','utbildning','segregation','betyg','lärare'],
   ARRAY['UbU']),
  ('L', 'Möta klimathotet med teknik och framtidstro istället för förbud',
   'klimat', 'directional', 'Valmanifest 2022',
   'https://www.liberalerna.se/wp-content/uploads/valmanifest-2022.pdf',
   'Och vi är partiet som vet att klimathotet måste mötas med teknik och framtidstro.',
   ARRAY['klimat','teknik','klimatomställning','innovation','grön','utsläpp'],
   ARRAY['MJU','NU']),
  ('L', 'Bygga nya kärnkraftverk och skapa ett elnät för all grön el',
   'energi', 'directional', 'Valmanifest 2022',
   'https://www.liberalerna.se/wp-content/uploads/valmanifest-2022.pdf',
   'Med en omställning med nya kärnkraftverk och en autobahn för all ny grön el som förbättrar överföringen i hela Sverige – i stället för nedlagd kärnkraft och rekord-höga elpriser.',
   ARRAY['kärnkraft','elnät','energi','elöverföring','elpriser','grön el'],
   ARRAY['NU']),
  ('L', 'Avlasta lärare och ge dem ordentlig lön och betald fortbildning',
   'skola', 'directional', 'Valmanifest 2022',
   'https://www.liberalerna.se/wp-content/uploads/valmanifest-2022.pdf',
   'Lärare ska avlastas för att kunna vara just lärare, kunna göra karriär, få betald fortbildning och få ordentligt bra betalt.',
   ARRAY['lärare','lärarlön','fortbildning','karriär','skola','avlastning'],
   ARRAY['UbU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- === MP (Miljöpartiet) — 5 new goals, 10 total ===
-- Source: Valmanifest 2022 "Alla ska med när Sverige ställer om"
INSERT INTO party_goals
  (party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees) VALUES
  ('MP', 'Satsa 100 miljarder per år på den gröna omställningen',
   'klimat', 'concrete', 'Valmanifest 2022',
   'https://www.mp.se/wp-content/uploads/2022/08/valmanifest2022_hela_a4_klart.pdf',
   'Miljöpartiet vill satsa 100 miljarder om året det kommande decenniet på den gröna omställningen.',
   ARRAY['klimat','grön omställning','klimatinvestering','klimatbudget','fossilfri'],
   ARRAY['MJU','FiU']),
  ('MP', 'Skydda 30 procent av Sveriges natur och återställa våtmarker',
   'miljo', 'concrete', 'Valmanifest 2022',
   'https://www.mp.se/wp-content/uploads/2022/08/valmanifest2022_hela_a4_klart.pdf',
   'Vi vill skydda 30 procent av Sveriges natur, fjällskogar och skyddsvärda skogar i hela landet och återställa våtmarker.',
   ARRAY['naturskydd','biologisk mångfald','skog','fjäll','våtmark','artskydd'],
   ARRAY['MJU']),
  ('MP', 'Göra tåget billigare än flyget och stärka nattågen i hela Sverige',
   'transport', 'directional', 'Valmanifest 2022',
   'https://www.mp.se/wp-content/uploads/2022/08/valmanifest2022_hela_a4_klart.pdf',
   'Vi vill att tåget ska vara billigare än flyget, nattågen ska stärkas och säkras i hela Sverige, och kollektivtrafiken ska göras billigare och byggas ut.',
   ARRAY['tåg','flyg','nattåg','järnväg','kollektivtrafik','transport'],
   ARRAY['TU']),
  ('MP', 'Kräva att nyproducerade bilar är helt fossilfria från 2025',
   'klimat', 'concrete', 'Valmanifest 2022',
   'https://www.mp.se/wp-content/uploads/2022/08/valmanifest2022_hela_a4_klart.pdf',
   'Miljöpartiet vill att nyproducerade bilar som säljs i Sverige ska vara helt fossilfria från 2025.',
   ARRAY['fossilfri','bil','elfordon','fordonsflotta','utsläpp','klimat'],
   ARRAY['MJU','TU']),
  ('MP', 'Finansiera klimatåtgärder med progressiv kapitalbeskattning',
   'skatt', 'directional', 'Valmanifest 2022',
   'https://www.mp.se/wp-content/uploads/2022/08/valmanifest2022_hela_a4_klart.pdf',
   'de med störst resurser, som också orsakar större utsläpp, ska bidra mer till omställningen och välfärden genom progressiv kapitalbeskattning',
   ARRAY['kapitalbeskattning','skatt','klimat','ojämlikhet','förmögenhet','omställning'],
   ARRAY['SkU','MJU'])
ON CONFLICT (party, goal_text) DO NOTHING;
