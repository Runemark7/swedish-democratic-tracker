import { test, expect, type Page } from "@playwright/test";

// Printing a decision (issue #306). Browsers drop backgrounds by default when
// printing, so anything that is light text on a filled background disappears,
// and the default dark theme prints near-white text on white paper.

const PARTIES = ["S", "M", "SD", "C", "V", "KD", "L", "MP"];

function betankande() {
  return {
    riksmote: "2025/26", beteckning: "FiU47", title: "Extra ändringsbudget", organ: "FiU", dokId: "HD01FiU47",
    sourceUrl: "https://data.riksdagen.se/dokument/HD01FiU47.html", held: true, status: "Webbpublicering",
    decided: true, decisionDate: "2026-03-25", decisionText: "",
    punkter: [{
      number: "1", rubrik: "Extra ändringsbudget", text: "Utskottet föreslår bifall.", beslutstyp: "röstning",
      mode: "votering", modeLabel: "Avgjord genom omröstning", alternativ: "Utskottets förslag mot reservation 1 (S)",
      vinnare: "utskottet", voteringId: "V1", voteringStored: true,
      parties: PARTIES.map((party) => ({ party, partilinje: "Ja", ja: 8, nej: 0, avstar: 0, franvarande: 1 })),
      avvikare: [], reasoning: "", reasoningAvailable: false, reservationsAvailable: false,
      reservations: [], sarskildaYttranden: [], votedReservation: null, motions: [],
    }],
  };
}

async function open(pw: Page, theme: "dark" | "light") {
  await pw.addInitScript((t) => localStorage.setItem("sdt-theme", t), theme);
  await pw.route("**/api/betankanden/**", (route) => route.fulfill({ json: betankande() }));
  await pw.route("**/api/speeches**", (route) => route.fulfill({ json: [] }));
  await pw.goto("/beslut/2025/26/FiU47");
  await expect(pw.getByTestId("party-table")).toBeVisible();
}

/** Relative luminance (0 black – 1 white) of a computed rgb()/rgba() colour. */
function luminance(color: string): number {
  const [r, g, b] = (color.match(/[\d.]+/g) ?? []).slice(0, 3).map(Number);
  return (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255;
}

for (const viewport of [{ width: 1280, height: 800 }, { width: 390, height: 844 }]) {
  test.describe(`print at ${viewport.width}px`, () => {
    test.use({ viewport });

    test("the menu and footer links are not printed", async ({ page: pw }) => {
      await open(pw, "dark");
      await pw.emulateMedia({ media: "print" });
      await expect(pw.locator("header")).toHaveCount(2);
      for (const chrome of await pw.locator("header, footer").all()) {
        await expect(chrome).toBeHidden();
      }
      await expect(pw.getByRole("heading", { level: 1 })).toBeVisible();
      await expect(pw.getByTestId("party-table")).toBeVisible();
    });
  });
}

test.describe("print colours", () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  for (const theme of ["dark", "light"] as const) {
    test(`every party badge prints as black text without a fill (${theme} theme)`, async ({ page: pw }) => {
      await open(pw, theme);
      const badges = pw.getByTestId("party-table").locator(".party-badge");
      await expect(badges).toHaveCount(PARTIES.length);
      await pw.emulateMedia({ media: "print" });
      for (const party of PARTIES) {
        const badge = badges.filter({ hasText: new RegExp(`^${party}$`) });
        const style = await badge.evaluate((el) => {
          const s = getComputedStyle(el);
          return { color: s.color, background: s.backgroundColor, border: s.borderTopStyle, borderWidth: s.borderTopWidth };
        });
        expect(style.color, party).toBe("rgb(0, 0, 0)");
        expect(style.background, party).toBe("rgba(0, 0, 0, 0)");
        expect(style.border, party).toBe("solid");
        expect(style.borderWidth, party).toBe("1px");
      }
    });
  }

  test("the dark theme prints dark text, not near-white on white paper", async ({ page: pw }) => {
    await open(pw, "dark");
    await pw.emulateMedia({ media: "print" });
    for (const el of [pw.getByRole("heading", { level: 1 }), pw.getByTestId("party-table").locator("td").nth(1), pw.locator(".sdt-page").first()]) {
      const color = await el.evaluate((node) => getComputedStyle(node).color);
      expect(luminance(color), color).toBeLessThan(0.4);
    }
    const paper = await pw.locator(".sdt-page").first().evaluate((node) => getComputedStyle(node).backgroundColor);
    expect(paper).toBe("rgb(255, 255, 255)");
  });

  test("screen styling is unchanged: badges keep their party colour", async ({ page: pw }) => {
    await open(pw, "dark");
    const s = pw.getByTestId("party-table").locator(".party-badge").filter({ hasText: /^S$/ });
    await expect(s).toHaveCSS("background-color", "rgb(232, 25, 44)");
    await expect(s).toHaveCSS("color", "rgb(255, 255, 255)");
    await expect(s).toHaveCSS("border-top-width", "0px");
    await expect(pw.locator("header.desktop-only")).toBeVisible();
  });
});
