import { test, expect } from "@playwright/test";

test.describe("Data section: motions and betänkanden coverage", () => {
  test("lists the riksmöten covered per source", async ({ page }) => {
    await page.route("**/api/motions/backfill", (route) =>
      route.fulfill({
        json: {
          data: [
            { riksmote: "2021/22", source: "motioner", expected: 4777, held: 4777, complete: true, completedAt: "2026-11-11T10:00:00Z" },
            { riksmote: "2021/22", source: "betankanden", expected: 300, held: 298, complete: true, completedAt: "2026-11-11T11:00:00Z" },
            { riksmote: "2018/19", source: "motioner", expected: 3129, held: 100, complete: false, completedAt: null },
          ],
        },
      }),
    );
    await page.goto("/data");
    const rows = page.getByTestId("motion-coverage-row");
    await expect(rows).toHaveCount(2);
    await expect(rows.nth(0)).toContainText("2021/22");
    await expect(rows.nth(0)).toContainText("klar · 4777 av 4777");
    await expect(rows.nth(0)).toContainText("klar · 298 av 300");
    await expect(rows.nth(1)).toContainText("delvis · 100 av 3129");
    await expect(rows.nth(1)).toContainText("ej inläst");
  });

  test("says so when nothing older is read in", async ({ page }) => {
    await page.route("**/api/motions/backfill", (route) => route.fulfill({ json: { data: [] } }));
    await page.goto("/data");
    await expect(page.getByTestId("motion-coverage-empty")).toBeVisible();
  });
});

test.describe("Data section: speeches coverage", () => {
  test("lists speeches fetched, linked and unlinked per riksmöte", async ({ page }) => {
    await page.route("**/api/speeches/backfill", (route) =>
      route.fulfill({
        json: {
          data: [
            { riksmote: "2021/22", expected: 9000, stored: 9000, linked: 6100, unlinked: 2900, textBytes: 31000000, rowsAdded: 9000, tableBytesAdded: 52000000, indexBytesAdded: 3000000, politiciansCreated: 4, unresolvedSpeakers: 0, complete: true, completedAt: "2026-11-11T10:00:00Z" },
            { riksmote: "2020/21", expected: 8000, stored: 120, linked: 100, unlinked: 20, textBytes: 400000, rowsAdded: 120, tableBytesAdded: 700000, indexBytesAdded: 60000, politiciansCreated: 0, unresolvedSpeakers: 0, complete: false, completedAt: null },
          ],
        },
      }),
    );
    await page.goto("/data");
    const rows = page.getByTestId("speech-coverage-row");
    await expect(rows).toHaveCount(2);
    await expect(rows.nth(0)).toContainText("2021/22");
    await expect(rows.nth(0)).toContainText("klar");
    await expect(rows.nth(0)).toContainText("9000 av 9000");
    await expect(rows.nth(0)).toContainText("6100");
    await expect(rows.nth(0)).toContainText("2900");
    await expect(rows.nth(1)).toContainText("delvis");
    await expect(rows.nth(1)).toContainText("120 av 8000");
  });

  test("says so when no riksmöte is read in", async ({ page }) => {
    await page.route("**/api/speeches/backfill", (route) => route.fulfill({ json: { data: [] } }));
    await page.goto("/data");
    await expect(page.getByTestId("speech-coverage-empty")).toBeVisible();
  });
});
