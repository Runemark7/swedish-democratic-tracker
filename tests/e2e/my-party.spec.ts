import { test, expect } from "@playwright/test";

const PARTIES = ["S", "M", "SD", "C", "V", "KD", "MP", "L"];
const motion = (n: string, parties: string[]) => ({
  dokId: `H${n}`, riksmote: "2017/18", beteckning: n, title: `Motion ${n}`, subtyp: "kommittémotion",
  kind: "Kommittémotion", parties, date: "2017-10-04", utskott: "SkU", utskottName: "Skatteutskottet",
  handledIn: ["2017/18:SkU11"], sourceUrl: "https://data.riksdagen.se/dokument/x.html",
  outcome: { word: "Avslagen", counts: { Avslagen: 1 }, total: 1 }, decidedAt: "2018-03-28",
});
const FEED = {
  decided: [{
    riksmote: "2025/26", beteckning: "UbU30", title: "Friskolor", utskott: "UbU", utskottName: "Utbildningsutskottet",
    date: "2026-08-13", decided: true, status: "x", outcome: "Avgjord genom omröstning", sourceUrl: "x",
    parties: PARTIES.map((party) => ({ party, partilinje: "Ja", voteringar: 1 })),
  }],
  motions: [motion("1", ["S"]), motion("2", ["M"])],
  upcoming: [],
  coverage: { oldest: "2026-08-13", newest: "2026-08-13", stale: false },
  partyTiles: PARTIES.slice().sort().map((party) => ({ party, total: 1, for: 0, emot: 0, blandat: 0, inteProvat: 1, arkiv: 0 })),
};
const LIST = { data: [motion("1", ["S"]), motion("2", ["M"])].map((m) => ({ ...m })), nextCursor: null };

test.describe("Mitt parti", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test("persists across reloads, highlights without reordering, clears, and is never sent", async ({ page }) => {
    const requests: string[] = [];
    page.on("request", (r) => requests.push(`${r.url()} ${r.postData() ?? ""} ${JSON.stringify(r.headers())}`));
    await page.route("**/api/frontpage", (route) => route.fulfill({ json: FEED }));
    await page.route("**/api/motions?**", (route) => route.fulfill({ json: LIST }));
    await page.route("**/api/motions", (route) => route.fulfill({ json: LIST }));
    await page.route("**/api/parties", (route) => route.fulfill({ json: [] }));

    await page.goto("/");
    await expect(page.locator("[data-my-party=true]")).toHaveCount(0);
    await page.getByLabel("Mitt parti").selectOption("M");

    const motions = page.getByRole("region", { name: "Senaste motionsutfall" });
    await expect(motions.locator('[data-my-party="true"]')).toHaveCount(1);
    await expect(motions.getByTestId("front-motion").nth(1)).toHaveAttribute("data-my-party", "true");
    await expect(page.getByRole("region", { name: "Senaste besluten" }).locator('[data-party="M"]')).toHaveAttribute("data-my-party", "true");

    await page.reload();
    await expect(page.getByLabel("Mitt parti")).toHaveValue("M");
    await expect(motions.getByTestId("front-motion").nth(1)).toHaveAttribute("data-my-party", "true");
    // Ordering unchanged.
    await expect(motions.getByTestId("front-motion").first()).toContainText("Motion 1");

    await page.goto("/motioner");
    const rows = page.getByTestId("motion-row");
    await expect(rows).toHaveCount(2);
    await expect(rows.first()).toContainText("Motion 1");
    await expect(rows.nth(1).locator("xpath=..")).toHaveAttribute("data-my-party", "true");
    await expect(rows.first().locator("xpath=..")).not.toHaveAttribute("data-my-party", "true");

    await page.getByRole("button", { name: "Rensa" }).click();
    await expect(page.locator("[data-my-party=true]")).toHaveCount(0);
    await page.reload();
    await expect(page.getByLabel("Mitt parti")).toHaveValue("");

    // Select again, browse, and verify nothing carried the choice.
    await page.getByLabel("Mitt parti").selectOption("M");
    await page.goto("/");
    expect(await page.evaluate(() => localStorage.getItem("sdt-my-party"))).toBe("M");
    for (const r of requests.filter((r) => r.includes("/api/"))) {
      expect(r).not.toMatch(/sdt-my-party/);
      expect(r).not.toMatch(/mitt|my-?party/i);
    }
  });

  test("Om sajten says the choice stays in the browser", async ({ page }) => {
    await page.goto("/om-sajten");
    await expect(page.getByTestId("my-party-note")).toContainText("bara i din webbläsare");
  });
});
