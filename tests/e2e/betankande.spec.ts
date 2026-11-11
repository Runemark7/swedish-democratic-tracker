import { test, expect } from "@playwright/test";

// Shapes follow GET /api/betankanden/{rm}/{bet} as served by the seam tests
// (fixtures recorded from Riksdagen, 2026-11-11).
function punkt(over: Record<string, unknown> = {}) {
  return {
    number: "6", rubrik: "Övriga frågor", text: "Utskottet föreslår avslag.", beslutstyp: "röstning",
    mode: "votering", modeLabel: "Avgjord genom omröstning", alternativ: "Utskottets förslag mot reservation 12 (M)",
    vinnare: "utskottet", voteringId: "V1", voteringStored: true,
    parties: [{ party: "M", partilinje: "Nej", ja: 0, nej: 8, avstar: 0, franvarande: 1 }, { party: "S", partilinje: "Ja", ja: 8, nej: 0, avstar: 0, franvarande: 0 }],
    avvikare: [{ politicianId: "0999", name: "Anna Exempel", party: "S", voteResult: "Nej", partilinje: "Ja" }],
    reasoning: "Utskottet avstyrker.", reasoningAvailable: true, reservationsAvailable: true,
    reservations: [{ number: 12, parties: ["M"], signers: ["A (M)"], proposal: "", text: "Vi vill något annat." }],
    sarskildaYttranden: [], votedReservation: { number: 12, party: "M", parties: ["M"] },
    motions: [{ dokId: "H9023638", riksmote: "2021/22", beteckning: "3638", title: "Virtuella varor", subtyp: "Kommittémotion", parties: ["M"], yrkanden: [9] }],
    ...over,
  };
}

function page(rm: string, over: Record<string, unknown> = {}) {
  return {
    riksmote: rm, beteckning: "SkU16", title: "Mervärdesskatt", organ: "SkU", dokId: "H901SkU16",
    sourceUrl: "https://data.riksdagen.se/dokument/H901SkU16.html", held: true, status: "Webbpublicering",
    decided: true, decisionDate: "2022-04-20", decisionText: "", punkter: [punkt()], ...over,
  };
}

test.describe("canonical betänkande page", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test("2025/26:SkU16 loads by its slash-form URL and asks the API with the encoded riksmöte", async ({ page: pw }) => {
    let requested = "";
    await pw.route("**/api/betankanden/**", async (route) => {
      requested = new URL(route.request().url()).pathname;
      await route.fulfill({ json: page("2025/26", { decisionDate: "2026-03-25" }) });
    });
    await pw.route("**/api/speeches**", (route) => route.fulfill({ json: [] }));
    await pw.goto("/beslut/2025/26/SkU16");
    await expect(pw.getByRole("heading", { level: 1 })).toContainText("Mervärdesskatt");
    await expect(pw.getByText("Betänkande 2025/26:SkU16")).toBeVisible();
    expect(requested).toContain("/api/betankanden/2025%2F26/SkU16");
    await expect(pw.getByTestId("decision-status")).toContainText("25 mars 2026");
    await expect(pw.getByTestId("party-table")).toContainText("Partilinje");
    await expect(pw.getByTestId("avvikare").getByRole("link", { name: "Anna Exempel" })).toHaveAttribute("href", "/politicians/0999");
    await expect(pw.getByTestId("punkt-motions").getByRole("link", { name: "2021/22:3638" })).toHaveAttribute("href", "/motioner/2021%2F22/3638");
    await expect(pw.getByTestId("reservations")).toBeVisible();
    await expect(pw.getByTestId("reasoning")).toBeVisible();
  });

  test("a planned betänkande says Ej avgjort än with the planned date", async ({ page: pw }) => {
    await pw.route("**/api/betankanden/**", (route) =>
      route.fulfill({ json: page("2026/27", { beteckning: "SkU8", decided: false, decisionDate: "2026-11-25", decisionText: "Ej avgjort än", punkter: [] }) }));
    await pw.route("**/api/speeches**", (route) => route.fulfill({ json: [] }));
    await pw.goto("/beslut/2026/27/SkU8");
    await expect(pw.getByTestId("decision-status")).toContainText("Ej avgjort än");
    await expect(pw.getByTestId("decision-status")).toContainText("25 november 2026");
  });

  test("a vote detail URL redirects to the canonical page, anchored to the förslagspunkt", async ({ page: pw }) => {
    await pw.route("**/api/betankanden/resolve**", (route) =>
      route.fulfill({ json: { riksmote: "2021/22", beteckning: "SkU16", punkt: "6" } }));
    await pw.route("**/api/betankanden/2021**", (route) => route.fulfill({ json: page("2021/22") }));
    await pw.route("**/api/speeches**", (route) => route.fulfill({ json: [] }));
    await pw.goto("/votes/SkU16/6");
    await expect(pw).toHaveURL(/\/beslut\/2021\/22\/SkU16#punkt-6$/);
    await expect(pw.locator("#punkt-6")).toBeVisible();
  });

  test("the old betänkande URL without riksmöte redirects to the most recent riksmöte", async ({ page: pw }) => {
    let query = "";
    await pw.route("**/api/betankanden/resolve**", (route) => {
      query = new URL(route.request().url()).search;
      return route.fulfill({ json: { riksmote: "2025/26", beteckning: "SkU16", punkt: "" } });
    });
    await pw.route("**/api/betankanden/2025**", (route) => route.fulfill({ json: page("2025/26") }));
    await pw.route("**/api/speeches**", (route) => route.fulfill({ json: [] }));
    await pw.goto("/beslut/SkU16");
    await expect(pw).toHaveURL(/\/beslut\/2025\/26\/SkU16$/);
    expect(query).not.toContain("rm=");
    await expect(pw.getByText("Betänkande 2025/26:SkU16")).toBeVisible();
  });

  test("the vote share URL still opens the share page", async ({ page: pw }) => {
    await pw.goto("/votes/SkU16/6/dela");
    await expect(pw).toHaveURL(/\/votes\/SkU16\/6\/dela$/);
  });
});
