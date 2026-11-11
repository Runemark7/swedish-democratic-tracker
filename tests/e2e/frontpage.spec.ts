import { test, expect } from "@playwright/test";

// Shape of GET /api/frontpage, as the seam test in backend/internal/apitest
// (frontpage_test.go) asserts it.
const PARTIES = ["S", "M", "SD", "C", "V", "KD", "MP", "L"];
const LINE: Record<string, string> = { C: "Ja", KD: "Ja", L: "Ja", M: "Ja", SD: "Ja", S: "Nej", MP: "Nej", V: "Nej" };

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
  loefteCount: 2,
};
const UPCOMING = {
  riksmote: "2026/27",
  beteckning: "SkU8",
  title: "Nya förenklingsregler",
  utskott: "SkU",
  utskottName: "Skatteutskottet",
  date: "2026-11-25",
  decided: false,
  status: "planerat",
  outcome: "Ej avgjord än",
  sourceUrl: "https://data.riksdagen.se/dokument/HE01SkU8.html",
  parties: [],
  loefteCount: 0,
};
const MOTION = {
  dokId: "H5023293",
  riksmote: "2017/18",
  beteckning: "3293",
  title: "Skattemässig hantering av virtuella varor",
  subtyp: "kommittémotion",
  kind: "Kommittémotion",
  parties: ["M"],
  date: "2017-10-04",
  utskott: "SkU",
  utskottName: "Skatteutskottet",
  handledIn: ["2017/18:SkU11"],
  sourceUrl: "https://data.riksdagen.se/dokument/H5023293.html",
  outcome: { word: "Avslagen", counts: { Avslagen: 1 }, total: 1 },
  decidedAt: "2018-03-28",
};

// Alphabetical by party code, as the API returns them. M is the only party with
// a Löfte of its own; the counts must not influence the order.
const TILES = ["C", "KD", "L", "M", "MP", "S", "SD", "V"].map((party) => ({
  party, total: party === "M" ? 3 : 0, for: party === "M" ? 2 : 0, emot: 0, blandat: 0, inteProvat: party === "M" ? 1 : 0, arkiv: 0,
}));

function feed(stale: boolean) {
  return {
    decided: [DECIDED],
    motions: [MOTION],
    upcoming: [UPCOMING],
    coverage: { oldest: "2026-08-13", newest: "2026-08-13", stale },
    partyTiles: TILES,
  };
}

test.describe("Front page", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test("asks the question", async ({ page }) => {
    await page.route("**/api/frontpage", (route) => route.fulfill({ json: feed(false) }));
    await page.goto("/");
    await expect(page.getByRole("heading", { level: 1 })).toHaveText("Vad beslutar riksdagen — och håller de vad de lovar?");
  });

  test("shows eight party tiles in alphabetical order, each linking to the party's Löften", async ({ page }) => {
    await page.route("**/api/frontpage", (route) => route.fulfill({ json: feed(false) }));
    await page.goto("/");
    const tiles = page.getByTestId("party-tile");
    await expect(tiles).toHaveCount(8);
    expect(await tiles.evaluateAll((els) => els.map((e) => e.getAttribute("data-party")))).toEqual(
      ["C", "KD", "L", "M", "MP", "S", "SD", "V"],
    );
    await expect(page.locator('[data-party="M"][data-testid="party-tile"]')).toContainText("Röstat för 2");
    await expect(page.locator('[data-party="C"][data-testid="party-tile"]')).toContainText("Inga löften publicerade ännu");

    await page.route("**/api/loften?**", (r) => r.fulfill({ json: [] }));
    await page.route("**/api/loften/summary", (r) => r.fulfill({ json: [] }));
    await page.route("**/api/valmanifest", (r) => r.fulfill({ json: [] }));
    await page.locator('[data-party="MP"][data-testid="party-tile"]').click();
    await expect(page).toHaveURL(/\/loften\?party=MP$/);
  });

  test("Mitt parti is highlighted on its tile but the tiles keep their order", async ({ page }) => {
    await page.route("**/api/frontpage", (route) => route.fulfill({ json: feed(false) }));
    await page.addInitScript(() => localStorage.setItem("sdt-my-party", "V"));
    await page.goto("/");
    const tiles = page.getByTestId("party-tile");
    await expect(page.locator('[data-party="V"][data-testid="party-tile"]')).toHaveAttribute("data-my-party", "true");
    expect(await tiles.evaluateAll((els) => els.map((e) => e.getAttribute("data-party")))).toEqual(
      ["C", "KD", "L", "M", "MP", "S", "SD", "V"],
    );
  });

  test("'berör 2 löften' on a decided row links to exactly those Löften; upcoming rows have none", async ({ page }) => {
    await page.route("**/api/frontpage", (route) => route.fulfill({ json: feed(false) }));
    await page.goto("/");
    const link = page.getByRole("region", { name: "Senaste besluten" }).getByTestId("beror-loften");
    await expect(link).toHaveText("berör 2 löften");
    await expect(page.getByRole("region", { name: "Kommande beslut" }).getByTestId("beror-loften")).toHaveCount(0);

    let asked = "";
    await page.route("**/api/loften?**", (r) => { asked = r.request().url(); return r.fulfill({ json: [] }); });
    await page.route("**/api/loften/summary", (r) => r.fulfill({ json: [] }));
    await page.route("**/api/valmanifest", (r) => r.fulfill({ json: [] }));
    await link.click();
    await expect(page).toHaveURL(/\/loften\?betankande=2025%2F26%3AUbU30$/);
    await expect(page.getByTestId("betankande-filter")).toContainText("2025/26:UbU30");
    expect(decodeURIComponent(asked)).toContain("betankande=2025/26:UbU30");
  });

  test("keeps decided, motion outcomes and upcoming apart, and every row, tag and party mark is a link", async ({ page }) => {
    await page.route("**/api/frontpage", (route) => route.fulfill({ json: feed(false) }));
    await page.goto("/");

    const decided = page.getByRole("region", { name: "Senaste besluten" });
    const motions = page.getByRole("region", { name: "Senaste motionsutfall" });
    const upcoming = page.getByRole("region", { name: "Kommande beslut" });
    await expect(decided.getByTestId("front-decided")).toHaveCount(1);
    await expect(decided.getByTestId("front-upcoming")).toHaveCount(0);
    await expect(upcoming.getByTestId("front-upcoming")).toHaveCount(1);
    await expect(upcoming.getByTestId("party-mark")).toHaveCount(0);
    await expect(motions.getByTestId("front-motion")).toHaveCount(1);

    await expect(decided.getByTestId("row-date")).toContainText("Beslutad 13 augusti 2026");
    await expect(upcoming.getByTestId("row-date")).toContainText("Planerad beslutsdag 25 november 2026");
    await expect(decided.getByTestId("party-mark")).toHaveCount(8);
    await expect(decided.locator('[data-party="S"]')).toHaveAttribute("href", "/parties/S");
    await expect(decided.locator('[data-party="S"]')).toContainText("Nej");

    await expect(decided.getByRole("link", { name: /Skärpta villkor/ })).toHaveAttribute("href", "/beslut/2025/26/UbU30");
    await expect(decided.getByTestId("utskott-tag")).toHaveAttribute("href", "/beslut?utskott=UbU");
    await expect(decided.getByTestId("row-outcome")).toHaveAttribute("href", "/beslut/2025/26/UbU30");
    await expect(upcoming.getByRole("link", { name: /Nya förenklingsregler/ })).toHaveAttribute("href", "/beslut/2026/27/SkU8");

    await expect(motions.getByRole("link", { name: /virtuella varor/ })).toHaveAttribute("href", "/motioner/2017%2F18/3293");
    await expect(motions.getByTestId("motion-party")).toHaveAttribute("href", "/parties/M");
    await expect(motions.getByTestId("motion-kind")).toHaveAttribute("href", /\/motioner\?kind=Kommit/);
    await expect(motions.getByTestId("motion-outcome")).toHaveAttribute("href", "/motioner?outcome=Avslagen");
    await expect(motions.getByTestId("motion-handled-in")).toHaveAttribute("href", "/beslut/2017/18/SkU11");

    // No <a> without a target anywhere in the feed.
    for (const a of await page.locator("main a, [class*=space-y-8] a").all()) {
      await expect(a).toHaveAttribute("href", /.+/);
    }
    await expect(page.getByRole("link", { name: "Alla beslut" })).toHaveAttribute("href", "/beslut?state=decided");
    await expect(page.getByRole("link", { name: "Alla motioner" })).toHaveAttribute("href", "/motioner");
  });

  test("states which dates the decisions cover only when the newest is older than this week", async ({ page }) => {
    await page.route("**/api/frontpage", (route) => route.fulfill({ json: feed(true) }));
    await page.goto("/");
    await expect(page.getByTestId("coverage")).toContainText("13 augusti 2026 – 13 augusti 2026");
  });

  test("shows no coverage sentence when decisions are from this week", async ({ page }) => {
    await page.route("**/api/frontpage", (route) => route.fulfill({ json: feed(false) }));
    await page.goto("/");
    await expect(page.getByTestId("coverage")).toHaveCount(0);
  });

  test("a row link navigates to the betänkande page", async ({ page }) => {
    await page.route("**/api/frontpage", (route) => route.fulfill({ json: feed(false) }));
    await page.route("**/api/betankanden/**", (route) => route.fulfill({ status: 404, json: { error: "x" } }));
    await page.goto("/");
    await page.getByRole("region", { name: "Senaste besluten" }).getByRole("link", { name: /Skärpta villkor/ }).click();
    await expect(page).toHaveURL(/\/beslut\/2025\/26\/UbU30$/);
  });
});
