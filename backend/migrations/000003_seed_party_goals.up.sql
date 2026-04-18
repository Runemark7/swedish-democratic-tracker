-- Party goals from Swedish party platforms (Valmanifest 2022, Tidöavtalet, partiprogram)
-- Idempotent: uses ON CONFLICT on (party, goal_text) unique constraint from migration 000002.

-- S (Socialdemokraterna)
INSERT INTO party_goals (party, goal_text, topic, specificity, source_document, keywords, relevant_committees) VALUES
  ('S', 'Korta vårdköerna och anställa fler inom sjukvården',
   'sjukvard', 'directional', 'Valmanifest 2022',
   ARRAY['vård','sjukvård','vårdkö','hälso','1177','sjukhus'],
   ARRAY['SoU']),
  ('S', 'Stärka den offentliga skolan och öka likvärdigheten',
   'skola', 'directional', 'Valmanifest 2022',
   ARRAY['skola','utbildning','lärare','likvärdig','elev','undervisning'],
   ARRAY['UbU']),
  ('S', 'Höja lägstalönerna och stärka löntagarnas ställning',
   'arbete', 'directional', 'Valmanifest 2022',
   ARRAY['lön','kollektivavtal','arbetsmarknad','fack','anställning'],
   ARRAY['AU']),
  ('S', 'Bygga fler hyresrätter med rimliga hyror',
   'bostad', 'directional', 'Valmanifest 2022',
   ARRAY['bostad','hyresrätt','hyra','bostadsbyggande','allmännytta'],
   ARRAY['CU']),
  ('S', 'Genomföra en klimatomställning med rättvisa',
   'klimat', 'directional', 'Valmanifest 2022',
   ARRAY['klimat','utsläpp','fossilfri','energi','miljö','grön'],
   ARRAY['MJU']),
  ('S', 'Stärka välfärden och minska privatiseringen av offentlig service',
   'valfard', 'directional', 'Valmanifest 2022',
   ARRAY['välfärd','privatisering','offentlig','kommun','region','vinstjakt'],
   ARRAY['FiU','SoU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- M (Moderaterna)
INSERT INTO party_goals (party, goal_text, topic, specificity, source_document, keywords, relevant_committees) VALUES
  ('M', 'Sänka skatten för alla som arbetar',
   'skatt', 'directional', 'Valmanifest 2022',
   ARRAY['skatt','skattelättnad','jobbskatteavdrag','skattesänkning','inkomstskatt'],
   ARRAY['FiU','SkU']),
  ('M', 'Skärpa straffen för gängkriminalitet och grova brott',
   'brott', 'directional', 'Tidöavtalet 2022',
   ARRAY['brott','straff','kriminalitet','gäng','polis','fängelse','trygghet'],
   ARRAY['JuU']),
  ('M', 'Öka arbetslinjen och minska bidragsberoendet',
   'arbete', 'directional', 'Valmanifest 2022',
   ARRAY['bidrag','arbete','sysselsättning','arbetsmarknad','arbetslinje'],
   ARRAY['AU','SfU']),
  ('M', 'Stärka det militära försvaret och uppfylla NATO-åtaganden',
   'forsvar', 'directional', 'Tidöavtalet 2022',
   ARRAY['försvar','NATO','militär','försvarsbudget','värnplikt'],
   ARRAY['FöU']),
  ('M', 'Bygga ut kärnkraften för stabil energiförsörjning',
   'energi', 'directional', 'Tidöavtalet 2022',
   ARRAY['kärnkraft','energi','el','elproduktion','atom','reaktor'],
   ARRAY['NU','MJU']),
  ('M', 'Korta vårdköerna genom privata och offentliga vårdgivare',
   'sjukvard', 'directional', 'Valmanifest 2022',
   ARRAY['vård','sjukvård','vårdkö','vårdgivare','patient','sjukhus'],
   ARRAY['SoU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- SD (Sverigedemokraterna)
INSERT INTO party_goals (party, goal_text, topic, specificity, source_document, keywords, relevant_committees) VALUES
  ('SD', 'Minska invandringen till EU:s miniminivåer',
   'invandring', 'concrete', 'Valmanifest 2022',
   ARRAY['invandring','asyl','migration','flyktingpolitik','uppehållstillstånd'],
   ARRAY['SfU']),
  ('SD', 'Kraftigt skärpa straffen för gängkriminella',
   'brott', 'directional', 'Tidöavtalet 2022',
   ARRAY['brott','straff','gäng','kriminalitet','polis','fängelse','trygghet'],
   ARRAY['JuU']),
  ('SD', 'Stärka den svenska välfärden för svenska medborgare',
   'valfard', 'directional', 'Valmanifest 2022',
   ARRAY['välfärd','medborgare','pension','sjukvård','äldre'],
   ARRAY['SoU','SfU']),
  ('SD', 'Satsa på kärnkraft och avveckla vindkraftssubventioner',
   'energi', 'directional', 'Valmanifest 2022',
   ARRAY['kärnkraft','energi','vindkraft','elproduktion','el'],
   ARRAY['NU','MJU']),
  ('SD', 'Stärka det svenska försvaret',
   'forsvar', 'directional', 'Tidöavtalet 2022',
   ARRAY['försvar','militär','försvarsbudget','NATO','värnplikt'],
   ARRAY['FöU']),
  ('SD', 'Införa hårdare krav på integration och språkkunskaper',
   'invandring', 'directional', 'Tidöavtalet 2022',
   ARRAY['integration','svenska','språkkrav','medborgarskap','samhällsorientering'],
   ARRAY['SfU','AU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- C (Centerpartiet)
INSERT INTO party_goals (party, goal_text, topic, specificity, source_document, keywords, relevant_committees) VALUES
  ('C', 'Stärka landsbygden och decentralisera statliga jobb',
   'landsbygd', 'directional', 'Valmanifest 2022',
   ARRAY['landsbygd','regional','decentralisering','kommun','glesbygd'],
   ARRAY['CU','NU']),
  ('C', 'Underlätta för småföretagare och sänka arbetsgivaravgifter',
   'foretagande', 'directional', 'Valmanifest 2022',
   ARRAY['företag','småföretag','arbetsgivaravgift','entreprenör','näringsverksamhet'],
   ARRAY['NU','SkU']),
  ('C', 'Driva på den gröna omställningen med marknadslösningar',
   'klimat', 'directional', 'Valmanifest 2022',
   ARRAY['klimat','grön','omställning','fossilfri','förnybar','miljö'],
   ARRAY['MJU','NU']),
  ('C', 'Värna individens frihet och den personliga integriteten',
   'frihet', 'directional', 'Partiprogram',
   ARRAY['frihet','integritet','individ','rättigheter','övervakning'],
   ARRAY['KU','JuU']),
  ('C', 'Reformera skolan med fokus på kvalitet och valfrihet',
   'skola', 'directional', 'Valmanifest 2022',
   ARRAY['skola','utbildning','lärare','valfrihet','friskola'],
   ARRAY['UbU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- V (Vänsterpartiet)
INSERT INTO party_goals (party, goal_text, topic, specificity, source_document, keywords, relevant_committees) VALUES
  ('V', 'Införa sex timmars arbetsdag med bibehållen lön',
   'arbete', 'concrete', 'Valmanifest 2022',
   ARRAY['arbetstid','arbetstidsförkortning','sex timmar','arbetsdag'],
   ARRAY['AU']),
  ('V', 'Stoppa vinstuttag i välfärden och förstatliga skolan',
   'valfard', 'directional', 'Valmanifest 2022',
   ARRAY['vinst','välfärd','privatisering','förstatliga','skola','friskola'],
   ARRAY['UbU','FiU']),
  ('V', 'Genomföra en rättvis klimatomställning som minskar ojämlikhet',
   'klimat', 'directional', 'Valmanifest 2022',
   ARRAY['klimat','utsläpp','ojämlikhet','miljö','fossilfri','rättvisa'],
   ARRAY['MJU']),
  ('V', 'Bygga 50 000 nya hyresrätter per år',
   'bostad', 'concrete', 'Valmanifest 2022',
   ARRAY['hyresrätt','bostad','bostadsbyggande','allmännytta','hyra'],
   ARRAY['CU']),
  ('V', 'Höja taken i sjukförsäkringen och stärka anställningstryggheten',
   'arbete', 'directional', 'Valmanifest 2022',
   ARRAY['sjukförsäkring','anställningstrygghet','LAS','sjukpenning','arbetsmiljö'],
   ARRAY['AU','SfU']),
  ('V', 'Stärka den offentliga sjukvården och avskaffa marknadsstyrningen',
   'sjukvard', 'directional', 'Valmanifest 2022',
   ARRAY['sjukvård','vård','offentlig','marknad','patient','region'],
   ARRAY['SoU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- KD (Kristdemokraterna)
INSERT INTO party_goals (party, goal_text, topic, specificity, source_document, keywords, relevant_committees) VALUES
  ('KD', 'Stärka familjepolitiken och höja barnbidraget',
   'familj', 'directional', 'Valmanifest 2022',
   ARRAY['familj','barnbidrag','föräldrar','barn','föräldraförsäkring'],
   ARRAY['SfU']),
  ('KD', 'Förbättra äldreomsorgen och ge de äldre ett värdigt liv',
   'aldre', 'directional', 'Valmanifest 2022',
   ARRAY['äldre','äldreomsorg','hemtjänst','pension','särskilt boende'],
   ARRAY['SoU']),
  ('KD', 'Stärka försvaret och fullt ut stödja NATO-medlemskapet',
   'forsvar', 'directional', 'Tidöavtalet 2022',
   ARRAY['försvar','NATO','militär','försvarsbudget','värnplikt'],
   ARRAY['FöU']),
  ('KD', 'Sänka skatten för pensionärer',
   'skatt', 'directional', 'Valmanifest 2022',
   ARRAY['skatt','pensionär','pension','skattesänkning','inkomst'],
   ARRAY['SkU','SfU']),
  ('KD', 'Bekämpa gängkriminalitet med hårdare straff',
   'brott', 'directional', 'Tidöavtalet 2022',
   ARRAY['brott','straff','gäng','kriminalitet','polis','trygghet'],
   ARRAY['JuU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- L (Liberalerna)
INSERT INTO party_goals (party, goal_text, topic, specificity, source_document, keywords, relevant_committees) VALUES
  ('L', 'Satsa på skolan med fokus på kunskap och ordning',
   'skola', 'directional', 'Valmanifest 2022',
   ARRAY['skola','utbildning','lärare','kunskap','elev','undervisning'],
   ARRAY['UbU']),
  ('L', 'Stärka rättsstaten och domstolarnas oberoende',
   'rattsstat', 'directional', 'Valmanifest 2022',
   ARRAY['rättsstat','domstol','lag','rättssäkerhet','grundlag'],
   ARRAY['JuU','KU']),
  ('L', 'Fullt ut stödja NATO och stärka det svenska försvaret',
   'forsvar', 'directional', 'Tidöavtalet 2022',
   ARRAY['försvar','NATO','militär','försvarsbudget','säkerhetspolitik'],
   ARRAY['FöU']),
  ('L', 'Förbättra integrationspolitiken med krav och möjligheter',
   'invandring', 'directional', 'Tidöavtalet 2022',
   ARRAY['integration','svenska','språkkrav','arbetsmarknad','nyanländ'],
   ARRAY['SfU','AU']),
  ('L', 'Sänka skatterna för att stimulera tillväxt och företagande',
   'skatt', 'directional', 'Valmanifest 2022',
   ARRAY['skatt','tillväxt','företag','skattesänkning','ekonomi'],
   ARRAY['FiU','SkU'])
ON CONFLICT (party, goal_text) DO NOTHING;

-- MP (Miljöpartiet)
INSERT INTO party_goals (party, goal_text, topic, specificity, source_document, keywords, relevant_committees) VALUES
  ('MP', 'Nå netto-noll utsläpp i Sverige senast 2035',
   'klimat', 'concrete', 'Valmanifest 2022',
   ARRAY['klimat','utsläpp','fossilfri','koldioxid','miljö','netto-noll'],
   ARRAY['MJU']),
  ('MP', 'Skydda den biologiska mångfalden och stärka naturvården',
   'miljo', 'directional', 'Valmanifest 2022',
   ARRAY['biologisk','mångfald','naturvård','artskydd','skog','hav','natur'],
   ARRAY['MJU']),
  ('MP', 'Bygga ut kollektivtrafiken och minska bilberoendet',
   'transport', 'directional', 'Valmanifest 2022',
   ARRAY['kollektivtrafik','tåg','järnväg','bil','transport','cykel'],
   ARRAY['TU']),
  ('MP', 'Stärka Sveriges klimatbistånd och internationella ansvar',
   'klimat', 'directional', 'Valmanifest 2022',
   ARRAY['bistånd','klimat','internationell','global','utveckling'],
   ARRAY['UU','MJU']),
  ('MP', 'Fasa ut alla fossila subventioner',
   'klimat', 'concrete', 'Valmanifest 2022',
   ARRAY['fossil','subvention','diesel','bensin','klimat','skattelättnad'],
   ARRAY['SkU','MJU'])
ON CONFLICT (party, goal_text) DO NOTHING;
