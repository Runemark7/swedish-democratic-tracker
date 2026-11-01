import assert from "node:assert/strict";
import test from "node:test";
import { budgetAreaPath, regeringenRedirect, regeringenSection } from "./regeringen.ts";

test("budget overview lands on the budget section", () => {
  assert.deepEqual(regeringenRedirect("/budget"), regeringenSection("budget"));
});

test("a budget area page becomes a Regeringen sub-page and keeps the code", () => {
  assert.equal(regeringenRedirect("/budget/areas/1"), "/regeringen/budget/1");
  assert.equal(regeringenRedirect("/budget/areas/UO6"), "/regeringen/budget/UO6");
  assert.equal(budgetAreaPath("27"), "/regeringen/budget/27");
});

test("the Riksdag overview lands on the Riksdag section", () => {
  assert.deepEqual(regeringenRedirect("/riksdag"), regeringenSection("riksdag"));
});

test("Regering lands on Regeringen", () => {
  assert.equal(regeringenRedirect("/regering"), "/regeringen");
});

test("minister pages, the agency register and its detail pages are not redirected", () => {
  assert.equal(regeringenRedirect("/regering/abc"), null);
  assert.equal(regeringenRedirect("/riksdag/myndigheter"), null);
  assert.equal(regeringenRedirect("/riksdag/myndigheter/arbetsformedlingen"), null);
  assert.equal(regeringenRedirect("/regeringen"), null);
  assert.equal(regeringenRedirect("/regeringen/budget/1"), null);
});
