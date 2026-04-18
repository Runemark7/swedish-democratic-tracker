import { test, expect } from "@playwright/test";

test.describe("Politicians page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/politicians");
  });

  test("shows heading", async ({ page }) => {
    await expect(page.getByRole("heading", { name: "Politiker" })).toBeVisible();
  });

  test("shows politician rows from seed data", async ({ page }) => {
    // seed has 11 politicians
    await expect(
      page.locator('a[href*="/politicians/"]').first()
    ).toBeVisible({ timeout: 15_000 });

    const rows = page.locator('a[href*="/politicians/"]');
    await expect(rows).toHaveCount(11);
  });

  test("party filter narrows results", async ({ page }) => {
    await page.waitForSelector('a[href*="/politicians/"]', { timeout: 15_000 });

    // Click "S" party filter
    await page.getByRole("button", { name: "S" }).click();

    // seed has 2 S politicians (Anna Andersson, Björn Bergström)
    await expect(page.locator('a[href*="/politicians/"]')).toHaveCount(2);
  });

  test("'Alla' filter restores full list", async ({ page }) => {
    await page.waitForSelector('a[href*="/politicians/"]', { timeout: 15_000 });

    await page.getByRole("button", { name: "S" }).click();
    await expect(page.locator('a[href*="/politicians/"]')).toHaveCount(2);

    await page.getByRole("button", { name: "Alla" }).click();
    await expect(page.locator('a[href*="/politicians/"]')).toHaveCount(11);
  });

  test("clicking a politician navigates to detail", async ({ page }) => {
    await page.waitForSelector('a[href*="/politicians/"]', { timeout: 15_000 });
    await page.locator('a[href*="/politicians/"]').first().click();
    await expect(page).toHaveURL(/\/politicians\/.+/);
  });
});
