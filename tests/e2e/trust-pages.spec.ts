import { test, expect } from "@playwright/test";

// Gaps #301, #302, #303, #326: contact, corrections, method in the footer, who assesses.

const ISSUES = "https://github.com/Runemark7/swedish-democratic-tracker/issues/new";

test("the footer links Metod and Ändringslogg on every page", async ({ page }) => {
  for (const path of ["/", "/om-sajten", "/finns-inte-alls"]) {
    await page.goto(path);
    const footer = page.locator("footer");
    await expect(footer.getByRole("link", { name: "Metod", exact: true })).toHaveAttribute("href", "/metod");
    await expect(footer.getByRole("link", { name: "Ändringslogg", exact: true })).toHaveAttribute("href", "/andringslogg");
  }
});

test("Om sajten has a Kontakt section that points at the public issue tracker and Metod", async ({ page }) => {
  await page.goto("/om-sajten");
  const kontakt = page.getByTestId("kontakt");
  await expect(kontakt.getByRole("link", { name: "GitHub" })).toHaveAttribute("href", ISSUES);
  await expect(page.getByRole("link", { name: "Metod", exact: true }).first()).toBeVisible();
  await expect(page.getByRole("link", { name: "Ändringsloggen" })).toHaveAttribute("href", "/andringslogg");
});

test("Metod explains who assesses and how corrections work, without naming anyone", async ({ page }) => {
  await page.goto("/metod");
  const vem = page.getByTestId("vem-bedomer");
  await expect(vem).toContainText("Boten föreslår");
  await expect(vem).toContainText("Redaktören");
  await expect(vem).toContainText("loggas");
  const r = page.getByTestId("rattelser");
  await expect(r.getByRole("link", { name: "GitHub" })).toHaveAttribute("href", ISSUES);
  await expect(r.getByRole("link", { name: "Ändringsloggen" })).toHaveAttribute("href", "/andringslogg");
  await expect(page.getByTestId("metod")).not.toContainText("ansvarig utgivare");
});

test("the Ändringslogg page lists the public log", async ({ page }) => {
  await page.route("**/api/aendringslogg*", (route) =>
    route.fulfill({
      json: {
        entries: [
          {
            id: 7, createdAt: "2026-11-01T10:00:00Z", actor: "bot", onBehalfOf: "Redaktör",
            chatRef: "abc1", action: "undo", entity: "loefte", entityId: 3,
            oldValue: null, newValue: null, undoes: 5,
          },
        ],
      },
    }),
  );
  await page.goto("/andringslogg");
  await expect(page.getByRole("heading", { name: "Ändringslogg" })).toBeVisible();
  const list = page.getByTestId("andringslogg-list");
  await expect(list).toContainText("Ångrad");
  await expect(list).toContainText("loefte 3");
  await expect(list).toContainText("ångrar post 5");
  await expect(page).toHaveTitle("Riksdagskollen – Ändringslogg");
});
