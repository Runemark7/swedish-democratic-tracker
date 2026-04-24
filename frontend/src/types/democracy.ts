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
}

export interface Kpi {
  label: string;
  value: string;
  raw: number;
  target: number;
  worseHigher: boolean;
  unit: string;
  trend: "up" | "down" | "flat";
  delta: string;
  note: string;
}

export interface Authority {
  name: string;
  role: string;
  headcount: string;
  budget: string;
}

export interface LevelData {
  title: string;
  subtitle: string;
  population?: string;
  ruling: Ruling;
  liveVotes: LiveVote[];
  budget: Budget;
  agenda: string[];
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
