import { test } from "node:test";
import assert from "node:assert/strict";
import { checkSource } from "../verify-sources.mjs";

function mockFetch(status, contentType = "application/json", body = "ok") {
  return async (_url, _opts) => ({
    status,
    ok: status >= 200 && status < 300,
    headers: new Map([["content-type", contentType]]),
    text: async () => body,
  });
}

test("api source returning 200 is verified", async () => {
  const result = await checkSource(
    { id: "x", kind: "api", upstream: "https://example.com" },
    { fetch: mockFetch(200, "application/json") }
  );
  assert.equal(result.status, "ok");
  assert.equal(result.httpCode, 200);
});

test("api source returning 404 is unreachable", async () => {
  const result = await checkSource(
    { id: "x", kind: "api", upstream: "https://example.com" },
    { fetch: mockFetch(404) }
  );
  assert.equal(result.status, "unreachable");
  assert.equal(result.httpCode, 404);
});

test("api source returning 405 falls back to GET", async () => {
  let calls = 0;
  const fakeFetch = async (_url, opts) => {
    calls++;
    if (opts.method === "HEAD") return { status: 405, ok: false, headers: new Map(), text: async () => "" };
    return { status: 200, ok: true, headers: new Map([["content-type", "application/json"]]), text: async () => "{}" };
  };
  const result = await checkSource(
    { id: "x", kind: "api", upstream: "https://example.com" },
    { fetch: fakeFetch }
  );
  assert.equal(result.status, "ok");
  assert.equal(calls, 2);
});

test("csv source requires content-type to contain csv|zip|octet-stream", async () => {
  const ok = await checkSource(
    { id: "x", kind: "csv", upstream: "https://example.com/file.zip" },
    { fetch: mockFetch(200, "application/zip") }
  );
  assert.equal(ok.status, "ok");

  const bad = await checkSource(
    { id: "x", kind: "csv", upstream: "https://example.com/file.html" },
    { fetch: mockFetch(200, "text/html") }
  );
  assert.equal(bad.status, "wrong-content-type");
});

test("seed and synthesized sources are skipped", async () => {
  for (const kind of ["seed", "synthesized"]) {
    const result = await checkSource(
      { id: "x", kind, upstream: null },
      { fetch: mockFetch(500) }
    );
    assert.equal(result.status, "skipped");
  }
});

test("timeout produces timeout status", async () => {
  const fakeFetch = async () => {
    const err = new Error("timed out");
    err.name = "AbortError";
    throw err;
  };
  const result = await checkSource(
    { id: "x", kind: "api", upstream: "https://example.com" },
    { fetch: fakeFetch }
  );
  assert.equal(result.status, "timeout");
});
