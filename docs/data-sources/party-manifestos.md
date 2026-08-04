---
id: party-manifestos
name: Partiernas valmanifest (katalog)
kind: seed
upstream: ~
license: Partiernas egna publicerade dokument; länkas, återges inte
freshness: manuell — uppdateras när ett parti publicerar ett nytt dokument
last_verified: 2026-08-04
verification_status: verified
verification_notes: ~
used_by:
  - /manifestos (katalogen över valmanifest och valplattformar)
  - /parties/:party (länk till partiets källdokument)
---

## Vad det är
En katalog över de valmanifest och valplattformar partierna själva har
publicerat, från 2014 till 2026. Katalogen innehåller inga citat och ingen
sammanfattning av politiken — den pekar på partiets eget dokument så att du
kan läsa originaltexten.

Katalogen är källskiktet för partimålen: varje mål i
[seed-party-goals](./seed-party-goals.md) anger ett `source_document`, och det
dokumentet finns i den här katalogen.

## Hur du själv kommer åt datan
Varje dokument publiceras av partiet självt. Katalogen lagrar en `url` till
partiets publicering och, där en separat PDF finns, en `pdfUrl`.

För val före 2026 finns dokumenten även samlade hos Svensk nationell
datatjänst (SND), i deras arkiv "Vi vill":

<https://snd.gu.se/sv/vivill>

Primärkällor för 2026, kontrollerade 2026-08-04:

- Socialdemokraterna — valplattform, publicerad 2026-02-05:
  <https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026>
- Centerpartiet — valmanifest "Sverige kan mer":
  <https://val2026.centerpartiet.se/>
- Liberalerna — valmanifest "För din frihet", presenterat 2026-06-02:
  <https://www.liberalerna.se/pressmeddelanden/liberalerna-presenterar-valmanifest-infor-valet-2026-for-din-frihet>
- Sverigedemokraterna — "Valplattform 2026" (PDF):
  <https://www.sd.se/wp-content/uploads/2026/07/valplattform-2026.pdf>
- Vänsterpartiet — **preliminär** valplattform, beslutad på kongressen i
  april 2026:
  <https://www.vansterpartiet.se/wp-content/uploads/2026/04/Preliminar-Valplattform-efter-beslut-pa-kongressen-2026.pdf>
- Miljöpartiet — valmanifest i **utkast**:
  <https://www.mp.se/valmanifest-2026-utkast/>

Moderaterna och Kristdemokraterna hade inget publicerat valmanifest för 2026
i katalogen vid kontrollen 2026-08-04. Moderaterna publicerar "Vallöften
2026" på <https://moderaterna.se/valmanifest> men inget manifestdokument;
Kristdemokraterna har presenterat fyra grundkrav ("Sverigedrömmen") men
inget manifest vi har kunnat hitta.

## Schema/fält vi använder
- `party` — partikod (S, M, SD, C, V, KD, L, MP).
- `year` — valår dokumentet gäller.
- `title` — dokumentets titel som partiet skrivit den.
- `type` — `valmanifest`, `valplattform`, `partiprogram` eller
  `politiska-riktlinjer`, enligt vad partiet kallar dokumentet.
- `status` — `published` (slutligt dokument) eller `draft` (partiet har
  publicerat ett utkast eller en preliminär version). Saknas dokumentet helt
  finns ingen post, och sidan skriver då att **vår** katalog är tom för det
  partiet och året.
- `url` — länk till partiets publicering.
- `pdfUrl` — separat PDF där en sådan finns.

## Begränsningar och kända problem
- **Katalogen är vår, inte partiernas.** Att en post saknas betyder att vi
  inte har registrerat något dokument — inte att partiet inte har publicerat
  något. Sidan formulerar frånvaron så, och tillskriver den aldrig partiet.
- **Vi säger inget om framtiden.** Katalogen listar bara dokument som finns.
  Tidigare stod "förväntas presenteras inför valet september 2026" för sju
  partier; det var en förutsägelse om partiernas beteende, inte ett faktum,
  och den var felaktig för Centerpartiet från omkring 2026-07-01. Sådana
  poster finns inte längre.
- **`draft` är partiets egen märkning, inte vår bedömning.** Vänsterpartiets
  dokument heter "Preliminär valplattform" och Miljöpartiets ligger på en
  URL märkt `utkast`. Vi återger den märkningen och rangordnar inte dokument
  efter den — läsaren avgör vad ett utkast är värt.
- **Uppdatering kräver driftsättning.** Katalogen ligger i frontendkoden, så
  varje nytt manifest kräver en ny version av sajten. Under valrörelsens
  sista veckor är det den mest sannolika orsaken till att katalogen ligger
  efter.
- Listan för 2014 och 2018 saknar sammanfattningar för flera partier.

## Hur vi bearbetar
Ingen bearbetning: dokumenten läses inte in, sammanfattas inte och tolkas
inte. Katalogen består av titel, typ, år och länk, kontrollerade mot partiets
egen publicering. Sammanfattningarna som visas för vissa år är korta
ämnesangivelser, inte citat.
