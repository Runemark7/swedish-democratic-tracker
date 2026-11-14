import assert from "node:assert/strict";
import test from "node:test";
import { punktOutcome, type OutcomeInput } from "./punktOutcome.ts";

const party = (p: string, partilinje: string, ja: number, nej: number) => ({ party: p, partilinje, ja, nej, avstar: 0, franvarande: 0 });

const voted: OutcomeInput = {
  mode: "votering",
  vinnare: "utskottet",
  voteringStored: true,
  parties: [party("M", "Ja", 66, 0), party("S", "Ja", 105, 0), party("V", "Nej", 0, 21), party("C", "Nej", 0, 23), party("-", "", 5, 4)],
};

test("utskottet won a votering: says so, with the tally and the parties behind it", () => {
  assert.deepEqual(punktOutcome(voted, true), {
    label: "Riksdagen biföll utskottets förslag",
    tally: "176 ja, 48 nej",
    winningParties: ["M", "S"],
    losingParties: ["V", "C"],
  });
});

test("a reservation won: the utskott's proposal is said to have lost", () => {
  const o = punktOutcome({ ...voted, vinnare: "reservation 1", parties: [party("S", "Ja", 100, 0), party("M", "Nej", 0, 70), party("SD", "Nej", 0, 62)] }, true);
  assert.equal(o?.label, "Riksdagen biföll reservation 1 – utskottets förslag avslogs");
  assert.equal(o?.tally, "100 ja, 132 nej");
  assert.deepEqual(o?.winningParties, ["M", "SD"]);
  assert.deepEqual(o?.losingParties, ["S"]);
});

test("decided without a vote: outcome without tally or parties", () => {
  assert.deepEqual(punktOutcome({ mode: "acklamation", vinnare: "utskottet", voteringStored: false, parties: [] }, true), {
    label: "Riksdagen biföll utskottets förslag", tally: "", winningParties: [], losingParties: [],
  });
});

test("the winner's casing and spacing do not matter", () => {
  assert.equal(punktOutcome({ ...voted, vinnare: " Utskottet " }, true)?.label, "Riksdagen biföll utskottets förslag");
  assert.equal(punktOutcome({ ...voted, vinnare: "Reservation  12" }, true)?.label, "Riksdagen biföll reservation 12 – utskottets förslag avslogs");
});

test("a winner worded another way is shown as Riksdagen words it, without naming sides", () => {
  assert.deepEqual(punktOutcome({ ...voted, vinnare: "återförvisning" }, true), {
    label: "Vinnande förslag enligt Riksdagen: återförvisning", tally: "176 ja, 48 nej", winningParties: [], losingParties: [],
  });
});

test("no winner from Riksdagen but a stored votering: only what the tally shows", () => {
  assert.deepEqual(punktOutcome({ ...voted, vinnare: "" }, true), {
    label: "Majoriteten röstade ja", tally: "176 ja, 48 nej", winningParties: ["M", "S"], losingParties: ["V", "C"],
  });
  assert.equal(punktOutcome({ ...voted, vinnare: "", parties: [party("S", "Ja", 10, 0), party("M", "Nej", 0, 20)] }, true)?.label, "Majoriteten röstade nej");
});

test("unknown outcome shows nothing", () => {
  // No winner and no votering to count.
  assert.equal(punktOutcome({ mode: "votering", vinnare: "", voteringStored: false, parties: [] }, true), null);
  assert.equal(punktOutcome({ mode: "annat", vinnare: "", voteringStored: false, parties: [] }, true), null);
  // A tie is not an outcome we can read from the tally.
  assert.equal(punktOutcome({ ...voted, vinnare: "", parties: [party("S", "Ja", 10, 0), party("M", "Nej", 0, 10)] }, true), null);
});

test("a betänkande that is not decided has no outcome, whatever the punkt says", () => {
  assert.equal(punktOutcome(voted, false), null);
  assert.equal(punktOutcome({ ...voted, mode: "ej_beslutad" }, true), null);
});

test("decision state unknown (Riksdagen unreachable): the stored winner is still shown", () => {
  assert.equal(punktOutcome(voted, null)?.label, "Riksdagen biföll utskottets förslag");
});
