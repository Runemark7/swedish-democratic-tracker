import { test, expect } from "@playwright/test";

// Shape of GET /api/motions/2017%2F18/3293, as recorded from Riksdagen on 2026-11-10.
const MOTION_3293 = {
  dokId: "H5023293",
  riksmote: "2017/18",
  beteckning: "3293",
  title: "Skattemässig hantering av virtuella varor",
  subtyp: "Kommittémotion",
  kind: "Kommittémotion",
  kindExplanation:
    "Väckt i namn av ett partis ledamöter i ett utskott och uttrycker oftast partiets linje på området.",
  date: "2017-10-05",
  utskott: "SkU",
  utskottName: "Skatteutskottet",
  status: "Klar",
  behandlasI: "2017/18:SkU11",
  sourceUrl: "https://data.riksdagen.se/dokument/H5023293.html",
  authors: [
    { politicianId: "0572788110724", name: "Maria Malmer Stenergard", party: "M", role: "undertecknare" },
    { politicianId: "0729710260118", name: "Helena Bouveng", party: "M", role: "undertecknare" },
  ],
  yrkanden: [
    {
      number: 1,
      wording:
        "Riksdagen ställer sig bakom det som anförs i motionen om att tillsätta en utredning för att se över den skatterättsliga bedömningen av handel med virtuella varor och tillkännager detta för regeringen.",
      utskottProposal: "Avslag",
      chamberDecision: "Avslag",
      behandlasI: "2017/18:SkU11",
      behandlasIPunkt: "",
    },
  ],
};

test.describe("motion page 2017/18:3293", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test("shows the header and the yrkande list", async ({ page }) => {
    let requested = "";
    await page.route("**/api/motions/**", async (route) => {
      requested = new URL(route.request().url()).pathname;
      await route.fulfill({ json: MOTION_3293 });
    });

    await page.goto("/motioner/2017%2F18/3293");

    await expect(page.getByRole("heading", { level: 1 })).toContainText("Skattemässig hantering av virtuella varor");
    await expect(page.getByText("Motion 2017/18:3293")).toBeVisible();
    await expect(page.getByTestId("motion-kind")).toContainText("Kommittémotion");
    await expect(page.getByTestId("motion-kind")).toContainText("uttrycker oftast partiets linje");
    await expect(page.getByText("5 oktober 2017")).toBeVisible();
    await expect(page.getByText("Skatteutskottet (SkU)")).toBeVisible();
    await expect(page.getByText("2017/18:SkU11").first()).toBeVisible();
    await expect(page.getByRole("link", { name: "Maria Malmer Stenergard" })).toBeVisible();
    await expect(page.getByRole("link", { name: /Läs motionen hos Riksdagen/ })).toHaveAttribute(
      "href",
      "https://data.riksdagen.se/dokument/H5023293.html",
    );

    const yrkanden = page.getByTestId("yrkande");
    await expect(yrkanden).toHaveCount(1);
    await expect(yrkanden.first()).toContainText("Yrkande 1");
    await expect(yrkanden.first()).toContainText("tillsätta en utredning");
    await expect(yrkanden.first()).toContainText("handel med virtuella varor");

    // The slash form is sent encoded, never as 2017-18.
    expect(requested).toBe("/api/motions/2017%2F18/3293");
    expect(page.url()).not.toContain("2017-18");
  });

  test("says so when the motion is not in the record", async ({ page }) => {
    await page.route("**/api/motions/**", (route) =>
      route.fulfill({ status: 404, json: { error: "motion not found" } }),
    );
    await page.goto("/motioner/2017%2F18/999999");
    await expect(page.getByText("finns inte i vårt register")).toBeVisible();
  });
});
