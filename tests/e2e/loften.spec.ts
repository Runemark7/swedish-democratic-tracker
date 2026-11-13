import { test, expect } from "@playwright/test";

// Shapes as backend/internal/apitest/loften_test.go asserts them.
const ARCHIVED = {
  id: 7, party: "M", valmanifestId: 3, text: "Fler poliser", topic: "trygghet",
  sourceDocument: "Valmanifest 2022", utskott: ["JuU"], mandatePeriod: "2022-2026",
  visibility: "archived", archived: true, createdAt: "2022-08-01T00:00:00Z",
};
const VALMANIFEST = [
  { id: 3, party: "M", electionYear: 2022, title: "Så får vi ordning på Sverige", url: "https://moderaterna.se/nyhet/moderaternasvalmanifest/", state: "final" },
  { id: 9, party: "M", electionYear: 2018, title: "Moderaternas valmanifest 2018", url: "https://snd.gu.se/sv/vivill/party/m/v/2018", state: "draft" },
];

test.describe("Löften under Partier", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test.beforeEach(async ({ page }) => {
    await page.route("**/api/parties/M/loften", (r) => r.fulfill({ json: [ARCHIVED] }));
    await page.route("**/api/parties/M/valmanifest", (r) => r.fulfill({ json: VALMANIFEST }));
    await page.route("**/api/parties/M/record", (r) => r.fulfill({ json: { party: "M", decisions: [], motions: [] } }));
    await page.route("**/api/loften", (r) => r.fulfill({ json: [ARCHIVED] }));
    await page.route("**/api/valmanifest", (r) => r.fulfill({ json: VALMANIFEST }));
    await page.route("**/api/loften/7", (r) => r.fulfill({ json: ARCHIVED }));
  });

  test("Löften are reached under Partier and there is no Löften item in the top menu", async ({ page }) => {
    await page.route("**/api/party-meta", (r) => r.fulfill({ json: [] }));
    await page.goto("/parties");
    const primary = page.getByRole("navigation", { name: "Huvudmeny" });
    await expect(primary.getByRole("link", { name: /Löfte|Manifest/ })).toHaveCount(0);

    await page.getByRole("link", { name: "Löften för alla partier" }).click();
    await expect(page).toHaveURL(/\/loften$/);
    await expect(page.getByTestId("loefte-card")).toContainText("Fler poliser");
    await expect(primary.getByRole("link", { name: /Löfte/ })).toHaveCount(0);

    await page.goto("/parties/M?tab=loften");
    await expect(page.getByTestId("archive-note")).toContainText("2022–2026");
    await expect(page.getByTestId("loefte-card")).toHaveCount(1);
    // The archive shows no Löftesstatus.
    await expect(page.getByText(/Röstat för|Röstat emot|Blandat|Inte prövat än/)).toHaveCount(0);
  });

  test("the party page lists each Valmanifest as a source with its draft or final marking", async ({ page }) => {
    await page.goto("/parties/M");
    const rows = page.getByTestId("valmanifest-row");
    await expect(rows).toHaveCount(2);
    await expect(rows.first()).toContainText("Slutligt");
    await expect(rows.nth(1)).toContainText("Utkast");
    await expect(rows.first().getByRole("link")).toHaveAttribute("href", /moderaterna\.se/);
  });

  test("the old party goal votes URL redirects to the archived Löfte with the same id", async ({ page }) => {
    await page.goto("/parties/M/goals/7/votes");
    await expect(page).toHaveURL(/\/loften\/7$/);
    await expect(page.getByTestId("loefte-text")).toHaveText("Fler poliser");
  });

  test("the old party goals URL redirects to that party's Löften", async ({ page }) => {
    await page.goto("/parties/M/goals");
    await expect(page).toHaveURL(/\/parties\/M\?tab=loften/);
  });

  test("the manifestos URL redirects to the party's Valmanifest list", async ({ page }) => {
    await page.goto("/manifestos?party=M");
    await expect(page).toHaveURL(/\/parties\/M\?tab=loften/);
    await expect(page.getByTestId("valmanifest-row")).toHaveCount(2);

    await page.goto("/manifestos");
    await expect(page).toHaveURL(/\/loften#valmanifest$/);
    await expect(page.getByTestId("valmanifest-row")).toHaveCount(2);
  });
});

// Gap promise-list-no-term-filter: Mandatperiod is its own filter, separate
// from Löftesstatus, on /loften and on a party's Löften tab.
test.describe("Mandatperiod filter on Löften", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  const OLD = {
    id: 7, party: "S", valmanifestId: 3, text: "Gammalt löfte", topic: "skatt",
    sourceDocument: "Valmanifest 2022", utskott: ["SkU"], mandatePeriod: "2022-2026",
    visibility: "archived", archived: true, createdAt: "2022-08-01T00:00:00Z", status: null,
  };
  const NEW = {
    id: 90, party: "S", valmanifestId: 12, text: "Nytt löfte", topic: "skatt",
    sourceDocument: "Valmanifest 2026", utskott: ["SkU"], mandatePeriod: "2026-2030",
    visibility: "approved", archived: false, createdAt: "2026-08-01T00:00:00Z", status: "Inte prövat än",
  };
  const row = (n: number, arkiv: number) =>
    ({ party: "S", total: n, for: 0, emot: 0, blandat: 0, inteProvat: n - arkiv, arkiv });

  test.beforeEach(async ({ page }) => {
    // The mock filters as the API does (backend/internal/apitest/loften_mandatperiod_test.go).
    const list = (url: string) => {
      const q = new URL(url).searchParams;
      return [OLD, NEW].filter((l) =>
        (!q.get("mandatperiod") || l.mandatePeriod === q.get("mandatperiod")) &&
        (q.get("status") !== "inte-provat" || l.status === "Inte prövat än"));
    };
    await page.route(/\/api\/loften(\?.*)?$/, (r) => r.fulfill({ json: list(r.request().url()) }));
    await page.route(/\/api\/parties\/S\/loften(\?.*)?$/, (r) => r.fulfill({ json: list(r.request().url()) }));
    await page.route(/\/api\/loften\/summary(\?.*)?$/, (r) => {
      const p = new URL(r.request().url()).searchParams.get("mandatperiod");
      return r.fulfill({ json: [p === "2026-2030" ? row(1, 0) : p === "2022-2026" ? row(1, 1) : row(2, 1)] });
    });
    await page.route("**/api/valmanifest", (r) => r.fulfill({ json: [] }));
    await page.route("**/api/parties/S/valmanifest", (r) => r.fulfill({ json: [] }));
    await page.route("**/api/parties/S/record", (r) => r.fulfill({ json: { party: "S", decisions: [], motions: [] } }));
  });

  test("/loften has a Mandatperiod filter that combines with Löftesstatus", async ({ page }) => {
    await page.goto("/loften");
    const select = page.getByTestId("mandatperiod-select");
    await expect(select).toHaveValue("");
    await expect(page.getByTestId("loften-summary-period")).toHaveText("Alla mandatperioder");
    await expect(page.getByTestId("loefte-card")).toHaveCount(2);

    await select.selectOption({ label: "2026–2030" });
    await expect(page).toHaveURL(/mandatperiod=2026-2030/);
    await expect(page.getByTestId("loefte-card")).toHaveCount(1);
    await expect(page.getByTestId("loefte-card")).toContainText("Nytt löfte");
    await expect(page.getByTestId("loften-hit-count")).toHaveText("1 löfte, mandatperiod 2026–2030");
    await expect(page.getByTestId("loften-summary-period")).toHaveText("Mandatperiod 2026–2030");
    await expect(page.getByTestId("summary-S").locator("td").last()).toHaveText("0");

    await page.getByTestId("loften-filters").getByRole("combobox", { name: "Löftesstatus" }).selectOption("inte-provat");
    await expect(page).toHaveURL(/mandatperiod=2026-2030/);
    await expect(page).toHaveURL(/status=inte-provat/);
    await expect(page.getByTestId("loefte-card")).toHaveCount(1);

    await select.selectOption({ label: "2022–2026" });
    await expect(page.getByTestId("loften-hit-count")).toHaveText("Inga löften, mandatperiod 2022–2026");
    await expect(page.getByTestId("loefte-card")).toHaveCount(0);
  });

  // Gap no-data-export-csv: the download is the visible selection, so its
  // link carries every active filter, Mandatperiod included.
  test("the CSV download link follows the active filters", async ({ page }) => {
    await page.goto("/loften");
    const csv = page.getByTestId("loften-csv");
    await expect(csv).toHaveText("Ladda ner urvalet (CSV)");
    await expect(csv).toHaveAttribute("href", "/api/loften.csv");

    await page.getByTestId("mandatperiod-select").selectOption({ label: "2026–2030" });
    await expect(csv).toHaveAttribute("href", "/api/loften.csv?mandatperiod=2026-2030");

    await page.getByTestId("loften-filters").getByRole("combobox", { name: "Löftesstatus" }).selectOption("inte-provat");
    await expect(csv).toHaveAttribute("href", "/api/loften.csv?status=inte-provat&mandatperiod=2026-2030");

    // Nothing to download when the selection is empty.
    await page.getByTestId("mandatperiod-select").selectOption({ label: "2022–2026" });
    await expect(page.getByTestId("loefte-card")).toHaveCount(0);
    await expect(csv).toHaveCount(0);
  });

  test("a link with ?mandatperiod= opens with that period chosen", async ({ page }) => {
    await page.goto("/loften?mandatperiod=2022-2026");
    await expect(page.getByTestId("mandatperiod-select")).toHaveValue("2022-2026");
    await expect(page.getByTestId("loefte-card")).toHaveCount(1);
    await expect(page.getByTestId("loefte-card")).toContainText("Gammalt löfte");
  });

  test("the party's Löften tab has the same Mandatperiod filter", async ({ page }) => {
    await page.goto("/parties/S?tab=loften");
    const select = page.getByTestId("mandatperiod-select");
    await expect(select).toHaveValue("");
    await expect(page.getByTestId("loefte-card")).toHaveCount(2);

    await select.selectOption({ label: "2026–2030" });
    await expect(page).toHaveURL(/tab=loften/);
    await expect(page).toHaveURL(/mandatperiod=2026-2030/);
    await expect(page.getByTestId("loefte-card")).toHaveCount(1);
    await expect(page.getByTestId("loefte-card")).toContainText("Nytt löfte");
    await expect(page.getByTestId("loften-hit-count")).toHaveText("1 löfte för S, mandatperiod 2026–2030");
  });
});

// Gap promise-summary-zeros-for-missing-data: "no Löften published for the
// period" must not look like "zero Löften with this status".
test.describe("Löften summary table", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  const SUMMARY = [
    // Only archived Löften: nothing published for the current period.
    { party: "M", total: 10, for: 0, emot: 0, blandat: 0, inteProvat: 0, arkiv: 10 },
    // Published Löften, with genuine zeros for some statuses.
    { party: "S", total: 16, for: 3, emot: 0, blandat: 0, inteProvat: 7, arkiv: 6 },
  ];

  test.beforeEach(async ({ page }) => {
    await page.route("**/api/loften", (r) => r.fulfill({ json: [] }));
    await page.route("**/api/valmanifest", (r) => r.fulfill({ json: [] }));
    await page.route("**/api/loften/summary", (r) => r.fulfill({ json: SUMMARY }));
  });

  test("a party without published Löften shows a text, not zeros, in the status columns", async ({ page }) => {
    await page.goto("/loften");
    const m = page.getByTestId("summary-M");
    await expect(m.getByTestId("summary-no-loften")).toHaveText("Inga löften publicerade ännu");
    await expect(m.getByTestId("summary-status-count")).toHaveCount(0);
    // The archive count is real data and stays.
    await expect(m.getByTestId("summary-arkiv")).toHaveText("10");
    await expect(m.getByRole("link", { name: "M" })).toHaveAttribute("href", "/loften?party=M");
  });

  test("a party the summary does not mention at all gets the same text", async ({ page }) => {
    await page.goto("/loften");
    await expect(page.getByTestId("loften-summary").locator("tbody tr")).toHaveCount(8);
    const v = page.getByTestId("summary-V");
    await expect(v.getByTestId("summary-no-loften")).toHaveText("Inga löften publicerade ännu");
    await expect(v.getByTestId("summary-status-count")).toHaveCount(0);
  });

  test("a party with published Löften keeps its counts, genuine zeros included", async ({ page }) => {
    await page.goto("/loften");
    const s = page.getByTestId("summary-S");
    await expect(s.getByTestId("summary-no-loften")).toHaveCount(0);
    await expect(s.getByTestId("summary-status-count")).toHaveText(["3", "0", "0", "7"]);
    await expect(s.getByTestId("summary-arkiv")).toHaveText("6");
  });
});
