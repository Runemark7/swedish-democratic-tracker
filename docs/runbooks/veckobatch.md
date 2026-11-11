# Runbook: Veckobatch (the weekly bot routine and its chat format)

Issue #160, spec #129, ADR-003. This describes what the weekly bot routine
**does when someone runs it**. Nothing in the repository schedules it, and
nothing may answer for the Redaktör: every decision comes from the Redaktör's
own chat reply, relayed verbatim.

All calls need `Authorization: Bearer <write token>`; writes need an
`Idempotency-Key`. Föreslagna kopplingar, batches and their decisions are
**never public**: no public endpoint reads them and the public Ändringslogg
omits `koppling` and `veckobatch` entries.

## The Sunday routine

1. **Read candidates** and **submit suggestions** as in
   [suggest-kopplingar.md](suggest-kopplingar.md). The "since" cutoff is the
   *creation time* of the last closed Veckobatch.
2. **Create the batch**

   ```bash
   curl -X POST -H "Authorization: Bearer $TOKEN" -H "Idempotency-Key: veckobatch-2026-11-15" \
     -H 'Content-Type: application/json' "$BASE/api/write/veckobatch" \
     -d '{"chatRef":"bot-run-2026-11-15","onBehalfOf":"redaktor"}'
   ```

   - Picks open (`foreslagen`) kopplingar, at most 20 (`max` may lower it):
     items the Redaktör did not answer in earlier batches first (flag
     `rolledOver`), then strongest confidence first. They are numbered 1..N;
     **the numbers never change** (re-reading, retrying the same key, or new
     suggestions arriving later do not renumber). A rolled-over item gets a new
     number in the new batch.
   - Creating a batch **closes the previous one**; answers to a closed batch
     are 409. 422 when there is nothing open: post nothing then.
   - Retrying with the same key and body returns the same batch
     (`Idempotent-Replayed: true`).
3. **Post it in chat** (format below) and stop. Wait for the Redaktör.
4. **When the Redaktör answers**, submit their sentence **unchanged**:

   ```bash
   curl -X POST -H "Authorization: Bearer $TOKEN" -H "Idempotency-Key: veckobatch-2026-11-15-answer-1" \
     -H 'Content-Type: application/json' "$BASE/api/write/veckobatch/$ID/answers" \
     -d '{"text":"1,3,7 ja, resten nej","chatRef":"<link or id of the chat message>","onBehalfOf":"redaktor"}'
   ```

   The server parses the sentence strictly. `422` means it could not be read
   (or names a missing item): nothing was stored; **ask the Redaktör again**,
   never guess. Structured form: `{"answers":[{"nr":4,"answer":"ja-omvaent"}],"rest":"nej",…}`.

## Answer vocabulary

| Chat | Meaning | Result |
|---|---|---|
| `1,3,7 ja` / `2-4 ja` | confirm with the suggested direction | `bekraftad`, direction as suggested |
| `4 ja men omvänt` | confirm with the **direction reversed** (Ja↔Nej) | `bekraftad`, direction flipped |
| `5 nej` | reject | `avvisad` |
| `resten nej` / `övriga ja` | apply to every item not named | as above |
| (not mentioned, no "resten") | no answer | stays `foreslagen`, **rolls over** into next Sunday's batch |

- A Bekräftad koppling (`bekraftad`) with its direction is what Löftesstatus
  will later be derived from (#161). It is not shown publicly by this ticket.
- The whole answer is atomic: either every named item is decided or none.
  An item that is already answered, or a closed batch, is 409.
- When every item is answered the batch closes by itself.
- Retrying with the same key returns the same result; the same key with a
  different body is 409.

## Chat format (post exactly this shape)

```
Veckobatch #12 — 14 förslag (3 kvarstår från förra veckan)
Svara kort, t.ex. "1,3,7 ja, resten nej" eller "4 ja men omvänt".
Ej nämnda förslag stannar kvar till nästa vecka.

1. [S] "Sänkt skatt på arbete" — röstning SkU2 p. 3 (Skatt på kapital), 2026-10-20
   Löftet: <quote> (<locator>, <sourceUrl>)
   Föreslagen riktning: Nej-röst stödjer löftet · säkerhet 0.92 · kvarstår
   Skäl: <reasoning>
2. …
```

One block per item, in `nr` order, with the quote, the votering (beteckning,
förslagspunkt, betänkande title, date) and the bot's reasoning, all from
`GET /api/write/veckobatch/{id}` (or `/latest`). Mark `rolledOver` items.
Never recommend an answer, never answer, never pre-fill.

## Every decision is logged and undoable

Each decision is one Ändringslogg entry (`entity: koppling`, old and new
value, `chatRef`, `onBehalfOf`). Find the entries (`entryId` is also in the
answer response):

```bash
curl -H "Authorization: Bearer $TOKEN" … # entry ids come back in the answers response
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Idempotency-Key: undo-123" \
  -H 'Content-Type: application/json' "$BASE/api/write/aendringslogg/$ENTRY_ID/undo" \
  -d '{"chatRef":"<chat message asking for the undo>","onBehalfOf":"redaktor"}'
```

- Undo puts the koppling back to `foreslagen` with the suggested direction and
  clears the item's answer. It is open again and is offered in the next batch
  (as a rolled-over item). The undo is itself logged; undoing the undo
  re-applies the decision.
- If the koppling changed after that entry, the undo is 409 and nothing happens.
- Entries that merely *suggested* a koppling (#159) cannot be undone (422).
- The undo is only made when the Redaktör asks for it.

## Not done here

No cron, no scheduler and no real batch is created by this repository. Setting
up the weekly trigger is a deliberate, separate step by whoever operates the bot.
