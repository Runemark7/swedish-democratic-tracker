import { test, expect } from "@playwright/test";

const API = process.env.API_URL ?? "http://localhost:8080";

test("GET /healthz returns {status: ok}", async ({ request }) => {
  const res = await request.get(`${API}/healthz`);
  expect(res.status()).toBe(200);
  const body = await res.json();
  expect(body).toMatchObject({ status: "ok" });
});
