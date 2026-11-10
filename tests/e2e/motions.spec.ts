import { test, expect } from "@playwright/test";

// Shape of GET /api/motions/2017%2F18/3293, as recorded from Riksdagen on 2026-11-10.
const DECISION_3293 = {
  betankandeRiksmote: "2017/18",
  betankandeBeteckning: "SkU11",
  betankandeTitle: "Mervärdesskatt",
  betankandeSourceUrl: "https://data.riksdagen.se/dokument/H501SkU11.html",
  punkt: "6",
  rubrik: "Övriga frågor",
  linkedBy: "forslag_text",
  beslutstyp: "acklamation",
  mode: "acklamation",
  modeLabel: "Avgjord utan omröstning",
  alternativ: "",
  outcome: "Avslagen",
  voteringStored: false,
  parties: [],
};

const MOTION_3293 = {
  outcome: { word: "Avslagen", counts: { Avslagen: 1 }, total: 1 },
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
      outcome: "Avslagen",
      decisions: [DECISION_3293],
    },
  ],
};

// 2021/22:3638 yrkande 9, decided by votering in 2021/22:SkU16 punkt 6.
const MOTION_3638 = {
  ...MOTION_3293,
  dokId: "H9023638",
  riksmote: "2021/22",
  beteckning: "3638",
  behandlasI: "2021/22:SkU16",
  yrkanden: [
    {
      ...MOTION_3293.yrkanden[0],
      decisions: [
        {
          ...DECISION_3293,
          betankandeRiksmote: "2021/22",
          betankandeBeteckning: "SkU16",
          betankandeSourceUrl: "https://data.riksdagen.se/dokument/H901SkU16.html",
          beslutstyp: "röstning",
          mode: "votering",
          modeLabel: "Avgjord genom omröstning",
          alternativ: "Utskottets förslag mot reservation 12 (M)",
          voteringStored: true,
          parties: [
            { party: "M", partilinje: "Nej", ja: 0, nej: 8, avstar: 0, franvarande: 1 },
            { party: "S", partilinje: "Ja", ja: 8, nej: 0, avstar: 0, franvarande: 1 },
            { party: "C", partilinje: "Avstår", ja: 0, nej: 0, avstar: 8, franvarande: 1 },
          ],
        },
      ],
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

  test("shows acklamation as Avgjord utan omröstning and no party table", async ({ page }) => {
    await page.route("**/api/motions/**", (route) => route.fulfill({ json: MOTION_3293 }));
    await page.goto("/motioner/2017%2F18/3293");
    const decision = page.getByTestId("decision");
    await expect(decision).toContainText("betänkande 2017/18:SkU11");
    await expect(decision).toContainText("förslagspunkt 6");
    await expect(page.getByTestId("decision-mode")).toContainText("Avgjord utan omröstning");
    await expect(page.getByTestId("yrkande-outcome")).toContainText("Avslagen");
    await expect(page.getByTestId("party-table")).toHaveCount(0);
  });

  test("shows the Partilinje of every party for a votering", async ({ page }) => {
    await page.route("**/api/motions/**", (route) => route.fulfill({ json: MOTION_3638 }));
    await page.goto("/motioner/2021%2F22/3638");
    await expect(page.getByTestId("decision-mode")).toContainText("reservation 12 (M)");
    const table = page.getByTestId("party-table");
    await expect(table).toBeVisible();
    await expect(table).toContainText("Partilinje");
    await expect(table.getByRole("row")).toHaveCount(4);
  });

  test("says so when the motion is not in the record", async ({ page }) => {
    await page.route("**/api/motions/**", (route) =>
      route.fulfill({ status: 404, json: { error: "motion not found" } }),
    );
    await page.goto("/motioner/2017%2F18/999999");
    await expect(page.getByText("finns inte i vårt register")).toBeVisible();
  });
});
