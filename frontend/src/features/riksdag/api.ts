import type { Authority, AuthorityDetail } from "@/types/democracy";

// ── Budget API types (from /api/budget/years) ─────────────────────────────────

export interface BudgetYearSummary {
  year: number;
  status: string;
  totalKsek: number;
}

export interface BudgetAllocation {
  area: { code: string; name: string; sortOrder: number };
  amountKsek: number;
  amountFormatted: string;
}

export interface BudgetYearDetail {
  year: number;
  status: string;
  totalKsek: number;
  allocations: BudgetAllocation[];
  documents: { type: string; title: string; url: string; description: string }[];
}

// ── KPI API type (from /api/riksdag/kpis) ─────────────────────────────────────

export interface RiksdagKpi {
  id: number;
  label: string;
  description: string;
  raw: number;
  target: number;
  worseHigher: boolean;
  unit: string;
  trend: "up" | "down" | "flat";
  delta: string;
  note: string;
  sourceUrl?: string;
  year: number;
}

// ── API clients ────────────────────────────────────────────────────────────────

export const riksdagApi = {
  getAuthorities: (): Promise<Authority[]> =>
    fetch("/api/riksdag/authorities").then((r) => {
      if (!r.ok) throw new Error(`riksdag/authorities: ${r.status}`);
      return r.json();
    }),

  getAuthority: (slug: string): Promise<AuthorityDetail> =>
    fetch(`/api/riksdag/authorities/${encodeURIComponent(slug)}`).then((r) => {
      if (!r.ok) throw new Error(`riksdag/authorities/${slug}: ${r.status}`);
      return r.json();
    }),

  getKpis: (): Promise<RiksdagKpi[]> =>
    fetch("/api/riksdag/kpis").then((r) => {
      if (!r.ok) throw new Error(`riksdag/kpis: ${r.status}`);
      return r.json();
    }),
};

export const budgetApi = {
  listYears: (): Promise<BudgetYearSummary[]> =>
    fetch("/api/budget/years").then((r) => {
      if (!r.ok) throw new Error(`budget/years: ${r.status}`);
      return r.json();
    }),

  getYear: (year: number): Promise<BudgetYearDetail> =>
    fetch(`/api/budget/years/${year}`).then((r) => {
      if (!r.ok) throw new Error(`budget/years/${year}: ${r.status}`);
      return r.json();
    }),
};
