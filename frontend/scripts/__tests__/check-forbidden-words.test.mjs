import { test } from "node:test";
import assert from "node:assert/strict";
import { findForbidden } from "../check-forbidden-words.mjs";

test("flags hållet and brutet as whole words, any case", () => {
  assert.equal(findForbidden("Löftet är hållet").length, 1);
  assert.equal(findForbidden("Brutet löfte").length, 1);
  assert.equal(findForbidden("a\nb brutet.").at(0).line, 2);
});

test("allows words that merely contain them", () => {
  assert.equal(findForbidden("innehållet i betänkandet, avbrutet möte").length, 0);
});

test("the shipped UI and data-source pages are clean", async () => {
  const { main } = await import("../check-forbidden-words.mjs");
  const exit = process.exit;
  let code = 0;
  process.exit = (c) => { code = c; };
  try { await main(); } finally { process.exit = exit; }
  assert.equal(code, 0);
});
