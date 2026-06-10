import type { Authority, AuthorityDetail } from "@/types/democracy";
import type { components } from "@/shared/api-contract";

export type RegisteredAuthority = components["schemas"]["RegisteredAuthority"];
export type RegisteredAuthorityList = components["schemas"]["RegisteredAuthorityList"];
export type AuthorityStats = components["schemas"]["AuthorityStats"];

export interface MyndigheterListFilter {
  q?: string;
  underGovernment?: boolean;
  page?: number;
  pageSize?: number;
}

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
  worseHigher: boolean;
  unit: string;
  trend: "up" | "down" | "flat";
  delta: string;
  note: string;
  sourceUrl?: string;
  year: number;
}

// ── Government API types (from /api/riksdag/government) ───────────────────────

export interface RiksdagGovernmentParty {
  name: string;
  short: string;
  seats: number;
  color: string;
  role: string;
  sortOrder: number;
}

export interface RiksdagGovernment {
  typeLabel: string;
  validFrom: string;
  parties: RiksdagGovernmentParty[];
  support: RiksdagGovernmentParty[];
  opposition: RiksdagGovernmentParty[];
}

// ── Agenda API types (from /api/riksdag/agenda) ────────────────────────────────

export interface RiksdagAgendaItem {
  id: number;
  title: string;
  description: string;
  source: string;
  status: string;
}

// ── Live votes API types (from /api/riksdag/live-votes) ───────────────────────

export interface RiksdagLiveVote {
  beteckning: string;
  title: string;
  status: string;
  jaCount: number;
  nejCount: number;
  margin: string;
  tag: string;
  date: string;
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

  getAuthorityStats: (): Promise<AuthorityStats> =>
    fetch("/api/riksdag/authorities/stats").then((r) => {
      if (!r.ok) throw new Error(`riksdag/authorities/stats: ${r.status}`);
      return r.json();
    }),

  listMyndigheter: (filter: MyndigheterListFilter = {}): Promise<RegisteredAuthorityList> => {
    const p = new URLSearchParams();
    if (filter.q) p.set("q", filter.q);
    if (filter.underGovernment !== undefined) p.set("underGovernment", String(filter.underGovernment));
    if (filter.page) p.set("page", String(filter.page));
    if (filter.pageSize) p.set("pageSize", String(filter.pageSize));
    const qs = p.toString();
    return fetch(`/api/riksdag/authorities/list${qs ? "?" + qs : ""}`).then((r) => {
      if (!r.ok) throw new Error(`riksdag/authorities/list: ${r.status}`);
      return r.json();
    });
  },

  getKpis: (): Promise<RiksdagKpi[]> =>
    fetch("/api/riksdag/kpis").then((r) => {
      if (!r.ok) throw new Error(`riksdag/kpis: ${r.status}`);
      return r.json();
    }),

  getGovernment: (): Promise<RiksdagGovernment> =>
    fetch("/api/riksdag/government").then((r) => {
      if (!r.ok) throw new Error(`riksdag/government: ${r.status}`);
      return r.json();
    }),

  getAgenda: (): Promise<RiksdagAgendaItem[]> =>
    fetch("/api/riksdag/agenda").then((r) => {
      if (!r.ok) throw new Error(`riksdag/agenda: ${r.status}`);
      return r.json();
    }),

  getAgendaItem: (id: number): Promise<RiksdagAgendaItem> =>
    fetch(`/api/riksdag/agenda/${id}`).then((r) => {
      if (!r.ok) throw new Error(`riksdag/agenda/${id}: ${r.status}`);
      return r.json();
    }),

  getLiveVotes: (): Promise<RiksdagLiveVote[]> =>
    fetch("/api/riksdag/live-votes").then((r) => {
      if (!r.ok) throw new Error(`riksdag/live-votes: ${r.status}`);
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
