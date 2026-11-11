# Runbook: curated motion links (fråga → motion, Återkommande krav)

Issue #163, spec #129, ADR-003. Phase 1 links only what is certain. Two kinds
of link need a human, so the bot **suggests** them and the Redaktör decides in
the weekly [Veckobatch](veckobatch.md):

| kind | what it links | example |
|---|---|---|
| `fraga` | a skriftlig fråga or interpellation to a motion it names **only in prose** | fråga 2022/23:339 ("en kommittémotion från Moderaterna redan 2017") → motion 2017/18:3293 |
| `krav` | two yrkanden of **different wording** that are the same Återkommande krav | Rickard Nordin's (C) 2022/23:1675 yrkande 1 ↔ M's 2021/22:3638 yrkande 9 |

The rule recorded with a link is `prose-citation` or `similar-wording`; its
state is `foreslagen` (suggested), `bekraftad` (confirmed), `avvisad`
(rejected) or `bortsorterad` (below the threshold, kept so it is not suggested
again).

## Visibility — the one rule

**Only `bekraftad` is public.** A suggested, rejected, bortsorterad or undone
link is in no public endpoint, and the public Ändringslogg omits
`motionkoppling` entries (and `veckobatch`). A confirmed link is shown with the
rule `redaktor-confirmed` ("Bekräftad av redaktören"):

- fråga: the motion page (`/api/motions/{rm}/{bet}/questions`) and the question
  page list each other;
- krav: the two Återkommande krav groups become one on every member's page.
  Undo removes the link and the groups separate again at once.

## Finding candidates

There is no special endpoint. Use the public API, then judge:

- fråga: `GET /api/questions/{kind}/{rm}/{bet}` has the full `text`; a question
  whose `motions` is empty and whose text names a motion ("kommittémotion",
  "motion 2017") is a candidate. Find the motion with `GET /api/motions`.
- krav: `GET /api/motions/{rm}/{bet}` lists each yrkande's `wording`; compare
  yrkanden from different parties/years that ask for the same thing.

Never suggest on a guess: put the reason (what in the text points at the
motion / why the wording means the same) in `reasoning`.

## Suggest

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Idempotency-Key: motion-links-2026-11-15" \
  -H 'Content-Type: application/json' "$BASE/api/write/motion-kopplingar" -d '{
  "chatRef":"bot-run-2026-11-15","onBehalfOf":"redaktor","threshold":0.6,
  "suggestions":[
    {"kind":"fraga","question":{"kind":"fraga","riksmote":"2022/23","beteckning":"339"},
     "motion":{"riksmote":"2017/18","beteckning":"3293"},
     "confidence":0.9,"reasoning":"Frågan citerar \"en kommittémotion från Moderaterna redan 2017\" (virtuella varor)."},
    {"kind":"krav","motion":{"riksmote":"2021/22","beteckning":"3638","yrkande":9},
     "other":{"riksmote":"2022/23","beteckning":"1675","yrkande":1},
     "confidence":0.85,"reasoning":"Båda kräver vinstmarginalbeskattning på digitala varor."}
  ]}'
```

- All-or-nothing: an unknown question, motion or yrkande is 422 and stores
  nothing. A krav pair is unordered (the same pair in the other order is the
  same link); a pair that exists is reported `existing` and left unchanged.
- Below `threshold` (default 0.6) the link is stored `bortsorterad`.
- Suggesting is logged in the Ändringslogg (never public) but cannot be undone;
  it changes nothing the public can see.

## Decide

`POST /api/write/veckobatch` puts open motion links in the same numbered batch
as Löfte kopplingar (at most 20 in all; strongest first, unanswered ones roll
over). Items have `kind: "fraga"|"krav"` and a `motion` block: for a fråga the
question (title, asker, excerpt, source link) and the motion; for a krav both
yrkanden with their wording, parties and source links.

Chat shape for these items:

```
3. [fråga → motion] Skriftlig fråga 2022/23:339 (Rickard Nordin, C) nämner motion 2017/18:3293 (M)
   Frågan: "<excerpt>"  <sourceUrl>
   Motionen: <title>  <sourceUrl>
   Säkerhet 0.90 · Skäl: <reasoning>
4. [samma krav?] 2021/22:3638 yrkande 9 (M): "<wording>"  <sourceUrl>
                2022/23:1675 yrkande 1 (C): "<wording>"  <sourceUrl>
   Säkerhet 0.85 · Skäl: <reasoning>
```

Answer with the Redaktör's own sentence, unchanged (`"3 ja, 4 nej"`). `ja`
confirms (public at once), `nej` rejects. `ja men omvänt` is refused (422).

## Undo

Each decision is an Ändringslogg entry (`entity: motionkoppling`, id in the
answer response). `POST /api/write/aendringslogg/{id}/undo` puts the link back
to `foreslagen`: it disappears from the public pages immediately and is offered
again in the next batch. Only when the Redaktör asks.

## Not done here

No link is created, suggested or confirmed by this repository, and nothing is
scheduled. The seam tests (`backend/internal/apitest/curated_links_test.go`)
use recorded fixtures only: fråga 2022/23:339 with motion 2017/18:3293
(`fixtures/questions`), and M's and Nordin's yrkanden (`fixtures/recurring-krav`).
