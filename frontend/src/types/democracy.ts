export interface Party {
  name: string;
  short: string;
  seats: number;
  color: string;
}

export interface Ruling {
  type: string;
  parties: Party[];
  support?: Party[];
  opposition: Party[];
}

export interface LiveVote {
  time: string;
  title: string;
  status: "Bifall" | "Avslag" | "Återremiss" | "Bordlagd";
  margin?: string;
  tag: string;
  beteckning?: string;
}

export interface BudgetArea {
  name: string;
  value: number;
  pct: number;
}

export interface Budget {
  total: string;
  year: string;
  areas: BudgetArea[];
  sourceUrl?: string;
}

export interface Kpi {
  kpiId: string;      // Kolada KPI code, e.g. "N00900"
  label: string;
  description: string;
  value: string;
  raw: number;
  worseHigher: boolean;
  unit: string;
  trend: "up" | "down" | "flat";
  delta: string;
  note: string;
  sourceUrl?: string;
}

export interface YearlyExpenditure {
  year: number;
  expenditureMdkr: number;
  budgetMdkr?: number;
}

export interface YearlyHeadcount {
  year: number;
  headcountInt: number;
}

export interface Authority {
  slug: string;
  name: string;
  role: string;
  ministry: string;
  headcount: string;
  headcountInt: number;
  description?: string;
  websiteUrl?: string;
  annualReportUrl?: string;
  expenditureMdkr: number;
  budgetMdkr?: number;
  year: number;
  history: YearlyExpenditure[];
  headcountHistory?: YearlyHeadcount[];
}

export interface AgencyRegleringsbrev {
  year: number;
  date: string;
  title: string;
  summary: string;
  url: string;
}

export interface AgencyDecision {
  date: string;
  title: string;
  docType: string;
  summary: string;
  url: string;
}

export interface AuthorityDetail extends Authority {
  mandate: string;
  mandateUrl?: string;
  regleringsbrev: AgencyRegleringsbrev[];
  recentDecisions: AgencyDecision[];
}

/** One whole government document we have registered.
 *
 *  Deliberately carries no description and no status. A description would be our
 *  paraphrase rather than the document's words; a status would be our unsourced
 *  claim about how the government is progressing, stale from the moment it is
 *  written. Title, issuer, date and link let the reader read the source itself. */
export interface AgendaItem {
  id: number;
  title: string;
  source: string;
  issuer: string;
  url: string;
  published: string | null;
}

export interface LevelData {
  title: string;
  subtitle: string;
  population?: string;
  ruling: Ruling;
  /** Riksdag-level live votes only. Region and municipality levels no longer
   *  carry a Riksdag feed (the invented committee→level feed was removed), so
   *  they omit this. */
  liveVotes?: LiveVote[];
  budget: Budget;
  // National agenda only (curated from government policy documents). Region
  // and municipality levels omit this — Sweden publishes no structured
  // primary source for their annual plans, so we show nothing rather than guess.
  agenda?: AgendaItem[];
  kpis?: Kpi[];
  authorities?: Authority[];
}

// Authoritative party colors for Tre Kammare pages (from design handoff)
export const TC_PARTY_COLORS: Record<string, string> = {
  S:  "#E8112D",
  M:  "#1E88E5",
  SD: "#DDD014",
  V:  "#AF0000",
  C:  "#009933",
  KD: "#00558E",
  L:  "#006AB3",
  MP: "#83CF39",
};
