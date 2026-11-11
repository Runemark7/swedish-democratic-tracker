import { test, expect, type Page } from "@playwright/test";

const CORE = ["Beslut", "Motioner", "Ledamöter", "Partier", "Regeringen"] as const;

/** Labels that used to sit in the core menu. */
const NOT_IN_CORE = [
  "Omröstningar",
  "Politiker",
  "Manifest",
  "Budget",
  "Riksdag",
  "Regering",
  "Lokalt",
] as const;

async function expectShell(page: Page) {
  const primary = page.getByRole("navigation", { name: "Huvudmeny" });
  await expect(primary.getByRole("link")).toHaveText([...CORE]);
  for (const label of NOT_IN_CORE) {
    await expect(primary.getByRole("link", { name: label, exact: true })).toHaveCount(0);
  }

  const lokalt = page.getByRole("navigation", { name: "Lokalt" });
  await expect(lokalt.getByRole("link", { name: "Lokalt", exact: true })).toBeVisible();
  await expect(lokalt.getByRole("link", { name: "Lokalt", exact: true })).toHaveAttribute("href", "/lokalt");

  const secondary = page.getByRole("navigation", { name: "Mer" });
  await expect(secondary.getByRole("link", { name: "Om sajten", exact: true })).toBeVisible();
  await expect(secondary.getByRole("link", { name: "Data", exact: true })).toBeVisible();
  await expect(secondary.getByRole("link")).toHaveCount(2);

  await expect(page.getByRole("link", { name: "Riksdagskollen", exact: true })).toBeVisible();
  await expect(page).toHaveTitle(/Riksdagskollen/);
  await expect(page.locator('meta[name="description"]')).toHaveAttribute("content", /Riksdagskollen/);
  await expect(page.locator('meta[property="og:title"]')).toHaveAttribute("content", "Riksdagskollen");
  await expect(page.locator('meta[name="application-name"]')).toHaveAttribute("content", "Riksdagskollen");
}

test.describe("desktop nav", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test("shows Beslut, Motioner, Ledamöter, Partier and Regeringen, with Lokalt beside them", async ({ page }) => {
    await page.goto("/");
    await expectShell(page);

    const primary = page.getByRole("navigation", { name: "Huvudmeny" });
    await expect(primary.getByRole("link", { name: "Beslut", exact: true })).toHaveAttribute("href", "/beslut");
    await expect(primary.getByRole("link", { name: "Motioner", exact: true })).toHaveAttribute("href", "/motioner");
    await expect(primary.getByRole("link", { name: "Ledamöter", exact: true })).toHaveAttribute("href", "/politicians");
    await expect(primary.getByRole("link", { name: "Partier", exact: true })).toHaveAttribute("href", "/parties");
    await expect(primary.getByRole("link", { name: "Regeringen", exact: true })).toHaveAttribute("href", "/regeringen");

    await page.getByRole("button", { name: "Visa kommuner och regioner" }).click();
    const menu = page.getByRole("menu");
    await expect(menu.getByRole("menuitem", { name: "Kommuner" })).toBeVisible();
    await expect(menu.getByRole("menuitem", { name: "Regioner" })).toBeVisible();
    await menu.getByRole("menuitem", { name: "Kommuner" }).click();
    await expect(page).toHaveURL(/\/kommun$/);
    await expect(page.getByRole("heading", { name: "Välj kommun" })).toBeVisible();

    await page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Beslut", exact: true }).click();
    await expect(page).toHaveURL(/\/beslut$/);
    await expect(
      page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Beslut", exact: true }),
    ).toHaveAttribute("aria-current", "page");
  });
});

test.describe("mobile nav", () => {
  test.use({ viewport: { width: 390, height: 844 } });

  test("reads the same menu, with Lokalt and secondary links off the core row", async ({ page }) => {
    await page.goto("/");
    await expectShell(page);
    await page.getByRole("button", { name: "Visa kommuner och regioner" }).click();
    await expect(page.getByRole("menuitem", { name: "Kommuner" })).toBeVisible();
    await page.getByRole("menuitem", { name: "Regioner" }).click();
    await expect(page).toHaveURL(/\/region$/);
    await expect(page.getByText("Välj region")).toBeVisible();
  });
});

test.describe("routes the menu replaced", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test("Beslut stays active on the list and on old vote and beslut URLs", async ({ page }) => {
    const beslut = () =>
      page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Beslut", exact: true });

    await page.goto("/votes");
    await expect(page).toHaveURL(/\/beslut$/);
    await expect(page).toHaveTitle("Riksdagskollen – Beslut");
    await expect(beslut()).toHaveAttribute("aria-current", "page");
    await expect(page.getByRole("heading", { level: 1, name: "Beslut" })).toBeVisible();

    await page.goto("/beslut");
    await expect(beslut()).toHaveAttribute("aria-current", "page");

    await page.goto("/votes/AU9/1");
    await expect(page).toHaveURL(/\/votes\/AU9\/1$/);
    await expect(beslut()).toHaveAttribute("aria-current", "page");
    await expect(page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Partier", exact: true }))
      .not.toHaveAttribute("aria-current", "page");

    await page.goto("/beslut/SkU11");
    await expect(page).toHaveURL(/\/beslut\/SkU11$/);
    await expect(beslut()).toHaveAttribute("aria-current", "page");

    await page.goto("/votes/AU9/1/dela");
    await expect(page).toHaveURL(/\/votes\/AU9\/1\/dela$/);
    await expect(page).toHaveTitle(/Riksdagskollen/);
  });

  test("bookmarked pages still open, and manifestos stays off the menu", async ({ page }) => {
    await page.route("**/api/loften", (r) => r.fulfill({ json: [] }));
    await page.route("**/api/valmanifest", (r) => r.fulfill({ json: [] }));
    await page.goto("/manifestos");
    await expect(page).toHaveURL(/\/loften#valmanifest$/);
    await expect(page.getByRole("heading", { name: "Valmanifest (källor)" })).toBeVisible();
    await expect(page).toHaveTitle("Riksdagskollen – Löften");
    await expect(
      page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Manifest", exact: true }),
    ).toHaveCount(0);

    await page.goto("/budget");
    await expect(page).toHaveURL(/\/regeringen#budget$/);
    await expect(page.getByRole("heading", { name: "Statsbudgeten" })).toBeVisible();
    await expect(
      page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Regeringen", exact: true }),
    ).toHaveAttribute("aria-current", "page");

    await page.goto("/budget/areas/1");
    await expect(page).toHaveURL(/\/regeringen\/budget\/1$/);

    await page.goto("/riksdag");
    await expect(page).toHaveURL(/\/regeringen#riksdag$/);
    await expect(page).toHaveTitle("Riksdagskollen – Regeringen");

    await page.goto("/regering");
    await expect(page).toHaveURL(/\/regeringen$/);
    await expect(page.getByRole("heading", { level: 2, name: /^Regeringen/ })).toBeVisible();
    await expect(
      page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Regeringen", exact: true }),
    ).toHaveAttribute("aria-current", "page");

    await page.goto("/politicians");
    await expect(page).toHaveURL(/\/politicians$/);
    await expect(
      page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Ledamöter", exact: true }),
    ).toHaveAttribute("aria-current", "page");

    await page.goto("/parties/S/goals");
    await expect(page).toHaveURL(/\/parties\/S\?tab=loften$/);

    await page.goto("/kommun");
    await expect(page.getByRole("heading", { name: "Välj kommun" })).toBeVisible();
    await expect(
      page.getByRole("navigation", { name: "Lokalt" }).getByRole("link", { name: "Lokalt", exact: true }),
    ).toHaveAttribute("aria-current", "page");
    await expect(
      page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Beslut", exact: true }),
    ).not.toHaveAttribute("aria-current", "page");

    await page.goto("/region");
    await expect(page.getByText("Välj region")).toBeVisible();
    await expect(
      page.getByRole("navigation", { name: "Lokalt" }).getByRole("link", { name: "Lokalt", exact: true }),
    ).toHaveAttribute("aria-current", "page");

    await page.goto("/lokalt");
    await expect(page.getByRole("heading", { name: "Kommuner och regioner" })).toBeVisible();

    await page.goto("/om-sajten");
    await expect(page).toHaveTitle("Riksdagskollen – Om sajten");
    await expect(
      page.getByRole("navigation", { name: "Mer" }).getByRole("link", { name: "Om sajten", exact: true }),
    ).toHaveAttribute("aria-current", "page");

    await page.goto("/data");
    await expect(page.getByRole("heading", { name: "Var kommer datan ifrån?" })).toBeVisible();
    await expect(page).toHaveTitle("Riksdagskollen – Data");
  });
});
