import { test, expect } from "@playwright/test";

// Shape of GET /api/parties/M/record, as the seam test in
// backend/internal/apitest/party_record_test.go asserts it.
const RECORD = {
  party: "M",
  decisions: [
    {
      riksmote: "2025/26", beteckning: "UbU30", title: "Skärpta villkor för friskolesektorn",
      utskott: "UbU", utskottName: "Utbildningsutskottet", date: "2026-08-13", decided: true,
      outcome: "Avgjord genom omröstning", partilinje: "Ja", voteringar: 1,
      sourceUrl: "https://data.riksdagen.se/dokument/HD01UbU30.html",
    },
  ],
  motions: [
    {
      dokId: "H5023293", riksmote: "2017/18", beteckning: "3293", title: "Moms på virtuella varor",
      subtyp: "kommittémotion", kind: "Kommittémotion", parties: ["M"], date: "2017-10-04",
      utskott: "SkU", utskottName: "Skatteutskottet", handledIn: ["2017/18:SkU11"],
      sourceUrl: "https://data.riksdagen.se/dokument/H5023293.html",
      outcome: { word: "Avslagen", counts: { Avslagen: 1 }, total: 1 },
    },
  ],
};

test.describe("Party page", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test.beforeEach(async ({ page }) => {
    await page.route("**/api/parties/M/record", (route) => route.fulfill({ json: RECORD }));
  });

  test("recent decisions show the Partilinje and link to the betänkande page", async ({ page }) => {
    await page.goto("/parties/M");
    const row = page.getByTestId("party-decision-row");
    await expect(row).toHaveCount(1);
    await expect(row.getByTestId("party-decision-partilinje")).toHaveText("Partilinje: Ja");
    await expect(row.getByRole("link", { name: /Skärpta villkor/ })).toHaveAttribute("href", "/beslut/2025/26/UbU30");
  });

  test("Kommittémotioner show their outcome and link to the motion page", async ({ page }) => {
    await page.goto("/parties/M?tab=motioner");
    const row = page.getByTestId("party-motion-row");
    await expect(row).toHaveCount(1);
    await expect(row.getByTestId("party-motion-outcome")).toHaveText("Avslagen");
    await expect(row.getByRole("link", { name: /Moms på virtuella varor/ })).toHaveAttribute("href", "/motioner/2017%2F18/3293");
  });

  test("the old goals URL still reaches the Löften tab", async ({ page }) => {
    await page.route("**/api/parties/M/goals", (route) => route.fulfill({ json: [] }));
    await page.goto("/parties/M/goals");
    await expect(page).toHaveURL(/\/parties\/M\?tab=loften/);
  });
});
