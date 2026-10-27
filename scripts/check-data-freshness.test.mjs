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
