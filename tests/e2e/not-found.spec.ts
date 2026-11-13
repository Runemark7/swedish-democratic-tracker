import { test, expect } from "@playwright/test";

// Gap unknown-url-blank-page: unmatched URLs must say so, not render a blank shell.
for (const path of ["/finns-inte-alls", "/samrostning", "/party/S", "/promises/47", "/beslut/2024/25/ett/två/tre"]) {
  test(`unknown path ${path} shows a not-found page`, async ({ page }) => {
    await page.goto(path);
    await expect(page.getByRole("heading", { name: "Sidan finns inte" })).toBeVisible();
    const nf = page.getByTestId("not-found");
    for (const name of ["Startsidan", "Beslut", "Motioner", "Ledamöter", "Partier"]) {
      await expect(nf.getByRole("link", { name, exact: true })).toBeVisible();
    }
  });
}

test("known route is not the not-found page", async ({ page }) => {
  await page.goto("/metod");
  await expect(page.getByTestId("not-found")).toHaveCount(0);
});
