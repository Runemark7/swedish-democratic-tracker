---
id: valmyndigheten-riksdagsval
name: Valmyndigheten — riksdagsmandat 2026
kind: seed
upstream: https://resultat.val.se/val2026/
license: Valmyndighetens publicerade valresultat — offentlig handling
freshness: 4-årscykel (val). Raden byts när en regering har utsetts.
last_verified: 2026-10-27
verification_status: verified
verification_notes: ""
used_by:
  - /riksdag (kammarens mandat)
---

## Vad det är

Valmyndighetens fastställda fördelning av riksdagens 349 mandat efter valet
den 13 september 2026. Beslutet är daterat 19 september 2026
(dnr VAL-735-2026). Riksdagen publicerar samma siffror.

Ingen ny regering var utsedd den 27 oktober 2026. Regeringen.se beskriver
en övergångsregering ledd av Ulf Kristersson sedan den 17 september 2026,
när den avgående statsministern begärde entledigande. Mandatsiffrorna är
kammarens sammansättning. De är inte ett påstående om vilket styre som
kommer att bildas.

## Hur du själv kommer åt datan

Hela rikets resultat, inklusive mandatfördelningen:

<https://resultat.val.se/val2026/>

Riksdagens sammanställning av samma beslut:

<https://www.riksdagen.se/sv/aktuellt/aktuelltnotiser/2026/sep/19/den-nya-riksdagen-efter-valet_cmsad780a37-f1ae-4d47-a4a7-b6ea2a69a21esv/>

Valmyndighetens beslutsprotokoll (PDF) finns länkat från resultatsidan
(«Beslut för riksdagsval 2026»).

## Schema/fält vi använder

- Mandat per parti — S 99, M 70, SD 62, V 30, C 25, KD 22, MP 22, L 19.
  Summan är 349.
- Övergångsregeringens startdatum 17 september 2026, från Regeringen.se.
  De partier som sitter i den avgående regeringen (M, KD, L) är markerade
  som regeringspartier i underlaget. Övriga partier ligger i kolumnen för
  opposition enbart för att kolumnen inte har något annat värde. Sidan visar
  dem som en mandatlista utan rubrikerna styre och opposition.

## Begränsningar och kända problem

- Regeringsbildningen pågick fortfarande den 27 oktober 2026. En utsedd
  regering kräver en ny rad. Den här raden ska inte skrivas om i efterhand.
- 2022 års mandat och Tidö-samarbetet ligger kvar som den stängda raden
  med slutdatum 17 september 2026. De visas inte som den aktuella kammaren.

## Hur vi bearbetar

Mandaten kopieras som Valmyndigheten publicerade dem. Inget parti slås ihop
och inget mandat räknas om. Övergångsregeringen visas som en lista över
mandat, inte som ett styre mot en opposition.
