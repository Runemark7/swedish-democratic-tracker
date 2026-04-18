import { test, expect } from "@playwright/test";

test.describe("Parties page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
  });

  test("shows heading", async ({ page }) => {
    await expect(page.getByRole("heading", { name: "Partier" })).toBeVisible();
  });

  test("shows party cards without error state", async ({ page }) => {
    // Wait for at least one party card link
    await page.waitForSelector('a[href*="/parties/"]', { timeout: 15_000 });
    await expect(page.getByText("Kunde inte ladda partier")).not.toBeVisible();
  });

  test("shows party badges from seed data", async ({ page }) => {
    await page.waitForSelector('a[href*="/parties/"]', { timeout: 15_000 });
    // All seeded parties with goals should be visible
    for (const party of ["S", "M", "SD", "V", "MP"]) {
      await expect(
        page.locator(`a[href*="/parties/"] span`).filter({ hasText: party }).first()
      ).toBeVisible();
    }
  });

  test("shows alignment percentage", async ({ page }) => {
    await page.waitForSelector('a[href*="/parties/"]', { timeout: 15_000 });
    // At least one card should show a percentage value
    await expect(page.getByText(/%/).first()).toBeVisible();
  });

  test("navigates to party goals on card click", async ({ page }) => {
    await page.waitForSelector('a[href*="/parties/S/goals"]', { timeout: 15_000 });
    await page.click('a[href*="/parties/S/goals"]');
    await expect(page).toHaveURL(/\/parties\/S\/goals/);
  });
});
