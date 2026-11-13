import assert from "node:assert/strict";
import test from "node:test";
import { loftenHitCount } from "./hitCount.ts";
import { mandatperiodFromQuery } from "./labels.ts";

test("count without filters", () => {
  assert.equal(loftenHitCount(83), "83 löften");
  assert.equal(loftenHitCount(1), "1 löfte");
  assert.equal(loftenHitCount(0), "Inga löften");
});

test("count with a party filter names the party", () => {
  assert.equal(loftenHitCount(16, "S"), "16 löften för S");
  assert.equal(loftenHitCount(0, "S"), "Inga löften för S");
});

test("count with a mandatperiod names the period", () => {
  assert.equal(loftenHitCount(10, "", "2026-2030"), "10 löften, mandatperiod 2026–2030");
  assert.equal(loftenHitCount(6, "S", "2022-2026"), "6 löften för S, mandatperiod 2022–2026");
  assert.equal(loftenHitCount(0, "", "2026-2030"), "Inga löften, mandatperiod 2026–2030");
  assert.equal(loftenHitCount(0, "S", "2022-2026"), "Inga löften för S, mandatperiod 2022–2026");
});

test("an unknown period code is shown as written", () => {
  assert.equal(loftenHitCount(0, "", "2018-2022"), "Inga löften, mandatperiod 2018-2022");
});

test("only a well-formed ?mandatperiod= is sent to the API", () => {
  assert.equal(mandatperiodFromQuery("2026-2030"), "2026-2030");
  assert.equal(mandatperiodFromQuery(null), "");
  assert.equal(mandatperiodFromQuery("nu"), "");
  assert.equal(mandatperiodFromQuery("2026–2030"), "");
});
