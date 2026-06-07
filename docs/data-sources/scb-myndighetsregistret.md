---
id: scb-myndighetsregistret
name: SCB Myndighetsregistret
kind: api
upstream: https://myndighetsregistret.scb.se
license: SCB öppna data — fri användning
freshness: löpande (uppdateras när myndigheter inrättas, avvecklas eller byter namn)
last_verified: 2026-06-06
verification_status: verified
verification_notes: ~
used_by:
  - /riksdag/myndigheter (lista över ~449 statliga myndigheter)
  - /riksdag/myndigheter/:slug (myndighetsprofil — namn, org-nummer, typ, webbadress)
---

## Vad det är

SCB:s Myndighetsregister är den officiella förteckningen över alla statliga
myndigheter i Sverige. Det förs av Statistiska centralbyrån på uppdrag av
regeringen och innehåller namn, organisationsnummer, typ (förvaltningsmyndighet,
domstol, utlandsmyndighet m.m.), huvudman och webbadress för varje myndighet.
Registret är källan för hela myndighetslistans struktur på sidan — utan det
vet vi inte ens vilka myndigheter som existerar.

## Hur du själv kommer åt datan

Registrets webbgränssnitt finns på <https://myndighetsregistret.scb.se> där
du kan söka och filtrera på typ, departementstillhörighet och status. Datan
är uppdelad i ett antal kategorier — förvaltningsmyndigheter under regeringen,
domstolar, AP-fonder, utlandsmyndigheter m.fl.

Varje kategori kan hämtas som en HTML-tabell via registrets egna sidor.
Registret kategoriserar myndigheterna i grupper som "Centrala myndigheter",
"Länsstyrelser", "Domstolar" och "Utlandsmyndigheter". Den fullständiga
täckningen kräver att samtliga grupper hämtas separat.

## Schema/fält vi använder

- Myndighetsnamn — officiellt namn.
- Organisationsnummer — 10-siffrig identifierare (format: XXXXXX-XXXX).
- Typ — myndighetskategori (t.ex. Förvaltningsmyndighet, Domstol).
- Webbadress — myndighetens officiella webbplats.
- Överordnat organ — anger om myndigheten lyder under regeringen, riksdagen
  eller annan principal.

## Begränsningar och kända problem

- Registret innehåller utlandsmyndigheter (ambassader, konsulat) vars
  org-nummer och typ skiljer sig från de inhemska. Vi importerar dem men
  visar dem separat i filtret "under riksdagen / utland".
- Myndigheter som nyligen inrättats eller avvecklats kan ha en viss
  fördröjning i registret.
- Webbadresser är inte alltid kompletta URL:er — en del är angivna utan
  protokoll-prefix.

## Hur vi bearbetar

En ingestionsjobb hämtar hela registret en gång per dygn och uppdaterar
vår databas. Vi normaliserar org-numren till ett kanoniskt format
(XXXXXX-XXXX) och matchar sedan mot annan myndighetsinformation (utgifter,
personaltal) via org-numret som nyckel. Listan på
<https://riksdagskollen.se/riksdag/myndigheter> hämtar resultatet direkt
från vår databas.
