import { test, expect } from "@playwright/test";

// Shapes as backend/internal/apitest/loften_test.go asserts them.
const ARCHIVED = {
  id: 7, party: "M", valmanifestId: 3, text: "Fler poliser", topic: "trygghet",
  sourceDocument: "Valmanifest 2022", utskott: ["JuU"], mandatePeriod: "2022-2026",
  visibility: "archived", archived: true, createdAt: "2022-08-01T00:00:00Z",
};
const VALMANIFEST = [
  { id: 3, party: "M", electionYear: 2022, title: "Så får vi ordning på Sverige", url: "https://moderaterna.se/nyhet/moderaternasvalmanifest/", state: "final" },
  { id: 9, party: "M", electionYear: 2018, title: "Moderaternas valmanifest 2018", url: "https://snd.gu.se/sv/vivill/party/m/v/2018", state: "draft" },
];

test.describe("Löften under Partier", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test.beforeEach(async ({ page }) => {
    await page.route("**/api/parties/M/loften", (r) => r.fulfill({ json: [ARCHIVED] }));
    await page.route("**/api/parties/M/valmanifest", (r) => r.fulfill({ json: VALMANIFEST }));
    await page.route("**/api/parties/M/record", (r) => r.fulfill({ json: { party: "M", decisions: [], motions: [] } }));
    await page.route("**/api/loften", (r) => r.fulfill({ json: [ARCHIVED] }));
    await page.route("**/api/valmanifest", (r) => r.fulfill({ json: VALMANIFEST }));
    await page.route("**/api/loften/7", (r) => r.fulfill({ json: ARCHIVED }));
  });

  test("Löften are reached under Partier and there is no Löften item in the top menu", async ({ page }) => {
    await page.route("**/api/party-meta", (r) => r.fulfill({ json: [] }));
    await page.goto("/parties");
    const primary = page.getByRole("navigation", { name: "Huvudmeny" });
    await expect(primary.getByRole("link", { name: /Löfte|Manifest/ })).toHaveCount(0);

    await page.getByRole("link", { name: "Löften för alla partier" }).click();
    await expect(page).toHaveURL(/\/loften$/);
    await expect(page.getByTestId("loefte-card")).toContainText("Fler poliser");
    await expect(primary.getByRole("link", { name: /Löfte/ })).toHaveCount(0);

    await page.goto("/parties/M?tab=loften");
    await expect(page.getByTestId("archive-note")).toContainText("2022–2026");
    await expect(page.getByTestId("loefte-card")).toHaveCount(1);
    // The archive shows no Löftesstatus.
    await expect(page.getByText(/Röstat för|Röstat emot|Blandat|Inte prövat än/)).toHaveCount(0);
  });

  test("the party page lists each Valmanifest as a source with its draft or final marking", async ({ page }) => {
    await page.goto("/parties/M");
    const rows = page.getByTestId("valmanifest-row");
    await expect(rows).toHaveCount(2);
    await expect(rows.first()).toContainText("Slutligt");
    await expect(rows.nth(1)).toContainText("Utkast");
    await expect(rows.first().getByRole("link")).toHaveAttribute("href", /moderaterna\.se/);
  });

  test("the old party goal votes URL redirects to the archived Löfte with the same id", async ({ page }) => {
    await page.goto("/parties/M/goals/7/votes");
    await expect(page).toHaveURL(/\/loften\/7$/);
    await expect(page.getByTestId("loefte-text")).toHaveText("Fler poliser");
  });

  test("the old party goals URL redirects to that party's Löften", async ({ page }) => {
    await page.goto("/parties/M/goals");
    await expect(page).toHaveURL(/\/parties\/M\?tab=loften/);
  });

  test("the manifestos URL redirects to the party's Valmanifest list", async ({ page }) => {
    await page.goto("/manifestos?party=M");
    await expect(page).toHaveURL(/\/parties\/M\?tab=loften/);
    await expect(page.getByTestId("valmanifest-row")).toHaveCount(2);

    await page.goto("/manifestos");
    await expect(page).toHaveURL(/\/loften#valmanifest$/);
    await expect(page.getByTestId("valmanifest-row")).toHaveCount(2);
  });
});
