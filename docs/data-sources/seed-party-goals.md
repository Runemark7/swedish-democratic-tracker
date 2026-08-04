---
id: seed-party-goals
name: Partimål (seed)
kind: seed
upstream: ~
license: Sammanställt från partiernas valmanifest 2022, valplattform 2026, Tidöavtalet och partiprogram
freshness: engångs (vid antagande); uppdateras manuellt vid nya manifest
last_verified: 2026-06-21
verification_status: frozen
verification_notes: ~
used_by:
  - /parties (partiscorecards)
  - /parties/:party/goals (mållista)
  - /parties/:party/goals/:goalId/votes (målets omröstningar)
---

## Vad det är
Hand-kuraterade mål från partiernas publicerade dokument — valmanifest 2022,
valplattform 2026, Tidöavtalet och partiprogram. Varje mål har en `source_url`
till källdokumentet och ett ordagrant `source_quote` ur det. På så sätt kan du
för varje mål gå tillbaka och verifiera exakt vad partiet skrev.

## Hur du själv kommer åt datan
Manifesten och valplattformarna publiceras av partierna själva. För varje
mål lagrar vi en `source_url` (länk till partiets publicerade dokument) och
ett `source_quote` (ordagrant citat ur dokumentet som målet bygger på), så
att du kan klicka dig vidare och läsa originalformuleringen.

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
- `goal_text` — målets formulering som vi har skrivit.
- `topic` — kategori (vård, skola, skatt, etc).
- `specificity` — vår klassificering: `concrete` (mätbart),
  `directional` (riktning utan siffra), `rhetorical` (slagord).
- `source_document` — källetikett (t.ex. "Valmanifest 2022").
- `source_url` — länk till partiets publicerade källdokument.
- `source_quote` — ordagrant citat ur källdokumentet som målet bygger på.
- `keywords` — sökord för att matcha relevanta riksdagsomröstningar.
- `relevant_committees` — utskott vars omröstningar är relevanta.

## Begränsningar och kända problem
- Manifest formuleras ofta vagt. Klassificeringen `concrete` vs
  `directional` är vår tolkning, inte partiets egen.
- Verifierbarheten backas numera av `source_url` + ordagrant `source_quote`
  per mål, så att varje mål går att kontrollera mot originaltexten.
- Täckningen för valet 2026 är ofullständig: i juni 2026 hade endast
  Socialdemokraterna publicerat en valplattform för 2026. Övriga mål
  bygger på 2022 års manifest, Tidöavtalet och partiprogram. Listan
  uppdateras manuellt allteftersom partierna publicerar nya dokument.
- Vissa mål spänner över flera teman; vi har valt det dominerande
  temat för enklare filtrering.

## Hur vi bearbetar
Manuell extrahering från PDF → införd i vår databas vid driftsättning.
Frontend hämtar målen via det publika API:et `/api/parties` och
`/api/parties/:party/goals` och visar dem på partisidor och i
manifestöversikten. Alignment mot ledamöternas faktiska röster
beräknas i [riksdagen](./riksdagen.md) och visas tillsammans med
varje mål.
