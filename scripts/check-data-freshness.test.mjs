import test from "node:test";
import assert from "node:assert/strict";
import { currentRiksmote, daysBetween, evaluateFreshness } from "./check-data-freshness.mjs";

test("current riksmöte rolls in September", () => {
  assert.equal(currentRiksmote(new Date("2026-08-13T00:00:00Z")), "2025/26");
  assert.equal(currentRiksmote(new Date("2026-09-01T00:00:00Z")), "2026/27");
  assert.equal(currentRiksmote(new Date("2026-10-27T12:00:00Z")), "2026/27");
  assert.equal(currentRiksmote(new Date("2030-10-15T00:00:00Z")), "2030/31");
});

test("daysBetween counts calendar days", () => {
  assert.equal(daysBetween("2026-08-13", "2026-10-21"), 69);
  assert.equal(daysBetween("2026-10-21", "2026-10-21"), 0);
  // The freshness bug reported 2026-10-20 for a speech stored as 2026-10-21.
  // The lag is that one calendar day, not a timezone-adjusted instant.
  assert.equal(daysBetween("2026-10-20", "2026-10-21"), 1);
});

test("speech dates are compared as calendar days", () => {
  // One day under the limit stays quiet. The comparison is still the calendar
  // gap: 2026-10-20 against 2026-10-21 is exactly one day, the shift a
  // Stockholm midnight cast produced.
  assert.equal(daysBetween("2026-10-20", "2026-10-21"), 1);
  const shifted = evaluateFreshness({
    now,
    freshness: report({ speeches: { latestDate: "2026-10-20" } }),
    upstreamSpeechDate: "2026-10-21",
    upstreamVoteDate: "2026-10-15",
  });
  assert.equal(shifted.ok, true, shifted.reasons.join("; "));
  assert.equal(
    shifted.reasons.some((reason) => reason.startsWith("speeches:")),
    false,
  );

  const matched = evaluateFreshness({
    now,
    freshness: report({ speeches: { latestDate: "2026-10-21" } }),
    upstreamSpeechDate: "2026-10-21",
    upstreamVoteDate: "2026-10-15",
  });
  assert.equal(matched.ok, true, matched.reasons.join("; "));
});

const now = new Date("2026-10-27T06:00:00Z");

function report(overrides = {}) {
  return {
    speeches: { latestDate: "2026-10-21" },
    votes: { latestDate: "2026-10-15" },
    jobs: [
      {
        name: "fetch-speeches",
        schedule: "daily",
        lastSuccessAt: "2026-10-27T00:10:00Z",
      },
      {
        name: "refresh-scorecards",
        schedule: "weekly",
        lastSuccessAt: "2026-10-25T00:10:00Z",
      },
    ],
    ...overrides,
  };
}

test("quiet when we are within a week of Riksdagen", () => {
  const result = evaluateFreshness({
    now,
    freshness: report(),
    upstreamSpeechDate: "2026-10-21",
    upstreamVoteDate: "2026-10-15",
  });
  assert.equal(result.ok, true, result.reasons.join("; "));
});

test("speech freeze is the gap against Riksdagen, not the age of the date", () => {
  const result = evaluateFreshness({
    now,
    freshness: report({ speeches: { latestDate: "2026-08-13" } }),
    upstreamSpeechDate: "2026-10-21",
    upstreamVoteDate: "2026-10-15",
  });
  assert.equal(result.ok, false);
  assert.match(result.reasons.join("\n"), /speeches: ours 2026-08-13 is 69 days behind/);
});

test("a recess where both sides are old is not a failure", () => {
  const result = evaluateFreshness({
    now,
    freshness: report({
      speeches: { latestDate: "2026-06-15" },
      votes: { latestDate: "2026-06-15" },
    }),
    upstreamSpeechDate: "2026-06-15",
    upstreamVoteDate: "2026-06-15",
  });
  assert.equal(result.ok, true, result.reasons.join("; "));
});

test("a daily job that has not succeeded in two days fails", () => {
  const result = evaluateFreshness({
    now,
    freshness: report({
      jobs: [
        {
          name: "fetch-speeches",
          schedule: "daily",
          lastSuccessAt: "2026-10-24T00:00:00Z",
        },
      ],
    }),
    upstreamSpeechDate: "2026-10-21",
    upstreamVoteDate: "2026-10-15",
  });
  assert.equal(result.ok, false);
  assert.match(result.reasons.join("\n"), /fetch-speeches: last success/);
});

test("a job that has never succeeded fails", () => {
  const result = evaluateFreshness({
    now,
    freshness: report({
      jobs: [{ name: "fetch-votes", schedule: "daily", lastSuccessAt: null }],
    }),
    upstreamSpeechDate: "2026-10-21",
    upstreamVoteDate: "2026-10-15",
  });
  assert.equal(result.ok, false);
  assert.match(result.reasons.join("\n"), /fetch-votes: has never succeeded/);
});

test("upstream outage is a failure rather than a pass", () => {
  const result = evaluateFreshness({
    now,
    freshness: report(),
    upstreamSpeechDate: null,
    upstreamVoteDate: null,
    upstreamSpeechError: "timeout",
    upstreamVoteError: null,
  });
  assert.equal(result.ok, false);
  assert.match(result.reasons.join("\n"), /timeout/);
});
