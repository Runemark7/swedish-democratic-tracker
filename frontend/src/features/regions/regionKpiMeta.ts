import type { KpiMeta } from "@/features/municipalities/kpiMeta";

export type { KpiMeta };

export const REGION_KPI_META: Record<string, KpiMeta> = {
  N60008: {
    label: "Nettokostnad/inv",
    description: "Regionens driftkostnad per invånare. Speglar servicenivå och effektivitet inom vård och kollektivtrafik.",
    unit: " kr/inv", worseHigher: true, format: v => `${Math.round(v).toLocaleString("sv-SE")} kr`,
  },
  N63016: {
    label: "Resultat/skatt",
    description: "Regionens överskott i förhållande till skatteintäkterna — ett mått på hur stor marginal regionen har i sin ekonomi.",
    unit: "%", worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N63007: {
    label: "Soliditet",
    description: "Hur stor del av regionens tillgångar som är skuldfria. Låg soliditet ökar sårbarheten vid ekonomiska kriser.",
    unit: "%", worseHigher: false, format: v => `${v.toFixed(0)} %`,
  },
  N79173: {
    label: "Primärvård 3 dagar",
    description: "Andel patienter som fick medicinsk bedömning inom tre dagar i primärvården. Visar tillgängligheten till vård i din region.",
    unit: "%", worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N79179: {
    label: "Svar primärvård",
    description: "Andel samtal till primärvården som besvarades samma dag. Låg andel kan tyda på underbemanning eller hög belastning.",
    unit: "%", worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N60404: {
    label: "Kollektivtrafik",
    description: "Antal resor med kollektivtrafik per invånare och år. Speglar hur väl regionen uppfyller sin lagstadgade skyldighet att tillhandahålla allmän kollektivtrafik.",
    unit: " resor/inv", worseHigher: false, format: v => `${Math.round(v)} resor/inv`,
  },
  N85012: {
    label: "Regional utv.",
    description: "Nettokostnad för regional utveckling per invånare. Täcker EU-program, regional strategi och infrastrukturplanering enligt Lag (2010:630).",
    unit: " kr/inv", worseHigher: true, format: v => `${Math.round(v).toLocaleString("sv-SE")} kr/inv`,
  },
};

export const REGION_STRIP_ORDER = ["N60008", "N63016", "N63007", "N79173", "N79179", "N60404", "N85012"];
