import { test, expect, type Page } from "@playwright/test";

// Shape of GET /api/motions/2017%2F18/3293 (recorded fixtures, 2026-11-10), with the debate,
// the Återkommande krav through 2021/22:3638 y9 and a linked fråga added.
const DEBATE = {
  date: "2018-03-28",
  total: 1,
  speeches: [
    {
      speechId: 1, politicianId: "0572788110724", name: "Maria Malmer Stenergard", party: "M", date: "2018-03-28",
      excerpt: "Herr talman! Handeln med virtuella varor växer.", excerptAvailable: true,
      riksdagenUrl: "https://data.riksdagen.se/anforande/H501-1", linkRule: "rel_dok_id",
    },
  ],
};
const DECISION = {
  betankandeRiksmote: "2017/18", betankandeBeteckning: "SkU11", betankandeTitle: "Mervärdesskatt",
  betankandeSourceUrl: "https://data.riksdagen.se/dokument/H501SkU11.html",
  debate: DEBATE, punkt: "6", rubrik: "Övriga frågor", linkedBy: "forslag_text", beslutstyp: "acklamation",
  mode: "acklamation", modeLabel: "Avgjord utan omröstning", alternativ: "", outcome: "Avslagen",
  voteringStored: false, voteringId: "", voteringUrl: "", parties: [],
  reasoning: "Utskottet gör ingen annan bedömning om virtuella varor än Skatteverket.", reasoningAvailable: true,
  reservationsAvailable: true,
  reservations: [
    { number: 8, parties: ["M"], signers: ["Cecilia Widegren (M)"], proposal: "", text: "Vi föreslår en utredning." },
    { number: 9, parties: ["KD"], signers: ["Larry Söder (KD)"], proposal: "", text: "KD-text." },
  ],
  sarskildaYttranden: [], votedReservation: null,
};
const MOTION = {
  outcome: { word: "Avslagen", counts: { Avslagen: 1 }, total: 1 },
  dokId: "H5023293", riksmote: "2017/18", beteckning: "3293",
  title: "Skattemässig hantering av virtuella varor", subtyp: "Kommittémotion", kind: "Kommittémotion",
  kindExplanation: "Väckt i namn av ett partis ledamöter i ett utskott.", date: "2017-10-05",
  utskott: "SkU", utskottName: "Skatteutskottet", status: "Klar", behandlasI: "2017/18:SkU11",
  sourceUrl: "https://data.riksdagen.se/dokument/H5023293.html",
  authors: [{ politicianId: "0572788110724", name: "Maria Malmer Stenergard", party: "M", role: "undertecknare" }],
  yrkanden: [
    {
      number: 1, wording: "Riksdagen ställer sig bakom det som anförs i motionen om en utredning.",
      utskottProposal: "Avslag", chamberDecision: "Avslag", behandlasI: "2017/18:SkU11", behandlasIPunkt: "",
      outcome: "Avslagen", decisions: [DECISION],
      recurring: {
        rule: "identical_wording", ruleLabel: "Identisk ordalydelse",
        members: [
          { riksmote: "2017/18", beteckning: "3293", number: 1, parties: ["M"], outcome: "Avslagen", rule: "identical_wording", current: true },
          { riksmote: "2021/22", beteckning: "3638", number: 9, parties: ["M"], outcome: "Avslagen", rule: "identical_wording", current: false },
        ],
      },
    },
  ],
};
const VOTERING_DECISION = {
  ...DECISION, betankandeRiksmote: "2021/22", betankandeBeteckning: "SkU16",
  betankandeSourceUrl: "https://data.riksdagen.se/dokument/H901SkU16.html", beslutstyp: "röstning",
  mode: "votering", modeLabel: "Avgjord genom omröstning", alternativ: "Utskottets förslag mot reservation 12 (M)",
  voteringStored: true, voteringId: "ABC-1", voteringUrl: "https://data.riksdagen.se/votering/ABC-1",
  parties: [{ party: "M", partilinje: "Nej", ja: 0, nej: 8, avstar: 0, franvarande: 1 }],
};
const QUESTION = {
  dokId: "H8339", kind: "fraga", riksmote: "2022/23", beteckning: "339", title: "Moms på digitala varor",
  sourceUrl: "https://data.riksdagen.se/dokument/H8339.html", rule: "cited-beteckning",
  asker: { name: "Rickard Nordin", party: "C" }, addressed: { name: "Elisabeth Svantesson" },
  answerer: { name: "Elisabeth Svantesson" }, answerDate: "2023-02-22",
};

async function setup(page: Page, theme: string, motion: unknown) {
  await page.addInitScript((v) => localStorage.setItem("sdt-theme", v), theme);
  await page.route("**/api/motions/**", (route) =>
    route.fulfill({ json: route.request().url().endsWith("/questions") ? { questions: [QUESTION] } : motion }),
  );
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`complete motion page (${theme})`, () => {
    test.use({ viewport: { width: 1280, height: 800 } });

    test("every section of the chain renders and links out", async ({ page }) => {
      await setup(page, theme, MOTION);
      await page.goto("/motioner/2017%2F18/3293");

      const steps = page.getByTestId("timeline-step");
      await expect(steps).toHaveCount(5);
      await expect(steps.nth(0).getByRole("link")).toHaveAttribute("href", MOTION.sourceUrl);
      await expect(steps.nth(1).getByRole("link")).toHaveAttribute("href", "/beslut/2017/18/SkU11#punkt-6");
      await expect(steps.nth(2)).toContainText("1 anföranden");
      await expect(steps.nth(3)).toContainText("Avslagen");
      await expect(steps.nth(3)).toContainText("Avgjord utan omröstning");

      await expect(page.getByRole("link", { name: "Maria Malmer Stenergard" }).first()).toBeVisible();
      await expect(page.getByTestId("decision")).toContainText("förslagspunkt 6");
      await expect(page.getByTestId("decision").getByRole("link", { name: /betänkande 2017\/18:SkU11/ })).toHaveAttribute("href", DECISION.betankandeSourceUrl);
      await expect(page.getByTestId("decision-links").getByRole("link", { name: /Betänkandet på Riksdagskollen/ })).toHaveAttribute("href", "/beslut/2017/18/SkU11#punkt-6");
      await expect(page.getByTestId("reasoning")).toContainText("Skatteverket");
      await expect(page.getByTestId("reservation")).toHaveCount(2);
      await expect(page.getByTestId("debate-speech").getByRole("link", { name: /Hos Riksdagen/ })).toHaveAttribute("href", DEBATE.speeches[0].riksdagenUrl);
      await expect(page.getByTestId("recurring-step").nth(1).getByRole("link")).toHaveAttribute("href", "/motioner/2021%2F22/3638");
      await expect(page.getByTestId("question-row").getByRole("link")).toHaveAttribute("href", QUESTION.sourceUrl);

      // Each term is explained on first use.
      const explain = page.getByTestId("explain");
      for (const word of ["Ett yrkande", "Ett betänkande", "En förslagspunkt", "Acklamation", "En reservation", "Återkommande krav"]) {
        await expect(explain.filter({ hasText: word }).first()).toBeVisible();
      }
    });

    test("a votering links to Riksdagen's votering", async ({ page }) => {
      const motion = { ...MOTION, yrkanden: [{ ...MOTION.yrkanden[0], decisions: [VOTERING_DECISION] }] };
      await setup(page, theme, motion);
      await page.goto("/motioner/2021%2F22/3638");
      await expect(page.getByTestId("decision-links").getByRole("link", { name: /Omröstningen hos Riksdagen/ })).toHaveAttribute("href", VOTERING_DECISION.voteringUrl);
      await expect(page.getByTestId("party-table")).toBeVisible();
    });

    test("an undecided motion says Ej avgjord än", async ({ page }) => {
      const motion = {
        ...MOTION, outcome: { word: "Ej avgjord än", counts: { "Ej avgjord än": 1 }, total: 1 },
        yrkanden: [{ ...MOTION.yrkanden[0], outcome: "Ej avgjord än", chamberDecision: "", decisions: [], recurring: null }],
      };
      await setup(page, theme, motion);
      await page.goto("/motioner/2017%2F18/3293");
      await expect(page.getByTestId("timeline-step").nth(3)).toContainText("Ej avgjord än");
      await expect(page.getByTestId("timeline-step").nth(1)).toContainText("Behandlas i 2017/18:SkU11");
      await expect(page.getByTestId("motion-outcome")).toContainText("Ej avgjord än");
    });
  });
}
