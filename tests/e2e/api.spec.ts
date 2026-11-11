import { test, expect } from "@playwright/test";

const API = process.env.API_URL ?? "http://localhost:8080";

test.describe("Politicians API", () => {
  test("GET /api/politicians returns paginated response", async ({ request }) => {
    const res = await request.get(`${API}/api/politicians`);
    expect(res.status()).toBe(200);

    const body = await res.json();
    expect(body).toHaveProperty("data");
    expect(body).toHaveProperty("total");
    expect(body).toHaveProperty("page");
    expect(body).toHaveProperty("pageSize");
    expect(Array.isArray(body.data)).toBe(true);
    expect(body.total).toBeGreaterThan(0);

    const p = body.data[0];
    expect(p).toHaveProperty("intressentId");
    expect(p).toHaveProperty("firstName");
    expect(p).toHaveProperty("lastName");
    expect(p).toHaveProperty("party");
    expect(p).toHaveProperty("constituency");
  });

  test("GET /api/politicians?party=S filters by party", async ({ request }) => {
    const res = await request.get(`${API}/api/politicians?party=S`);
    expect(res.status()).toBe(200);

    const body = await res.json();
    expect(body.data.length).toBeGreaterThan(0);
    for (const p of body.data) {
      expect(p.party).toBe("S");
    }
  });

  test("GET /api/politicians/{id} returns politician detail", async ({ request }) => {
    // Get first politician from list then fetch by ID
    const listRes = await request.get(`${API}/api/politicians`);
    const list = await listRes.json();
    const id = list.data[0].intressentId;

    const res = await request.get(`${API}/api/politicians/${id}`);
    expect(res.status()).toBe(200);

    const p = await res.json();
    expect(p.intressentId).toBe(id);
    expect(p).toHaveProperty("firstName");
    expect(p).toHaveProperty("lastName");
    expect(p).toHaveProperty("party");
  });
});

test.describe("Votes API", () => {
  test("GET /api/votes/SoU7/1 returns vote detail with positions", async ({ request }) => {
    const res = await request.get(`${API}/api/votes/SoU7/1`);
    expect(res.status()).toBe(200);

    const body = await res.json();
    expect(body).toHaveProperty("beteckning", "SoU7");
    expect(body).toHaveProperty("forslagspunkt", "1");
    expect(body).toHaveProperty("positions");
    expect(Array.isArray(body.positions)).toBe(true);
    expect(body.positions.length).toBeGreaterThan(0);

    const pos = body.positions[0];
    expect(pos).toHaveProperty("party");
    expect(pos).toHaveProperty("voteResult");
  });

  test("GET /api/votes/FiU3/1 returns vote detail", async ({ request }) => {
    const res = await request.get(`${API}/api/votes/FiU3/1`);
    expect(res.status()).toBe(200);

    const body = await res.json();
    expect(body.beteckning).toBe("FiU3");
    expect(body.positions.length).toBeGreaterThan(0);
  });
});
