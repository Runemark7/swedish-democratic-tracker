import { test, expect, type Page } from "@playwright/test";

const THEMES = ["light", "dark"] as const;

async function expectTheme(page: Page, theme: (typeof THEMES)[number]) {
  const attr = await page.locator("html").getAttribute("data-theme");
  if (theme === "light") {
    expect(attr).toBe("light");
  } else {
    expect(attr ?? "").toBe("");
  }
}

function regeringenLink(page: Page) {
  return page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Regeringen", exact: true });
}

for (const theme of THEMES) {
  test.describe(`regeringen redirects and sections (${theme})`, () => {
    test.use({ viewport: { width: 1280, height: 800 } });

    test.beforeEach(async ({ page }) => {
      await page.addInitScript((value) => {
        localStorage.setItem("sdt-theme", value);
      }, theme);
    });

    test("sends budget, riksdag and regering to Regeringen, and keeps the section in view", async ({ page }) => {
      await page.goto("/budget");
      await expect(page).toHaveURL(/\/regeringen#budget$/);
      await expectTheme(page, theme);
      await expect(page.locator("#budget")).toBeInViewport();
      await expect(page.locator("#budget").getByRole("heading", { name: "Statsbudgeten" })).toBeVisible();
      await expect(page).toHaveTitle("Riksdagskollen – Regeringen");
      await expect(
        page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Budget", exact: true }),
      ).toHaveCount(0);
      await expect(regeringenLink(page)).toHaveAttribute("aria-current", "page");
      await expect(regeringenLink(page)).toHaveAttribute("href", "/regeringen");

      await page.goto("/budget/areas/1");
      await expect(page).toHaveURL(/\/regeringen\/budget\/1$/);
      await expect(page.getByText(/Laddar\.\.\.|Kunde inte ladda data\./)).toBeVisible();
      await expect(regeringenLink(page)).toHaveAttribute("aria-current", "page");
      await expect(
        page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Budget", exact: true }),
      ).toHaveCount(0);

      await page.goto("/riksdag");
      await expect(page).toHaveURL(/\/regeringen#riksdag$/);
      await expect(page.locator("#riksdag")).toBeInViewport();
      await expect(page.locator("#riksdag").getByRole("heading", { name: "Riksdag", exact: true })).toBeVisible();
      await expect(page.locator("#riksdag").getByText("Riksdagen")).toBeVisible();
      await expect(
        page.getByRole("navigation", { name: "Huvudmeny" }).getByRole("link", { name: "Riksdag", exact: true }),
      ).toHaveCount(0);
      await expect(regeringenLink(page)).toHaveAttribute("aria-current", "page");

      await page.goto("/regering");
      await expect(page).toHaveURL(/\/regeringen$/);
      await expect(page.locator("#statsrad")).toBeInViewport();
      await expect(page.getByRole("heading", { level: 2, name: "Regeringen" })).toBeVisible();
      await expect(regeringenLink(page)).toHaveAttribute("aria-current", "page");
    });

    test("section anchors scroll, and minister and agency URLs stay", async ({ page }) => {
      await page.goto("/regeringen");
      await expectTheme(page, theme);
      const sections = page.getByRole("navigation", { name: "Avsnitt på sidan" });

      await sections.getByRole("link", { name: "Budget", exact: true }).click();
      await expect(page).toHaveURL(/\/regeringen#budget$/);
      await expect(page.locator("#budget")).toBeInViewport();

      await sections.getByRole("link", { name: "Riksdag", exact: true }).click();
      await expect(page).toHaveURL(/\/regeringen#riksdag$/);
      await expect(page.locator("#riksdag")).toBeInViewport();

      await sections.getByRole("link", { name: "Myndigheter", exact: true }).click();
      await expect(page).toHaveURL(/\/regeringen#myndigheter$/);
      await expect(page.locator("#myndigheter")).toBeInViewport();
      const register = page.locator("#myndigheter").getByRole("link", { name: /fler myndigheter/i });
      await expect(register).toHaveAttribute("href", "/riksdag/myndigheter");

      await sections.getByRole("link", { name: "Statsråd", exact: true }).click();
      await expect(page).toHaveURL(/\/regeringen#statsrad$/);
      await expect(page.locator("#statsrad")).toBeInViewport();

      await register.click();
      await expect(page).toHaveURL(/\/riksdag\/myndigheter$/);
      await expect(page.getByText("Sveriges myndigheter")).toBeVisible();
      await expect(regeringenLink(page)).toHaveAttribute("aria-current", "page");

      await page.goto("/riksdag/myndigheter/arbetsformedlingen");
      await expect(page).toHaveURL(/\/riksdag\/myndigheter\/arbetsformedlingen$/);
      await expect(regeringenLink(page)).toHaveAttribute("aria-current", "page");
      await expect(page.getByRole("link", { name: /Myndigheter/ })).toBeVisible();

      await page.goto("/regering/minister-1");
      await expect(page).toHaveURL(/\/regering\/minister-1$/);
      await expect(page.getByText(/Laddar\.\.\.|Kunde inte ladda statsrådet\./)).toBeVisible();
      await expect(regeringenLink(page)).toHaveAttribute("aria-current", "page");

      await page.goto("/regeringen/budget/9");
      await expect(page).toHaveURL(/\/regeringen\/budget\/9$/);
      await expect(regeringenLink(page)).toHaveAttribute("aria-current", "page");
    });
  });
}
