// ── Party Configuration ────────────────────────────────────────────────
export type PartyCode = "S" | "M" | "SD" | "C" | "V" | "KD" | "L" | "MP";

export const PARTIES: PartyCode[] = ["S", "M", "SD", "C", "V", "KD", "L", "MP"];

export interface PartyStyle {
  bg: string;
  text: string;
  light: string;
  tw: string; // tailwind bg class
}

export const PARTY_COLORS: Record<string, PartyStyle> = {
  S:  { bg: "#E8192C", text: "#fff", light: "var(--color-party-s-soft)", tw: "bg-party-s"  },
  M:  { bg: "#52BDEC", text: "#fff", light: "var(--color-party-m-soft)", tw: "bg-party-m"  },
  SD: { bg: "#DDDD00", text: "#1a1a1a", light: "var(--color-party-sd-soft)", tw: "bg-party-sd" },
  C:  { bg: "#009933", text: "#fff", light: "var(--color-party-c-soft)", tw: "bg-party-c"  },
  V:  { bg: "#DA291C", text: "#fff", light: "var(--color-party-v-soft)", tw: "bg-party-v"  },
  KD: { bg: "#005EA1", text: "#fff", light: "var(--color-party-kd-soft)", tw: "bg-party-kd" },
  L:  { bg: "#006AB3", text: "#fff", light: "var(--color-party-l-soft)", tw: "bg-party-l"  },
  MP: { bg: "#83CF39", text: "#1a1a1a", light: "var(--color-party-mp-soft)", tw: "bg-party-mp" },
};

const PARTY_NAMES: Record<string, string> = {
  S: "Socialdemokraterna",
  M: "Moderaterna",
  SD: "Sverigedemokraterna",
  C: "Centerpartiet",
  V: "Vänsterpartiet",
  KD: "Kristdemokraterna",
  L: "Liberalerna",
  MP: "Miljöpartiet",
};

export function partyShortToName(short: string): string {
  return PARTY_NAMES[short] ?? short;
}

// ── Document Type Labels ──────────────────────────────────────────────
export const DOC_TYPE_LABEL: Record<string, string> = {
  bet: "Betänkande",
  ip: "Interpellation",
  mot: "Motion",
  prop: "Proposition",
  prot: "Protokoll",
  fr: "Skriftlig fråga",
};

// ── Who-decides sentence ─────────────────────────────────────────────
// The constitutional division of responsibility, stated once and reused by
// both level-landing pages. A fact about who decides what, not a data link.
export function whoDecidesSentence(level: "region" | "kommun"): string {
  const body = level === "region" ? "Din region" : "Din kommun";
  return `Riksdagen beslutar om statens budget. ${body} beslutar om sin egen, med egna skatteintäkter och statsbidrag. De två är inte samma pengar.`;
}

// ── Budget Tier Labels ────────────────────────────────────────────────
export const TIER_LABELS: Record<string, string> = {
  national:     "Nationellt",
  regional:     "Regionalt",
  municipality: "Kommunalt",
};

// ── Helpers ───────────────────────────────────────────────────────────

/** Format KSEK to Swedish budget display: "145,2 mdkr" or "4 350 mnkr" */
export function formatBudgetAmount(ksek: number): string {
  const mdkr = ksek / 1_000_000;
  if (Math.abs(mdkr) >= 1) {
    return mdkr.toFixed(1).replace(".", ",") + " mdkr";
  }
  const mnkr = ksek / 1_000;
  return Math.round(mnkr).toLocaleString("sv-SE") + " mnkr";
}

/** Color for budget delta: green for increase, red for decrease */
export function deltaColor(pct: number): string {
  if (pct > 0) return "#16a34a";
  if (pct < 0) return "#dc2626";
  return "#9ca3af";
}
