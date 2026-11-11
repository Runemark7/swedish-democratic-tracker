# Runbook: suggest Föreslagna kopplingar (the bot's weekly read and submit)

Issue #159, spec #129. This is the mechanism only. It creates no suggestions by
itself and nothing runs on a schedule; the bot calls it. Föreslagna kopplingar
and bortsorterade suggestions are **never public**: no public endpoint reads
`kopplingar`, and the public Ändringslogg omits every `koppling` entry.

All calls need `Authorization: Bearer <write token>`; writes need an
`Idempotency-Key`.

## 1. Read the candidates

```bash
curl -H "Authorization: Bearer $TOKEN" "$BASE/api/write/candidates?limit=50"
```

Returns decided voteringar that were ingested **after the last closed
Veckobatch** (all of them when none is closed yet), that lie inside the mandate
period of at least one approved Löfte, and whose utskott (the letters of the
beteckning, `SkU16` → `SkU`) is one of that Löfte's `utskott`. Each votering
comes with its betänkande, förslagspunkt, the motions/yrkanden decided there,
and only **approved** Löften. Pairs already suggested (any state) are omitted.
A Löfte with no `utskott` is never offered. Page with `after=<next>` until
`next` is `null`.

## 2. Submit suggestions

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Idempotency-Key: kopplingar-2026-11-15-1" \
  -H 'Content-Type: application/json' "$BASE/api/write/kopplingar" -d '{
  "threshold": 0.6,
  "chatRef": "bot-run-2026-11-15", "onBehalfOf": "redaktor",
  "suggestions": [
    {"loefteId": 81, "voteringId": "<votering_id>", "direction": "Nej",
     "confidence": 0.82, "reasoning": "…"}
  ]}'
```

- `direction` is the ballot (Ja or Nej) that is in the Löfte's direction. A
  vote to reject a motion supporting the Löfte is Nej.
- Confidence **at or above** `threshold` (default 0.6) → `foreslagen`; below →
  `bortsorterad`. Both are stored with confidence and reasoning; bortsorterade
  are logged only so the pair is not suggested again.
- The whole batch is rejected with **422** (nothing stored) if any item names a
  Löfte that is not approved, an unknown votering, or a votering outside the
  Löfte's mandate period (a 2026 Löfte only takes voteringar of 2026–2030).
- A pair that already has a koppling comes back as `existing` and is not changed.
- A retry with the same key and body returns the same answer
  (`Idempotent-Replayed: true`); same key with another body is 409.

## What comes next

Creating the numbered Veckobatch from open `foreslagen` rows and deciding it is
#160; it also closes batches, which is what moves the "since" cutoff of the
candidates read. Undo of a suggestion entry is not supported yet.
