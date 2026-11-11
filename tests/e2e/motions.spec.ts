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
  reasoning:
    "Skatteverket gör bedömningen att virtuella varor kan vara momspliktiga.\n\nUtskottet gör ingen annan bedömning om virtuella varor än Skatteverket.\n\nUtskottet avstyrker därmed samtliga motionsyrkanden.",
  reasoningAvailable: true,
  reservationsAvailable: true,
  reservations: [
    {
      number: 8,
      parties: ["M"],
      signers: ["Cecilia Widegren (M)", "Helena Bouveng (M)", "Fredrik Schulte (M)", "Lotta Finstorp (M)"],
      proposal: "Riksdagen ställer sig bakom det som anförs i reservationen och tillkännager detta för regeringen.",
      text: "Vi föreslår en utredning för att se över den skatterättsliga bedömningen av handel med virtuella varor.",
    },
    {
      number: 9,
      parties: ["KD"],
      signers: ["Larry Söder (KD)"],
      proposal: "Riksdagen ställer sig bakom det som anförs i reservationen och tillkännager detta för regeringen.",
      text: "Privatpersoner som egenproducerar förnybar el har i dag möjlighet att sälja den el som blir över.",
    },
  ],
  sarskildaYttranden: [],
  votedReservation: null,
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

  test("shows the reservations and the utskott's reasoning on the 3293 page", async ({ page }) => {
    await page.route("**/api/motions/**", (route) => route.fulfill({ json: MOTION_3293 }));
    await page.goto("/motioner/2017%2F18/3293");
    const reasoning = page.getByTestId("reasoning");
    await expect(reasoning).toContainText("Utskottets ställningstagande");
    await expect(reasoning).toContainText("gör ingen annan bedömning om virtuella varor än Skatteverket");
    const reservations = page.getByTestId("reservation");
    await expect(reservations).toHaveCount(2);
    await expect(reservations.nth(0)).toContainText("Reservation 8 (M)");
    await expect(reservations.nth(1)).toContainText("Reservation 9 (KD)");
    await expect(page.getByTestId("voted-alternative")).toHaveCount(0);
  });

  test("names the reservation that was put to the vote", async ({ page }) => {
    const json = {
      ...MOTION_3638,
      yrkanden: [
        {
          ...MOTION_3638.yrkanden[0],
          decisions: [
            {
              ...MOTION_3638.yrkanden[0].decisions[0],
              votedReservation: { number: 2, party: "SD", parties: ["SD"] },
              reservations: [DECISION_3293.reservations[0]],
            },
          ],
        },
      ],
    };
    await page.route("**/api/motions/**", (route) => route.fulfill({ json }));
    await page.goto("/motioner/2021%2F22/3638");
    await expect(page.getByTestId("voted-alternative")).toContainText("reservation 2 (SD)");
    await expect(page.getByTestId("voted-alternative")).toContainText("Övriga reservationer");
  });

  test("says not available from Riksdagen, with the source link, instead of inventing text", async ({ page }) => {
    const json = {
      ...MOTION_3293,
      yrkanden: [
        {
          ...MOTION_3293.yrkanden[0],
          decisions: [
            {
              ...DECISION_3293,
              reasoning: "",
              reasoningAvailable: false,
              reservationsAvailable: false,
              reservations: [],
            },
          ],
        },
      ],
    };
    await page.route("**/api/motions/**", (route) => route.fulfill({ json }));
    await page.goto("/motioner/2017%2F18/3293");
    await expect(page.getByTestId("reasoning")).toHaveCount(0);
    await expect(page.getByTestId("reservations")).toHaveCount(0);
    const na = page.getByTestId("not-available");
    await expect(na).toHaveCount(2);
    await expect(na.first()).toContainText("inte tillgänglig från Riksdagen");
    await expect(na.first().getByRole("link")).toHaveAttribute("href", "https://data.riksdagen.se/dokument/H501SkU11.html");
  });
});

// Shape of GET /api/motions, as the seam test reads it from the recorded fixtures.
const ROW_3293 = {
  dokId: "H5023293",
  riksmote: "2017/18",
  beteckning: "3293",
  title: "Skattemässig hantering av virtuella varor",
  subtyp: "Kommittémotion",
  kind: "Kommittémotion",
  parties: ["M"],
  date: "2017-10-05",
  utskott: "SkU",
  utskottName: "Skatteutskottet",
  handledIn: ["2017/18:SkU11"],
  sourceUrl: "https://data.riksdagen.se/dokument/H5023293.html",
  outcome: { word: "Avslagen", counts: { Avslagen: 1 }, total: 1 },
};
const ROW_4202 = {
  ...ROW_3293,
  dokId: "H5024202",
  beteckning: "4202",
  title: "En motion av tre partier",
  subtyp: "Enskild motion",
  kind: "Flerpartimotion",
  parties: ["C", "KD", "M"],
  outcome: { word: "Delvis bifallen", counts: { Bifallen: 2, Avslagen: 1 }, total: 3 },
};

test.describe("Motioner list", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test("menu has five items and Motioner leads to the list", async ({ page }) => {
    await page.route("**/api/motions?**", (route) => route.fulfill({ json: { data: [ROW_3293], nextCursor: null } }));
    await page.goto("/");
    const primary = page.getByRole("navigation", { name: "Huvudmeny" });
    await expect(primary.getByRole("link")).toHaveText(["Beslut", "Motioner", "Ledamöter", "Partier", "Regeringen"]);
    await primary.getByRole("link", { name: "Motioner", exact: true }).click();
    await expect(page).toHaveURL(/\/motioner$/);
    await expect(page).toHaveTitle("Riksdagskollen – Motioner");
    await expect(primary.getByRole("link", { name: "Motioner", exact: true })).toHaveAttribute("aria-current", "page");
  });

  test("a row shows title, parties, kind, date and outcome and links to its motion page", async ({ page }) => {
    await page.route("**/api/motions?**", (route) =>
      route.fulfill({ json: { data: [ROW_3293, ROW_4202], nextCursor: null } }),
    );
    await page.route("**/api/motions/2017%2F18/3293", (route) => route.fulfill({ json: MOTION_3293 }));
    await page.goto("/motioner");

    const rows = page.getByTestId("motion-row");
    await expect(rows).toHaveCount(2);
    await expect(rows.first()).toContainText("Skattemässig hantering av virtuella varor");
    await expect(rows.first()).toContainText("2017/18:3293");
    await expect(rows.first().getByTestId("row-kind")).toHaveText("Kommittémotion");
    await expect(rows.first()).toContainText("5 oktober 2017");
    await expect(rows.first().getByTestId("row-outcome")).toHaveText("Avslagen");
    await expect(rows.nth(1).getByTestId("row-kind")).toHaveText("Flerpartimotion");
    await expect(rows.nth(1).getByTestId("row-outcome")).toHaveText(
      "Delvis bifallen (2 bifallen, 1 avslagen av 3 yrkanden)",
    );

    await rows.first().click();
    await expect(page).toHaveURL(/\/motioner\/2017%2F18\/3293$/);
    await expect(page.getByRole("heading", { level: 1 })).toContainText("Skattemässig hantering av virtuella varor");
  });

  test("filters and search are sent to the API and kept in the URL", async ({ page }) => {
    const seen: URLSearchParams[] = [];
    await page.route("**/api/motions?**", async (route) => {
      seen.push(new URL(route.request().url()).searchParams);
      await route.fulfill({ json: { data: [ROW_3293], nextCursor: null } });
    });
    await page.goto("/motioner");
    const filters = page.getByTestId("motion-filters");
    await filters.getByLabel("Parti", { exact: true }).selectOption("M");
    await filters.getByLabel("Utskott", { exact: true }).selectOption("SkU");
    await filters.getByLabel("Utfall", { exact: true }).selectOption("Avslagen");
    await filters.getByLabel("Riksmöte", { exact: true }).selectOption("2017/18");
    await filters.getByLabel("Typ", { exact: true }).selectOption("Kommittémotion");
    await page.getByRole("searchbox", { name: "Sök i titlar och yrkanden" }).fill("virtuella varor");
    await page.getByRole("button", { name: "Sök" }).click();

    await expect.poll(() => seen.at(-1)?.get("q")).toBe("virtuella varor");
    const last = seen.at(-1)!;
    expect(Object.fromEntries(last.entries())).toMatchObject({
      party: "M", utskott: "SkU", outcome: "Avslagen", riksmote: "2017/18", kind: "Kommittémotion", q: "virtuella varor",
    });
    await expect(page).toHaveURL(/party=M/);
    await expect(page).toHaveURL(/kind=Kommitt%C3%A9motion/);

    // The URL alone restores the view.
    await page.reload();
    await expect(filters.getByLabel("Utskott", { exact: true })).toHaveValue("SkU");
    await expect(page.getByRole("searchbox", { name: "Sök i titlar och yrkanden" })).toHaveValue("virtuella varor");
  });

  test("pages with the cursor and says so when nothing matches", async ({ page }) => {
    const cursors: (string | null)[] = [];
    await page.route("**/api/motions?**", async (route) => {
      const c = new URL(route.request().url()).searchParams.get("cursor");
      cursors.push(c);
      await route.fulfill({
        json: c ? { data: [ROW_4202], nextCursor: null } : { data: [ROW_3293], nextCursor: "abc" },
      });
    });
    await page.goto("/motioner");
    await page.getByRole("button", { name: /Nästa/ }).click();
    await expect(page.getByTestId("motion-row")).toContainText("En motion av tre partier");
    await expect(page.getByRole("button", { name: /Nästa/ })).toBeDisabled();
    expect(cursors).toContain("abc");

    await page.unroute("**/api/motions?**");
    await page.route("**/api/motions?**", (route) => route.fulfill({ json: { data: [], nextCursor: null } }));
    await page.goto("/motioner?q=zzz");
    await expect(page.getByText("Inga motioner matchar filtren.")).toBeVisible();
  });
});
