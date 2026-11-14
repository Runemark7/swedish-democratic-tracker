import { test, expect } from "@playwright/test";

// Shapes as backend/internal/apitest/loften_on_decisions_test.go asserts them.
const REFS = [
  { id: 11, party: "M", text: "Sänk skatten på kapital" },
  { id: 12, party: "S", text: "Stärk arbetsrätten" },
];

test.describe("Löften on a betänkande and a motion", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test("the betänkande page lists the Löften it touches and each links to the Löfte", async ({ page }) => {
    await page.route("**/api/betankanden/2021%2F22/SkU16", (r) => r.fulfill({
      json: {
        riksmote: "2021/22", beteckning: "SkU16", title: "Mervärdesskatt", organ: "SkU", dokId: "H901SkU16",
        sourceUrl: "https://data.riksdagen.se/dokument/H901SkU16.html", held: true, status: "x", decided: true,
        decisionDate: "2022-06-01", decisionText: "", punkter: [], debate: { date: null, total: 0, speeches: [] },
        loefteCount: 2, loften: REFS,
      },
    }));
    await page.route("**/api/loften/11", (r) => r.fulfill({ status: 404, json: { error: "x" } }));
    await page.goto("/beslut/2021%2F22/SkU16");
    const section = page.getByTestId("decision-loften");
    await expect(section.getByRole("heading")).toContainText("Berör 2 löften");
    await expect(section.getByTestId("decision-loefte")).toHaveCount(2);
    await section.getByRole("link", { name: "Sänk skatten på kapital" }).click();
    await expect(page).toHaveURL(/\/loften\/11$/);
  });

  test("a betänkande that touches no Löfte shows no section", async ({ page }) => {
    await page.route("**/api/betankanden/2021%2F22/SkU16", (r) => r.fulfill({
      json: {
        riksmote: "2021/22", beteckning: "SkU16", title: "Mervärdesskatt", organ: "SkU", dokId: "H901SkU16",
        sourceUrl: "x", held: true, status: "x", decided: true, decisionDate: "2022-06-01", decisionText: "",
        punkter: [], debate: { date: null, total: 0, speeches: [] }, loefteCount: 0, loften: [],
      },
    }));
    await page.goto("/beslut/2021%2F22/SkU16");
    await expect(page.getByRole("heading", { level: 1 })).toHaveText("Mervärdesskatt");
    await expect(page.getByTestId("decision-loften")).toHaveCount(0);
  });

  test("the party page summarises its Löften and links to them", async ({ page }) => {
    await page.route("**/api/loften/summary*", (r) => r.fulfill({
      json: [{ party: "M", total: 3, for: 2, emot: 0, blandat: 0, inteProvat: 1, arkiv: 0 }],
    }));
    await page.route("**/api/parties/M/record", (r) => r.fulfill({ json: { party: "M", decisions: [], motions: [] } }));
    await page.route("**/api/parties/M/valmanifest", (r) => r.fulfill({ json: [] }));
    await page.route("**/api/parties/M/loften", (r) => r.fulfill({ json: [] }));
    await page.route("**/api/loften?**", (r) => r.fulfill({ json: [] }));
    await page.route("**/api/valmanifest", (r) => r.fulfill({ json: [] }));
    await page.goto("/parties/M");
    const sum = page.getByTestId("party-loefte-summary");
    await expect(sum).toContainText("Röstat för 2");
    await sum.getByRole("link", { name: "Löften" }).click();
    await expect(page).toHaveURL(/\/loften\?party=M$/);
  });
});
