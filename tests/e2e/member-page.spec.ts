import { test, expect } from "@playwright/test";

// The member page's backend is mocked: this checks that each section links to
// its target page and that no Löften tab is shown.
const ID = "0572788110724";

test.describe("Member page", () => {
  test.beforeEach(async ({ page }) => {
    await page.route(`**/api/politicians/${ID}`, (r) =>
      r.fulfill({ json: { id: ID, firstName: "Maria", lastName: "Malmer Stenergard", party: "M", constituency: "Södermanlands län", imageUrl: "", isActive: true } }),
    );
    await page.route(`**/api/politicians/${ID}/votes**`, (r) =>
      r.fulfill({
        json: {
          data: [{ voteringId: "V1", politicianId: ID, party: "M", voteResult: "Nej", beteckning: "SkU16", forslagspunkt: "6", session: "2021/22", proposalOrigin: {}, createdAt: "2022-01-01T00:00:00Z", avvikare: true, partilinje: "Ja" }],
          total: 1, page: 1, pageSize: 100,
        },
      }),
    );
    await page.route(`**/api/motions?**`, (r) =>
      r.fulfill({
        json: { data: [{ dokId: "H5023293", riksmote: "2017/18", beteckning: "3293", title: "Moms på virtuella varor", subtyp: "kommittemotion", kind: "Kommittémotion", parties: ["M"], date: "2017-10-01", utskott: "SkU", utskottName: "Skatteutskottet", handledIn: ["2017/18:SkU11"], sourceUrl: "x", outcome: { word: "Avslagen", counts: {}, total: 1 } }], nextCursor: null },
      }),
    );
    await page.route(`**/api/politicians/${ID}/questions`, (r) =>
      r.fulfill({ json: { questions: [{ dokId: "HA11339", kind: "fraga", riksmote: "2022/23", beteckning: "339", title: "Moms", answerDate: "2023-02-22", answerDokId: "HA12339", asker: { id: "1", name: "A", party: "C" }, answerer: { id: ID, name: "B", party: "M" }, addressed: { id: ID, name: "B", party: "M" }, role: "answerer", rule: "", sourceUrl: "x" }] } }),
    );
    await page.route(`**/api/speeches**`, (r) => r.fulfill({ json: [] }));
    await page.goto(`/politicians/${ID}`);
  });

  test("has no Löften tab", async ({ page }) => {
    await expect(page.getByRole("button", { name: "Motioner" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Löften" })).toHaveCount(0);
  });

  test("the vote row marks Avvikare and links to the förslagspunkt", async ({ page }) => {
    await expect(page.getByTestId("avvikare-mark")).toBeVisible();
    await expect(page.getByTestId("member-vote-row")).toHaveAttribute("href", /\/votes\/SkU16\/6\?session=2021%2F22/);
  });

  test("the motions tab links to the motion page", async ({ page }) => {
    await page.getByRole("button", { name: "Motioner" }).click();
    await expect(page.getByTestId("member-motions")).toContainText("Avslagen");
    await expect(page.getByRole("link", { name: /3293/ })).toHaveAttribute("href", "/motioner/2017%2F18/3293");
  });

  test("the frågor tab links to the question", async ({ page }) => {
    await page.getByRole("button", { name: "Frågor och debatter" }).click();
    await expect(page.getByTestId("questions-section").getByRole("link").first()).toHaveAttribute("href", "x");
  });
});
