---
id: seed-party-goals
name: Partimål (seed)
kind: seed
upstream: ~
license: Sammanställt från partiernas valmanifest 2022
freshness: engångs (vid antagande); uppdateras manuellt vid nya manifest
last_verified: 2026-05-08
used_by:
  - /parties (partiscorecards)
  - /parties/:party/goals (mållista)
  - /parties/:party/goals/:goalId/votes (målets omröstningar)
  - /manifestos (manifestöversikt)
---

## Vad det är
Hand-kuraterade mål från partiernas valmanifest 2022. Varje mål har en
citation tillbaka till källtexten — vanligtvis en specifik sida i
partiets PDF-manifest. På så sätt kan du för varje mål gå tillbaka och
verifiera ordagrant vad partiet skrev i sitt manifest.

## Hur du själv kommer åt datan
Manifesten är PDF:er på partiernas egna hemsidor. För varje mål visar
vi den sida i manifestet som målet är hämtat från, så du kan slå upp
och läsa originalformuleringen.

Primärkällor:
- Socialdemokraterna: <https://www.socialdemokraterna.se/var-politik>
- Moderaterna: <https://moderaterna.se/var-politik>
- Sverigedemokraterna: <https://sd.se/var-politik/>
- Centerpartiet: <https://www.centerpartiet.se/var-politik>
- Vänsterpartiet: <https://www.vansterpartiet.se/politik/>
- Kristdemokraterna: <https://kristdemokraterna.se/politik/>
- Liberalerna: <https://www.liberalerna.se/politik/>
- Miljöpartiet: <https://www.mp.se/politik>

## Schema/fält vi använder
- `party` — partikod (S, M, SD, C, V, KD, L, MP).
- `topic` — kategori (Vård, Skola, Skatt, etc).
- `title` — målets korta rubrik.
- `description` — full målbeskrivning som vi har formulerat.
- `specificity` — vår klassificering: `concrete` (mätbart),
  `directional` (riktning utan siffra), `rhetorical` (slagord).
- `source_citation` — citat ur manifestet med sidnummer.

## Begränsningar och kända problem
- Manifest formuleras ofta vagt. Klassificeringen `concrete` vs
  `rhetorical` är vår tolkning, inte partiets egen.
- Endast 2022-valets mål är inlagda. Inför nästa val behöver listan
  uppdateras manuellt med nya manifest.
- Vissa mål spänner över flera teman; vi har valt det dominerande
  temat för enklare filtrering.

## Hur vi bearbetar
Manuell extrahering från PDF → införd i vår databas vid driftsättning.
Frontend hämtar målen via det publika API:et `/api/parties` och
`/api/parties/:party/goals` och visar dem på partisidor och i
manifestöversikten. Alignment mot ledamöternas faktiska röster
beräknas i [riksdagen](./riksdagen.md) och visas tillsammans med
varje mål.
