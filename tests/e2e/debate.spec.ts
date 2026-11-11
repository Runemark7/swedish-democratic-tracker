import { test, expect } from "@playwright/test";

// The debate on 2017/18:SkU11 (2018-03-22) as GET /api/betankanden/... and
// GET /api/motions/... serve it from the seam fixtures recorded 2026-11-11
// (apitest/fixtures/speech-debate-motion), cut to three speakers.
const DEBATE = {
  date: "2018-03-22",
  total: 7,
  speeches: [
    {
      speechId: 101, politicianId: "0117203219614", name: "Jörgen Warborn", party: "M", date: "2018-03-22",
      excerpt: "Fru talman! Vi debatterar i dag skatteutskottets betänkande om mervärdesskatt.", excerptAvailable: true,
      riksdagenUrl: "https://data.riksdagen.se/anforande/H50969-21/html", linkRule: "rel_dok_id",
    },
    {
      speechId: 102, politicianId: "0418929039326", name: "Patrik Lundqvist", party: "S", date: "2018-03-22",
      excerpt: "", excerptAvailable: false,
      riksdagenUrl: "https://data.riksdagen.se/anforande/H50969-22/html", linkRule: "avsnittsrubrik_datum",
    },
    ...[3, 4, 5, 6, 7].map((n) => ({
      speechId: 100 + n, politicianId: "0679228585116", name: "Per Åsling", party: "C", date: "2018-03-22",
      excerpt: `Anförande ${n}.`, excerptAvailable: true,
      riksdagenUrl: `https://data.riksdagen.se/anforande/H50969-2${n}/html`, linkRule: "rel_dok_id",
    })),
  ],
};

const DECISION = {
  betankandeRiksmote: "2017/18", betankandeBeteckning: "SkU11", betankandeTitle: "Mervärdesskatt",
  betankandeSourceUrl: "https://data.riksdagen.se/dokument/H501SkU11.html", punkt: "6", rubrik: "Övriga frågor",
  linkedBy: "forslag_text", beslutstyp: "acklamation", mode: "acklamation", modeLabel: "Avgjord utan omröstning",
  alternativ: "", outcome: "Avslagen", voteringStored: false, parties: [], reasoning: "", reasoningAvailable: false,
  reservationsAvailable: false, reservations: [], sarskildaYttranden: [], votedReservation: null, debate: DEBATE,
};

const MOTION = {
  outcome: { word: "Avslagen", counts: { Avslagen: 1 }, total: 1 }, dokId: "H5023293", riksmote: "2017/18", beteckning: "3293",
  title: "Skattemässig hantering av virtuella varor", subtyp: "Kommittémotion", kind: "Kommittémotion", kindExplanation: "",
  date: "2017-10-05", utskott: "SkU", utskottName: "Skatteutskottet", status: "Klar", behandlasI: "2017/18:SkU11",
  sourceUrl: "https://data.riksdagen.se/dokument/H5023293.html", authors: [],
  yrkanden: [{
    number: 1, wording: "Riksdagen tillkännager en utredning.", utskottProposal: "Avslag", chamberDecision: "Avslag",
    behandlasI: "2017/18:SkU11", behandlasIPunkt: "", outcome: "Avslagen", decisions: [DECISION], recurring: null,
  }],
};

const BETANKANDE = {
  riksmote: "2017/18", beteckning: "SkU11", title: "Mervärdesskatt", organ: "SkU", dokId: "H501SkU11",
  sourceUrl: "https://data.riksdagen.se/dokument/H501SkU11.html", held: true, status: "Webbpublicering",
  decided: true, decisionDate: "2018-03-28", decisionText: "", punkter: [], debate: DEBATE,
};

test.describe("debate on the betänkande and motion pages", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test("the betänkande page lists the speakers with party, excerpt and links", async ({ page }) => {
    await page.route("**/api/betankanden/**", (route) => route.fulfill({ json: BETANKANDE }));
    await page.goto("/beslut/2017/18/SkU11");

    const debate = page.getByTestId("debate");
    await expect(debate).toContainText("Debatten · 7 anföranden · 22 mars 2018");
    const rows = debate.getByTestId("debate-speech");
    await expect(rows).toHaveCount(5); // five shown, the rest behind a button
    await expect(rows.first()).toContainText("Jörgen Warborn");
    await expect(rows.first()).toContainText("M");
    await expect(rows.first()).toContainText("Vi debatterar i dag skatteutskottets betänkande");
    await expect(rows.first()).toContainText("kopplad via Riksdagens dokumentreferens");
    await expect(rows.first().getByRole("link", { name: "Jörgen Warborn" })).toHaveAttribute("href", "/politicians/0117203219614");
    await expect(rows.first().getByRole("link", { name: /Läs hela anförandet/ })).toHaveAttribute("href", "/anforanden/101");
    await expect(rows.first().getByRole("link", { name: /Hos Riksdagen/ })).toHaveAttribute("href", "https://data.riksdagen.se/anforande/H50969-21/html");
    // A speech whose text is not fetched yet says so and states the rule that linked it.
    await expect(rows.nth(1)).toContainText("Anförandetexten är inte hämtad än.");
    await expect(rows.nth(1)).toContainText("kopplad via debattens rubrik och datum");

    await debate.getByRole("button", { name: /Visa alla 7 anföranden/ }).click();
    await expect(rows).toHaveCount(7);
  });

  test("the full speech link opens the speech page", async ({ page }) => {
    await page.route("**/api/betankanden/**", (route) => route.fulfill({ json: BETANKANDE }));
    await page.route("**/api/speeches/**", (route) => route.fulfill({ status: 404, json: { error: "not found" } }));
    await page.goto("/beslut/2017/18/SkU11");
    await page.getByTestId("debate-speech").first().getByRole("link", { name: /Läs hela anförandet/ }).click();
    await expect(page).toHaveURL(/\/anforanden\/101$/);
  });

  test("the motion page shows the debate on the betänkande that decided the yrkande", async ({ page }) => {
    await page.route("**/api/motions/**", (route) => route.fulfill({ json: MOTION }));
    await page.goto("/motioner/2017%2F18/3293");

    const debate = page.getByTestId("decision").getByTestId("debate");
    await expect(debate).toContainText("Debatten · 7 anföranden");
    await expect(debate.getByTestId("debate-speech").first().getByRole("link", { name: "Jörgen Warborn" })).toHaveAttribute("href", "/politicians/0117203219614");
    await expect(debate.getByTestId("debate-speech").first().getByRole("link", { name: /Läs hela anförandet/ })).toHaveAttribute("href", "/anforanden/101");
    await expect(debate.getByTestId("debate-speech").first().getByRole("link", { name: /Hos Riksdagen/ })).toHaveAttribute("href", "https://data.riksdagen.se/anforande/H50969-21/html");
  });

  test("a betänkande with no linked speeches says so instead of guessing", async ({ page }) => {
    await page.route("**/api/betankanden/**", (route) =>
      route.fulfill({ json: { ...BETANKANDE, debate: { date: null, total: 0, speeches: [] } } }));
    await page.goto("/beslut/2017/18/SkU11");
    await expect(page.getByTestId("debate-empty")).toContainText("Inga anföranden är kopplade till betänkandet.");
  });
});
