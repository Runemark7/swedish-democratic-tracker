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

// ── Status Configuration ──────────────────────────────────────────────
export type GoalStatus = "aligned" | "partial" | "contradiction" | "no_vote";

export const STATUS_CONFIG: Record<GoalStatus, { label: string; color: string; bg: string; icon: string }> = {
  aligned:       { label: "I LINJE",     color: "#16a34a", bg: "#f0fdf4", icon: "✓" },
  partial:       { label: "DELVIS",      color: "#d97706", bg: "#fffbeb", icon: "~" },
  contradiction: { label: "MOTSÄGELSE",  color: "#dc2626", bg: "#fef2f2", icon: "✕" },
  no_vote:       { label: "EJ PRÖVAD",   color: "#9ca3af", bg: "#f9fafb", icon: "—" },
};

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
  other:      "Övrigt",
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
  UU:  "Utrikesutskottet",
};

/** Extract committee name from beteckning, e.g. "SoU12" → "Socialutskottet" */
export function committeeFromBeteckning(beteckning: string): string | undefined {
  const match = beteckning.match(/^([A-ZÅÄÖ][a-zåäöÅÄÖ]*U)/);
  return match ? COMMITTEES[match[1]] : undefined;
}

// ── Helpers ───────────────────────────────────────────────────────────
export function alignmentColor(pct: number): string {
  if (pct >= 75) return "#16a34a";
  if (pct >= 50) return "#d97706";
  return "#dc2626";
}

export function goalStatus(pct: number, votes: number): GoalStatus {
  if (votes === 0) return "no_vote";
  if (pct >= 75) return "aligned";
  if (pct >= 40) return "partial";
  return "contradiction";
}

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
