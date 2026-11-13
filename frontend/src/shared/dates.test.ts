import assert from "node:assert/strict";
import test from "node:test";
import { swedishDate } from "./dates.ts";
import { SOURCES } from "../components/sources/SourceRegistry.generated.ts";

test("an ISO date reads as a Swedish date", () => {
  assert.equal(swedishDate("2026-06-08"), "8 juni 2026");
  assert.equal(swedishDate("2026-11-11"), "11 november 2026");
});

test("a known UTC timestamp reads as a Swedish date", () => {
  assert.equal(swedishDate("2026-11-11T00:00:00Z"), "11 november 2026");
  assert.equal(swedishDate("2026-06-08T00:00:00.000Z"), "8 juni 2026");
});

test("a timestamp is dated by the Swedish calendar day, not the UTC one", () => {
  // 23:30 UTC on 10 November is already 00:30 on 11 November in Stockholm.
  assert.equal(swedishDate("2026-11-10T23:30:00Z"), "11 november 2026");
});

test("a raw JavaScript Date string never reaches the reader in English", () => {
  const raw = "Mon Jun 08 2026 00:00:00 GMT+0000 (Coordinated Universal Time)";
  assert.equal(swedishDate(raw), "8 juni 2026");
  assert.equal(swedishDate(String(new Date("2026-11-10T00:00:00Z"))), "10 november 2026");
});

test("a value that is no date is shown unchanged", () => {
  assert.equal(swedishDate("okänt"), "okänt");
  assert.equal(swedishDate("2026-13-01"), "2026-13-01");
});

test("every source in the registry states when it was verified in Swedish", () => {
  const ids = Object.keys(SOURCES);
  assert.ok(ids.length > 0);
  for (const id of ids) {
    const { lastVerified } = SOURCES[id];
    assert.match(lastVerified, /^\d{4}-\d{2}-\d{2}$/, `${id}: lastVerified must be an ISO date`);
    const shown = swedishDate(lastVerified);
    assert.match(shown, /^\d{1,2} [a-zåäö]+ \d{4}$/, `${id}: ${shown}`);
    assert.doesNotMatch(shown, /GMT|Universal|Standard Time/i);
  }
});
