import { test } from "node:test";
import assert from "node:assert/strict";
import { isoDate } from "../sync-data-sources.mjs";

test("a YAML date (parsed as midnight UTC) becomes an ISO date", () => {
  // gray-matter turns `last_verified: 2026-11-11` into a Date. String() of
  // that Date is what used to leak into the UI as "Wed Nov 11 2026 00:00:00
  // GMT+0000 (Coordinated Universal Time)".
  assert.equal(isoDate("x.md", new Date("2026-11-11T00:00:00Z")), "2026-11-11");
  assert.equal(isoDate("x.md", new Date(Date.UTC(2026, 5, 8))), "2026-06-08");
});

test("a quoted ISO date is kept as written", () => {
  assert.equal(isoDate("x.md", "2026-06-08"), "2026-06-08");
});

test("anything else fails the sync instead of reaching the UI", () => {
  assert.throws(() => isoDate("x.md", "Mon Jun 08 2026"), /x\.md.*last_verified/);
  assert.throws(() => isoDate("x.md", new Date("nope")), /last_verified/);
  assert.throws(() => isoDate("x.md", 20261111), /last_verified/);
});
