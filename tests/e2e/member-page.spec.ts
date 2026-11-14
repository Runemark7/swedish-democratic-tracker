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

  // Gap member-page-no-individual-votes (#350): dated speeches used to sort
  // ahead of the undated votes and fill the whole 50-row feed, so a member who
  // had spoken 50 times showed no vote at all.
  test.describe("a member with many speeches and many votes", () => {
    const BUSY = "0203877548421";
    const TOTAL = 120;
    const vote = (n: number) => ({
      voteringId: `V${n}`, politicianId: BUSY, party: "S", voteResult: n === 1 ? "Frånvarande" : "Ja",
      beteckning: `KU${n}`, forslagspunkt: "1", session: "2026/27", dokId: `HE01KU${n}`,
      proposalOrigin: { documentTitle: `Betänkande nummer ${n}` }, createdAt: "2026-11-11T00:00:00Z", partilinje: "Ja",
    });
    const speech = (n: number) => ({
      id: n, dokId: `HE09${n}`, politicianId: BUSY, politicianName: "Malin Larsson", party: "S",
      date: "2026-10-01T00:00:00Z", topicHeading: `Anförande ${n}`, snippet: `Anförande ${n}`,
    });
    const pagesAsked: string[] = [];

    test.beforeEach(async ({ page }) => {
      pagesAsked.length = 0;
      await page.route(`**/api/politicians/${BUSY}`, (r) =>
        r.fulfill({ json: { id: BUSY, firstName: "Malin", lastName: "Larsson", party: "S", constituency: "Västernorrlands län", imageUrl: "", isActive: true } }),
      );
      await page.route(`**/api/politicians/${BUSY}/votes**`, (r) => {
        const q = new URL(r.request().url()).searchParams;
        const pageNo = Number(q.get("page") ?? "1");
        const size = Number(q.get("pageSize") ?? "50");
        pagesAsked.push(String(pageNo));
        const from = (pageNo - 1) * size;
        const data = Array.from({ length: Math.max(0, Math.min(size, TOTAL - from)) }, (_, i) => vote(from + i + 1));
        return r.fulfill({
          json: { data, total: TOTAL, page: pageNo, pageSize: size, summary: { ja: 119, nej: 0, avstar: 0, franvarande: 1, total: TOTAL } },
        });
      });
      await page.route(`**/api/speeches**`, (r) =>
        r.fulfill({ json: Array.from({ length: 60 }, (_, i) => speech(i + 1)) }),
      );
      await page.goto(`/politicians/${BUSY}`);
    });

    test("lists individual votes with result, linked to the decision", async ({ page }) => {
      const votes = page.getByTestId("member-votes");
      const rows = votes.getByTestId("member-vote-row");
      await expect(rows).toHaveCount(50);
      await expect(rows.first()).toContainText("Betänkande nummer 1");
      await expect(rows.first()).toContainText("KU1 punkt 1");
      await expect(rows.first()).toContainText("2026/27");
      await expect(rows.first()).toContainText("Frånvarande");
      await expect(rows.first()).toHaveAttribute("href", /\/votes\/KU1\/1\?session=2026%2F27/);
      await expect(votes).toContainText("Visar 50 av 120");
      // The speeches are still there, in their own list.
      await expect(page.getByTestId("member-speeches")).toContainText("Anförande 1");
      await expect(page.getByTestId("member-speeches").getByTestId("member-vote-row")).toHaveCount(0);
    });

    test("Visa fler loads the next page until every vote is listed", async ({ page }) => {
      const votes = page.getByTestId("member-votes");
      const more = votes.getByRole("button", { name: "Visa fler röster" });
      await more.click();
      await expect(votes.getByTestId("member-vote-row")).toHaveCount(100);
      await more.click();
      await expect(votes.getByTestId("member-vote-row")).toHaveCount(120);
      await expect(votes).toContainText("Visar 120 av 120");
      await expect(more).toHaveCount(0);
      expect(pagesAsked).toEqual(["1", "2", "3"]);
    });
  });
});
