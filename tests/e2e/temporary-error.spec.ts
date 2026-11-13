import { test, expect } from "@playwright/test";

// Gap rate-limit-shown-as-not-found: when our own API answered 429 the pages
// said the record "finns inte i vårt register". A 429 or 5xx is a temporary
// failure; only a 404 means the record is missing.

const TEMPORARY = "Tillfälligt fel";
const LOEFTE = {
  id: 7, party: "M", valmanifestId: null, text: "Fler poliser", topic: "trygghet",
  sourceDocument: "Valmanifest 2022", utskott: ["JuU"], mandatePeriod: "2022-2026",
  visibility: "archived", archived: true, createdAt: "2022-08-01T00:00:00Z",
};

const tooMany = { status: 429, headers: { "Retry-After": "1" }, contentType: "text/plain", body: "Too Many Requests\n" };
const notFound = { status: 404, json: { error: "not found" } };

test.describe("temporary API failure is not shown as a missing record", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test("beslut page: 429 on every attempt shows a temporary error and a retry button, not 'finns inte'", async ({ page }) => {
    let calls = 0;
    let healthy = false;
    await page.route("**/api/betankanden/**", async (route) => {
      calls++;
      if (healthy) return route.fulfill(notFound); // proves the button refetches
      return route.fulfill(tooMany);
    });
    await page.goto("/beslut/2026/27/KU4");
    await expect(page.getByTestId("temporary-error")).toContainText(TEMPORARY, { timeout: 15_000 });
    await expect(page.getByText("finns inte i vårt register")).toHaveCount(0);
    expect(calls).toBe(3); // first attempt and two automatic retries

    healthy = true;
    await page.getByRole("button", { name: "Försök igen" }).click();
    await expect(page.getByText("Betänkande 2026/27:KU4 finns inte i vårt register.")).toBeVisible();
  });

  test("beslut page: a 429 followed by a good answer recovers without the reader doing anything", async ({ page }) => {
    let calls = 0;
    await page.route("**/api/speeches**", (route) => route.fulfill({ json: [] }));
    await page.route("**/api/betankanden/**", async (route) => {
      calls++;
      if (calls === 1) return route.fulfill(tooMany);
      return route.fulfill({
        json: {
          riksmote: "2026/27", beteckning: "KU4", title: "Riksrevisionens rapport", organ: "KU", dokId: "HE01KU4",
          sourceUrl: "https://data.riksdagen.se/dokument/HE01KU4.html", held: true, status: "Webbpublicering",
          decided: false, decisionDate: "", decisionText: "", punkter: [],
        },
      });
    });
    await page.goto("/beslut/2026/27/KU4");
    await expect(page.getByRole("heading", { level: 1 })).toContainText("Riksrevisionens rapport", { timeout: 15_000 });
    await expect(page.getByTestId("temporary-error")).toHaveCount(0);
  });

  test("beslut page: a real 404 keeps the not-found copy and is not retried", async ({ page }) => {
    let calls = 0;
    await page.route("**/api/betankanden/**", async (route) => {
      calls++;
      return route.fulfill(notFound);
    });
    await page.goto("/beslut/2026/27/KU999");
    await expect(page.getByText("Betänkande 2026/27:KU999 finns inte i vårt register.")).toBeVisible();
    await expect(page.getByTestId("temporary-error")).toHaveCount(0);
    expect(calls).toBe(1);
  });

  test("beslut page: a 503 is a temporary error too", async ({ page }) => {
    await page.route("**/api/betankanden/**", (route) => route.fulfill({ status: 503, json: { error: "unavailable" } }));
    await page.goto("/beslut/2026/27/KU4");
    await expect(page.getByTestId("temporary-error")).toContainText(TEMPORARY, { timeout: 15_000 });
    await expect(page.getByText("finns inte i vårt register")).toHaveCount(0);
  });

  test("old beslut URL: 429 on the lookup is a temporary error, 404 is not found", async ({ page }) => {
    await page.route("**/api/betankanden/resolve**", (route) => route.fulfill(tooMany));
    await page.goto("/beslut/KU4");
    await expect(page.getByTestId("temporary-error")).toContainText(TEMPORARY, { timeout: 15_000 });
    await expect(page.getByText("finns inte i vårt register")).toHaveCount(0);

    await page.unroute("**/api/betankanden/resolve**");
    await page.route("**/api/betankanden/resolve**", (route) => route.fulfill(notFound));
    await page.goto("/beslut/KU999");
    await expect(page.getByText("Betänkandet finns inte i vårt register.")).toBeVisible();
  });

  test("löfte page: 429 shows a temporary error, not 'Löftet finns inte' and not an endless 'Laddar'", async ({ page }) => {
    let healthy = false;
    await page.route("**/api/loften/7", (route) => (healthy ? route.fulfill({ json: LOEFTE }) : route.fulfill(tooMany)));
    await page.goto("/loften/7");
    await expect(page.getByTestId("temporary-error")).toContainText(TEMPORARY, { timeout: 15_000 });
    await expect(page.getByText("Löftet finns inte")).toHaveCount(0);
    await expect(page.getByText("Laddar...")).toHaveCount(0);

    healthy = true;
    await page.getByRole("button", { name: "Försök igen" }).click();
    await expect(page.getByTestId("loefte-text")).toHaveText("Fler poliser");
  });

  test("löfte page: a real 404 keeps the not-found copy", async ({ page }) => {
    await page.route("**/api/loften/999", (route) => route.fulfill(notFound));
    await page.goto("/loften/999");
    await expect(page.getByText("Löftet finns inte.")).toBeVisible();
    await expect(page.getByTestId("temporary-error")).toHaveCount(0);
  });

  test("löften list: 429 ends in a temporary error with a retry button", async ({ page }) => {
    await page.route("**/api/valmanifest", (route) => route.fulfill({ json: [] }));
    await page.route("**/api/loften", (route) => route.fulfill(tooMany));
    await page.route("**/api/loften?**", (route) => route.fulfill(tooMany));
    await page.goto("/loften");
    await expect(page.getByTestId("temporary-error")).toContainText(TEMPORARY, { timeout: 15_000 });
    await expect(page.getByText("Laddar...")).toHaveCount(0);
  });

  test("motion page: 429 is a temporary error, 404 is not found", async ({ page }) => {
    await page.route("**/api/motions/**", (route) => route.fulfill(tooMany));
    // The riksmöte sits in one path segment, so its slash is encoded.
    await page.goto("/motioner/2025%2F26/123");
    await expect(page.getByTestId("temporary-error")).toContainText(TEMPORARY, { timeout: 15_000 });
    await expect(page.getByText("finns inte i vårt register")).toHaveCount(0);

    await page.unroute("**/api/motions/**");
    await page.route("**/api/motions/**", (route) => route.fulfill(notFound));
    await page.goto("/motioner/2025%2F26/999999");
    await expect(page.getByText("Motion 2025/26:999999 finns inte i vårt register.")).toBeVisible();
    await expect(page.getByTestId("temporary-error")).toHaveCount(0);
  });

  test("header status line: a 429 on coverage does not claim 'Ingen omröstningsdata'", async ({ page }) => {
    await page.route("**/api/riksdag/coverage", (route) => route.fulfill(tooMany));
    await page.route("**/api/betankanden/**", (route) => route.fulfill(notFound));
    await page.goto("/beslut/2026/27/KU4");
    await expect(page.getByText("finns inte i vårt register")).toBeVisible();
    await expect(page.getByTestId("header-status")).toContainText("kunde inte hämtas", { timeout: 15_000 });
    await expect(page.getByText("Ingen omröstningsdata")).toHaveCount(0);
  });
});
