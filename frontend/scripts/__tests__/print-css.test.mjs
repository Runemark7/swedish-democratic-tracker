import { test } from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";

// Print uses the light palette whatever theme the reader has (issue #306):
// the dark theme's near-white text is unreadable on paper. CSS cannot share
// one declaration block between a selector and a media query, so the tokens
// are listed twice in index.css and this test keeps the two lists together.

const css = fs.readFileSync(path.resolve(import.meta.dirname, "../../src/index.css"), "utf8");

/** Custom properties declared in the block that starts at `start`. */
function tokens(start) {
  const from = css.indexOf(start);
  assert.notEqual(from, -1, `${start} not found in index.css`);
  const open = css.indexOf("{", from + start.length - 1);
  const block = css.slice(open + 1, css.indexOf("}", open));
  return new Map([...block.matchAll(/(--[\w-]+):\s*([^;]+);/g)].map((m) => [m[1], m[2].trim()]));
}

const light = tokens('[data-theme="light"] {');
const printStart = css.indexOf("@media print {");

test("index.css has a print stylesheet", () => {
  assert.notEqual(printStart, -1);
  assert.equal(css.indexOf("@media print {", printStart + 1), -1, "expected a single @media print block");
});

test("print redefines every light-theme token", () => {
  const print = tokens("@media print {\n  html:root {");
  assert.ok(light.size > 50, "light theme block not parsed");
  assert.deepEqual([...light.keys()].filter((k) => !print.has(k)), []);
  assert.deepEqual([...print.keys()].filter((k) => !light.has(k)), []);
  for (const [name, value] of light) {
    if (name === "--color-bg") continue; // paper is white, not the warm page tone
    assert.equal(print.get(name), value, name);
  }
  assert.equal(print.get("--color-bg"), "#ffffff");
});

test("print hides chrome and draws party badges as framed black text", () => {
  const print = css.slice(printStart);
  assert.match(print, /\.no-print\s*\{\s*display:\s*none\s*!important;/);
  assert.match(print, /\.party-badge\s*\{[^}]*background:\s*transparent\s*!important;[^}]*color:\s*#000000\s*!important;[^}]*border:\s*1px solid #000000;/);
  assert.equal(css.slice(0, printStart).includes(".party-badge"), false, "badge rules must not apply on screen");
  assert.equal(css.slice(0, printStart).includes(".no-print"), false, "no-print must not hide anything on screen");
});
