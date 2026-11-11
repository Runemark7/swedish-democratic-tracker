import { test, expect } from "@playwright/test";

// Shape of GET /api/beslut, as the seam test in backend/internal/apitest
// (beslut_list_test.go) asserts it against Riksdagen's 2026-11-11 listing.
const PARTIES = ["S", "M", "SD", "C", "V", "KD", "MP", "L"];
const LINE: Record<string, string> = { C: "Ja", KD: "Ja", L: "Ja", M: "Ja", SD: "Ja", S: "Nej", MP: "Nej", V: "Nej" };

const UPCOMING = {
  riksmote: "2026/27",
  beteckning: "SkU8",
  title: "Nya förenklingsregler och andra kompletteringar i lagen om tilläggsskatt",
  utskott: "SkU",
  utskottName: "Skatteutskottet",
  date: "2026-11-25",
  decided: false,
  status: "planerat",
  outcome: "Ej avgjord än",
  sourceUrl: "https://data.riksdagen.se/dokument/HE01SkU8.html",
  parties: [],
};
const DECIDED = {
  riksmote: "2025/26",
  beteckning: "UbU30",
  title: "Skärpta villkor för friskolesektorn",
  utskott: "UbU",
  utskottName: "Utbildningsutskottet",
  date: "2026-08-13",
  decided: true,
  status: "Webbpublicering",
  outcome: "Avgjord genom omröstning",
  sourceUrl: "https://data.riksdagen.se/dokument/HD01UbU30.html",
  parties: PARTIES.map((party) => ({ party, partilinje: LINE[party], voteringar: 1 })),
};

test.describe("Beslut list", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test("shows upcoming rows apart from decided ones, each linking to the canonical page", async ({ page }) => {
    await page.route("**/api/beslut?**", (route) => route.fulfill({ json: { data: [UPCOMING, DECIDED], nextCursor: null } }));
    await page.goto("/beslut");

    await expect(page.getByRole("heading", { level: 1 })).toHaveText("Beslut");
    const upcoming = page.getByRole("region", { name: "Kommande beslut" });
    const decided = page.getByRole("region", { name: "Fattade beslut" });
    await expect(upcoming.getByTestId("beslut-row-upcoming")).toHaveCount(1);
    await expect(decided.getByTestId("beslut-row")).toHaveCount(1);
    await expect(upcoming.getByTestId("row-date")).toContainText("Planerad beslutsdag 25 november 2026");
    await expect(upcoming.getByTestId("party-mark")).toHaveCount(0);

    await expect(decided.getByTestId("row-date")).toContainText("Beslutad 13 augusti 2026");
    await expect(decided.getByTestId("row-outcome")).toHaveText("Avgjord genom omröstning");
    await expect(decided.getByTestId("party-mark")).toHaveCount(8);
    await expect(decided.locator('[data-party="S"]')).toContainText("Nej");
    await expect(decided.locator('[data-party="M"]')).toContainText("Ja");
    await expect(decided.locator('[data-party="S"]')).toHaveAttribute("href", "/parties/S");

    await expect(upcoming.getByRole("link", { name: /Nya förenklingsregler/ })).toHaveAttribute("href", "/beslut/2026/27/SkU8");
    await expect(decided.getByRole("link", { name: /Skärpta villkor/ })).toHaveAttribute("href", "/beslut/2025/26/UbU30");
  });

  test("filters change the URL and are sent to the API; a shared URL restores them", async ({ page }) => {
    const requests: string[] = [];
    await page.route("**/api/beslut?**", (route) => {
      requests.push(new URL(route.request().url()).search);
      return route.fulfill({ json: { data: [DECIDED], nextCursor: null } });
    });
    await page.goto("/beslut");

    await page.getByLabel("Status", { exact: true }).selectOption("decided");
    await expect(page).toHaveURL(/[?&]state=decided/);
    await page.getByLabel("Utskott", { exact: true }).selectOption("UbU");
    await expect(page).toHaveURL(/[?&]utskott=UbU/);
    await page.getByLabel("Parti", { exact: true }).selectOption("V");
    await expect(page).toHaveURL(/[?&]party=V/);
    await expect
      .poll(() => requests[requests.length - 1])
      .toMatch(/^(?=.*state=decided)(?=.*utskott=UbU)(?=.*party=V)/);

    const shared = page.url();
    await page.goto("/");
    await page.goto(shared);
    await expect(page.getByLabel("Status", { exact: true })).toHaveValue("decided");
    await expect(page.getByLabel("Utskott", { exact: true })).toHaveValue("UbU");
    await expect(page.getByLabel("Parti", { exact: true })).toHaveValue("V");
  });

  test("Nästa follows the cursor", async ({ page }) => {
    const cursors: (string | null)[] = [];
    await page.route("**/api/beslut?**", (route) => {
      const c = new URL(route.request().url()).searchParams.get("cursor");
      cursors.push(c);
      return route.fulfill({ json: c ? { data: [DECIDED], nextCursor: null } : { data: [UPCOMING], nextCursor: "abc" } });
    });
    await page.goto("/beslut");
    await page.getByRole("button", { name: /Nästa/ }).click();
    await expect(page.getByTestId("beslut-row")).toHaveCount(1);
    expect(cursors).toEqual([null, "abc"]);
  });

  test("the old votes list URL redirects to Beslut and the menu points there", async ({ page }) => {
    await page.route("**/api/beslut?**", (route) => route.fulfill({ json: { data: [], nextCursor: null } }));
    await page.goto("/votes");
    await expect(page).toHaveURL(/\/beslut$/);
    await expect(page.getByRole("heading", { level: 1 })).toHaveText("Beslut");
    const primary = page.getByRole("navigation", { name: "Huvudmeny" });
    await expect(primary.getByRole("link", { name: "Beslut", exact: true })).toHaveAttribute("href", "/beslut");
  });
});
