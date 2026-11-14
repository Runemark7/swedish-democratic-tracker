import assert from "node:assert/strict";
import test from "node:test";
import { FOOTER_NAV, SECONDARY_NAV, documentTitle } from "./nav.ts";

// Gaps #301/#302/#303: Metod and Ändringslogg must be reachable from the footer on every page.

test("the footer links Om sajten, Data, Metod and Ändringslogg", () => {
  assert.deepEqual(
    FOOTER_NAV.map((i) => [i.label, i.to]),
    [
      ["Om sajten", "/om-sajten"],
      ["Data", "/data"],
      ["Metod", "/metod"],
      ["Ändringslogg", "/andringslogg"],
    ],
  );
});

test("the header and mobile 'Mer' menu stay at Om sajten and Data", () => {
  assert.deepEqual(SECONDARY_NAV.map((i) => i.label), ["Om sajten", "Data"]);
});

test("Metod and Ändringslogg get their own tab titles", () => {
  assert.equal(documentTitle("/metod"), "Riksdagskollen – Metod");
  assert.equal(documentTitle("/andringslogg"), "Riksdagskollen – Ändringslogg");
});
