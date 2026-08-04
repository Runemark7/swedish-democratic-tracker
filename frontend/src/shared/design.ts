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
  S:  { bg: "#E8192C", text: "#fff", light: "#fde8ea", tw: "bg-party-s"  },
  M:  { bg: "#52BDEC", text: "#fff", light: "#e6f5fc", tw: "bg-party-m"  },
  SD: { bg: "#DDDD00", text: "#1a1a1a", light: "#fafae6", tw: "bg-party-sd" },
  C:  { bg: "#009933", text: "#fff", light: "#e6f5ed", tw: "bg-party-c"  },
  V:  { bg: "#DA291C", text: "#fff", light: "#fce8e6", tw: "bg-party-v"  },
  KD: { bg: "#005EA1", text: "#fff", light: "#e6f0f8", tw: "bg-party-kd" },
  L:  { bg: "#006AB3", text: "#fff", light: "#e6f0f8", tw: "bg-party-l"  },
  MP: { bg: "#83CF39", text: "#1a1a1a", light: "#f0fae6", tw: "bg-party-mp" },
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

// ── Status Configuration ──────────────────────────────────────────────
// Alignment is presented as a neutral fact (percentage + raw vote counts),
// never as a verdict. No "aligned/contradiction" labels, icons, or good/bad
// colours — the reader draws the conclusion (fact/interpretation principle).

// ── Specificity Configuration ─────────────────────────────────────────
export const SPECIFICITY_CONFIG: Record<string, { label: string; color: string }> = {
  concrete:    { label: "Konkret",  color: "#7c3aed" },
  directional: { label: "Riktning", color: "#2563eb" },
  rhetorical:  { label: "Retorisk", color: "#9ca3af" },
};

// ── Topic Labels ──────────────────────────────────────────────────────
export const TOPIC_LABELS: Record<string, string> = {
  sjukvard:   "Sjukvård",
  arbete:     "Arbete",
  skatt:      "Skatt",
  bostad:     "Bostad",
  skola:      "Skola",
  klimat:     "Klimat",
  brott:      "Brott",
  invandring: "Invandring",
  forsvar:    "Försvar",
  energi:     "Energi",
  pension:    "Pension",
  other:      "Övrigt",
};

// ── Document Type Labels ──────────────────────────────────────────────
export const DOC_TYPE_LABEL: Record<string, string> = {
  bet: "Betänkande",
  ip: "Interpellation",
  mot: "Motion",
  prop: "Proposition",
  prot: "Protokoll",
  fr: "Skriftlig fråga",
};

// ── Budget Tier Labels ────────────────────────────────────────────────
export const TIER_LABELS: Record<string, string> = {
  national:     "Nationellt",
  regional:     "Regionalt",
  municipality: "Kommunalt",
};

// ── Committee Lookup ──────────────────────────────────────────────────
export const COMMITTEES: Record<string, string> = {
  SoU: "Socialutskottet",
  FöU: "Försvarsutskottet",
  SfU: "Socialförsäkringsutskottet",
  SkU: "Skatteutskottet",
  FiU: "Finansutskottet",
  UbU: "Utbildningsutskottet",
  MJU: "Miljö- och jordbruksutskottet",
  NU:  "Näringsutskottet",
  JuU: "Justitieutskottet",
  CU:  "Civilutskottet",
  AU:  "Arbetsmarknadsutskottet",
  TU:  "Trafikutskottet",
  KU:  "Konstitutionsutskottet",
  KrU: "Kulturutskottet",
  UU:  "Utrikesutskottet",
};

/** Extract committee name from beteckning, e.g. "SoU12" → "Socialutskottet".
 *
 *  The code runs from the start up to and including the final "U", and may
 *  contain interior uppercase ("MJU") or Swedish vowels ("FöU"). The previous
 *  pattern allowed only lowercase after the first letter, so "MJU" never
 *  matched and Miljö- och jordbruksutskottet was invisible site-wide.
 *
 *  The lookahead anchors the code to a digit or end-of-string so a trailing
 *  suffix is tolerated: "AU1y" (yttrande) resolves to "AU". */
export function committeeFromBeteckning(beteckning: string): string | undefined {
  const match = beteckning.match(/^([A-ZÅÄÖ][A-ZÅÄÖa-zåäö]*U)(?=\d|$)/);
  return match ? COMMITTEES[match[1]] : undefined;
}

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
