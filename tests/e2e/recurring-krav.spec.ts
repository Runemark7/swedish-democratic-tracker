import { test, expect } from "@playwright/test";

// Shape of GET /api/motions/{rm}/{bet} for M's virtual-goods demand, from the
// recorded seam fixtures (Riksdagen, 2026-11-11). Only the fields the page
// reads are filled in.
const STEPS = [
  { riksmote: "2017/18", beteckning: "3293", number: 1, outcome: "Avslagen" },
  { riksmote: "2018/19", beteckning: "2900", number: 26, outcome: "Avslagen" },
  { riksmote: "2019/20", beteckning: "2718", number: 26, outcome: "Avslagen" },
  { riksmote: "2021/22", beteckning: "3638", number: 9, outcome: "Avslagen" },
];

function motion(riksmote: string, beteckning: string, number: number) {
  return {
    outcome: { word: "Avslagen", counts: { Avslagen: 1 }, total: 1 },
    dokId: "X",
    riksmote,
    beteckning,
    title: "Skattemässig hantering av virtuella varor",
    subtyp: "Kommittémotion",
    kind: "Kommittémotion",
    kindExplanation: "",
    date: "2017-10-05",
    utskott: "SkU",
    utskottName: "Skatteutskottet",
    status: "Klar",
    behandlasI: "",
    sourceUrl: "https://data.riksdagen.se/dokument/H5023293.html",
    authors: [],
    yrkanden: [
      {
        number,
        wording: "Riksdagen ställer sig bakom det som anförs i motionen om att tillsätta en utredning.",
        utskottProposal: "Avslag",
        chamberDecision: "Avslag",
        behandlasI: "",
        behandlasIPunkt: "",
        outcome: "Avslagen",
        decisions: [],
        recurring: {
          rule: "identical_wording",
          ruleLabel: "Identisk ordalydelse",
          members: STEPS.map((s) => ({
            ...s,
            date: null,
            parties: ["M"],
            rule: "identical_wording",
            current: s.riksmote === riksmote && s.beteckning === beteckning,
            sourceUrl: "https://data.riksdagen.se/dokument/X.html",
          })),
        },
      },
    ],
  };
}

test.describe("Återkommande krav on the motion page", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test("3293 shows the history and each step links to its motion page", async ({ page }) => {
    await page.route("**/api/motions/**", async (route) => {
      const m = new URL(route.request().url()).pathname.match(/\/api\/motions\/([^/]+)\/([^/]+)$/);
      const rm = decodeURIComponent(m![1]);
      const step = STEPS.find((s) => s.riksmote === rm && s.beteckning === m![2]);
      if (!step) return route.fulfill({ status: 404, json: { error: "motion not found" } });
      await route.fulfill({ json: motion(step.riksmote, step.beteckning, step.number) });
    });

    await page.goto("/motioner/2017%2F18/3293");

    const block = page.getByTestId("recurring");
    await expect(block).toContainText("Återkommande krav");
    const steps = page.getByTestId("recurring-step");
    await expect(steps).toHaveCount(4);
    for (const [i, s] of STEPS.entries()) {
      await expect(steps.nth(i)).toContainText(`${s.riksmote}:${s.beteckning} yrkande ${s.number}`);
      await expect(steps.nth(i)).toContainText("M");
      await expect(steps.nth(i)).toContainText("Avslagen");
      await expect(steps.nth(i)).toContainText("Identisk ordalydelse");
    }
    // The current motion is not a link; the others are.
    await expect(steps.nth(0).getByRole("link")).toHaveCount(0);
    for (const i of [1, 2, 3]) {
      const s = STEPS[i];
      await expect(steps.nth(i).getByRole("link")).toHaveAttribute(
        "href",
        `/motioner/${encodeURIComponent(s.riksmote)}/${s.beteckning}`,
      );
    }

    await steps.nth(1).getByRole("link").click();
    await expect(page).toHaveURL(/\/motioner\/2018%2F19\/2900$/);
    await expect(page.getByText("Motion 2018/19:2900")).toBeVisible();
  });
});
