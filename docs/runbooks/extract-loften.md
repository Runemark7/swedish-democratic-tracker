# Runbook: extract Löften from a Valmanifest, one party at a time (issue #158)

For the bot (the holder of `WRITE_TOKEN`) working with the Redaktör in chat.
The bot **proposes**; only the Redaktör's answer in chat **approves**. A
proposed Löfte is never public: not in `/api/loften`, not on the party page,
and not in the public `/api/aendringslogg`.

Nothing in this flow runs on deploy, at startup or on a schedule. It is
conversation-driven, and every call is idempotent.

## The testability rule

A Löfte is included only if a **votering can test it**: a concrete statement
that the Riksdag could vote for or against. Ask of each candidate:

1. **Does it name a decision?** "Sänk skatten på arbete", "inför en ny
   straffskala", "höj bidraget till X" – yes. "Ett tryggare Sverige", "vi
   värnar friheten" – no (rhetorical).
2. **Is it for the Riksdag to decide?** State, law or budget. Not what a
   kommun does, not what the party "will work for", not a value statement.
3. **Is the wording in the manifest?** The `quote` is the party's **exact**
   words, copied, not paraphrased. `text` may be a shorter statement of the
   claim, but stays close to the quote and adds nothing the quote does not say.
4. **Is it one claim?** Split a sentence with two demands into two Löften.

Mark `specificity` honestly: `concrete` (a measurable decision), `directional`
(a direction a vote could still test), `rhetorical` (do not propose these; if
the Redaktör wants one, mark it so). Take 15–25 per party. Fewer is fine if
the manifest has fewer testable demands: never pad.

Neutrality: pick by the rule above, in the manifest's own order and emphasis,
not by what is politically convenient. Do not label a Löfte kept or broken.
Löftesstatus comes from votes later (ADR-002), never from you.

## Per-party flow

1. **Pick the Valmanifest.** `GET /api/parties/{party}/valmanifest` and take
   the 2026 entry's `id` and `url`/`pdfUrl`. If it is missing or wrong, write it
   with `POST /api/write/valmanifest` first. Read the manifest itself, not a
   summary of it.
2. **Extract** candidates against the rule above. For each: `text`, `quote`,
   `locator` (page number or section) and `sourceUrl`. The `sourceUrl` must
   lead to the original, with the page or anchor when possible
   (`…/valplattform.pdf#page=12`).
3. **Propose** each one:

   ```bash
   curl -sS -X POST "$BASE/api/write/loften" \
     -H "Authorization: Bearer $WRITE_TOKEN" \
     -H "Idempotency-Key: propose-S-2026-001" \
     -H "Content-Type: application/json" \
     -d '{"party":"S","valmanifestId":123,"text":"…","quote":"…","locator":"s. 12",
          "sourceUrl":"https://…#page=12","topic":"skatt","specificity":"concrete",
          "utskott":["SkU"],"chatRef":"<chat answer reference>","onBehalfOf":"<Redaktör>"}'
   ```

   A missing quote or `sourceUrl`, a Valmanifest of another party, or an
   unknown party is **422** and stores nothing. The same quote proposed twice is
   **409** (it is allowed again after the first was struck). A new, stable
   Idempotency-Key per Löfte (`propose-<party>-<year>-<n>`) makes a retry safe.
4. **Present the party's batch** to the Redaktör in chat, numbered, each with
   its quote, locator and link. One party at a time; do not start the next
   party before this one is decided.
5. **Write the decisions** the Redaktör gives, each with the chat answer's
   reference as `chatRef` (that reference is what the public Ändringslogg shows):
   - approve: `POST /api/write/loften/{id}/approve` body `{"chatRef","onBehalfOf"}`
   - strike: `POST /api/write/loften/{id}/strike` same body
   - edit then approve: `POST /api/write/loften/{id}` with the full
     `text, quote, locator, sourceUrl, topic, specificity, utskott` plus
     `chatRef` and `onBehalfOf`, then approve. An edit keeps the Löfte's
     visibility; quote and `sourceUrl` stay mandatory.
   Never approve what the Redaktör did not answer yes to.
6. **Check.** `GET /api/parties/{party}/loften` shows exactly the approved
   ones; `GET /api/aendringslogg` shows each approval with its `chatRef`.
7. **Mistakes:** `POST /api/write/aendringslogg/{entryId}/undo` reverses a
   change if nothing touched the Löfte since (409 otherwise). An undone approval
   returns the Löfte to proposed. Undone entries remain in the log.

## What the public sees

- Approved Löften: text, quote, locator, link to the original and the
  Valmanifest, on the party page (`?tab=loften`) and `/loften/{id}`.
- Public Ändringslogg: approvals and strikes of approved Löften. Changes where
  neither side was public (propose, edit while proposed, strike of a proposed
  one) are omitted; a value that was not public is withheld and the entry has
  `"redacted": true`.
