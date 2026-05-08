---
id: seed-party-goals
name: Partimål (seed)
kind: seed
upstream: ~
license: Sammanställt från partiernas valmanifest 2022
freshness: engångs (vid migration); uppdateras manuellt vid nya manifest
last_verified: 2026-05-08
used_by:
  - /parties (partiscorecards)
  - /parties/:party/goals (mållista)
  - /parties/:party/goals/:goalId/votes (målets omröstningar)
  - /manifestos (manifestöversikt)
---

## Vad det är
Hand-kuraterade mål från partiernas valmanifest 2022. Varje rad i
`backend/migrations/000003_seed_party_goals.sql` har en citation
tillbaka till källtexten — vanligtvis en specifik sida i partiets
PDF-manifest.

## Hur du själv kommer åt datan
Källfilerna är PDF:er på partiernas hemsidor. Citatuppgifter finns i
migrationen som kommentarer. För att kontrollera ett mål:

```bash
grep -B1 -A5 "title.*VARMÅL" backend/migrations/000003_seed_party_goals.sql
```

URL:er till manifestens primärkälla:
- Socialdemokraterna: https://www.socialdemokraterna.se/var-politik
- Moderaterna: https://moderaterna.se/var-politik
- (osv för alla 8 riksdagspartier)

## Schema/fält vi använder
Tabellen `party_goals`:
- `party` — partikod.
- `topic` — Vård, Skola, Skatt, etc.
- `title` — målets korta titel.
- `description` — full målbeskrivning.
- `specificity` — concrete | directional | rhetorical.
- `source_citation` — citat med sidnummer.

## Begränsningar och kända problem
- Manifest formuleras ofta vagt; klassificering som "concrete" vs
  "rhetorical" är vår tolkning, inte partiets.
- Endast målen från **2022**-valet är seedade. Inför 2026-valet behöver
  motsvarande migration skapas.

## Hur vi bearbetar
- En SQL-migration (`000003`) inserter alla rader vid första
  `make migrate`-körning.
- Frontend läser via `useQuery(["parties", ...])` mot
  `/api/parties` och `/api/parties/:party/goals`.
