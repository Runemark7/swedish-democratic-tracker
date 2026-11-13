import assert from "node:assert/strict";
import test from "node:test";
import { ApiError } from "./api-client.ts";
import { isNotFound, retryDelayMs, shouldRetry } from "./api-errors.ts";

// Gap rate-limit-shown-as-not-found: a 429 from our own API was shown as
// "finns inte i vårt register".

test("a 404 means the record is not there", () => {
  assert.equal(isNotFound(new ApiError(404, "not found")), true);
});

test("a 400 for a malformed id means the record is not there", () => {
  assert.equal(isNotFound(new ApiError(400, "bad id")), true);
});

test("a 429 is not a missing record", () => {
  assert.equal(isNotFound(new ApiError(429, "Too Many Requests")), false);
});

test("5xx and network failures are not a missing record", () => {
  assert.equal(isNotFound(new ApiError(500, "boom")), false);
  assert.equal(isNotFound(new ApiError(503, "unavailable")), false);
  assert.equal(isNotFound(new TypeError("Failed to fetch")), false);
  assert.equal(isNotFound(null), false);
});

test("a 429 is retried twice, then given up on", () => {
  const e = new ApiError(429, "Too Many Requests");
  assert.equal(shouldRetry(0, e), true);
  assert.equal(shouldRetry(1, e), true);
  assert.equal(shouldRetry(2, e), false);
});

test("5xx and network failures are retried", () => {
  assert.equal(shouldRetry(0, new ApiError(503, "unavailable")), true);
  assert.equal(shouldRetry(0, new TypeError("Failed to fetch")), true);
});

test("a 404 is never retried", () => {
  assert.equal(shouldRetry(0, new ApiError(404, "not found")), false);
  assert.equal(shouldRetry(0, new ApiError(400, "bad id")), false);
});

test("the retry waits as long as Retry-After says, within a ceiling", () => {
  assert.equal(retryDelayMs(0, new ApiError(429, "Too Many Requests", 2)), 2000);
  assert.equal(retryDelayMs(0, new ApiError(429, "Too Many Requests", 60)), 5000);
});

test("without Retry-After the retry backs off from one second", () => {
  assert.equal(retryDelayMs(0, new ApiError(429, "Too Many Requests")), 1000);
  assert.equal(retryDelayMs(1, new ApiError(503, "unavailable")), 2000);
  assert.equal(retryDelayMs(5, new TypeError("Failed to fetch")), 5000);
});
