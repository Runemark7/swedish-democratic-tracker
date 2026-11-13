import assert from "node:assert/strict";
import test from "node:test";
import { loftenHitCount } from "./hitCount.ts";

test("count without filters", () => {
  assert.equal(loftenHitCount(83), "83 löften");
  assert.equal(loftenHitCount(1), "1 löfte");
  assert.equal(loftenHitCount(0), "Inga löften");
});

test("count with a party filter names the party", () => {
  assert.equal(loftenHitCount(16, "S"), "16 löften för S");
  assert.equal(loftenHitCount(0, "S"), "Inga löften för S");
});
