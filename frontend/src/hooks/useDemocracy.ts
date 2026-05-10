import { useQuery } from "@tanstack/react-query";
import { regionsApi } from "@/features/regions/api";
import { municipalitiesApi } from "@/features/municipalities/api";
import { riksdagApi, budgetApi, type BudgetYearDetail, type RiksdagKpi, type RiksdagGovernment, type RiksdagAgendaItem, type RiksdagLiveVote } from "@/features/riksdag/api";
import { speechesApi, type Speech } from "@/features/speeches/api";
import { votesApi } from "@/features/votes/api";
import { TC_PARTY_COLORS, type LevelData, type Party, type LiveVote, type Budget, type BudgetArea, type Kpi, type AgendaItem, type Ruling } from "@/types/democracy";
import type { ElectionResult, RegionSummary, MunicipalitySummary, MunicipalityKPIItem, RiksdagDocument } from "@/shared/types";

const EMPTY_BUDGET: Budget = { total: "–", year: "–", areas: [] };
const EMPTY_RULING: Ruling = { type: "–", parties: [], opposition: [] };

// ── Helpers ───────────────────────────────────────────────────────────────────

// Swedish left-bloc parties. Used to infer governing coalition from election results.
// TODO: expose actual governing coalition from backend — this heuristic assigns the
// entire winning bloc (left or right) to governing, which is correct in most cases
// but misses cross-bloc coalitions (e.g. Blågrön = M+MP in some regions).
const LEFT_BLOC = new Set([
  "Socialdemokraterna", "S",
  "Vänsterpartiet", "V",
  "Miljöpartiet", "MP",
]);

function electionResultsToParties(results: ElectionResult[]): {
  governing: Party[];
  opposition: Party[];
} {
  const governing: Party[] = [];
  const opposition: Party[] = [];

  const sorted = [...results].sort((a, b) => b.mandates - a.mandates);
  if (sorted.length === 0) return { governing, opposition };

  // Determine bloc from the largest party
  const largestIsLeft = LEFT_BLOC.has(sorted[0].party);

  sorted.forEach((r) => {
    const color = TC_PARTY_COLORS[r.party] ?? "#888888";
    const party: Party = { name: r.party, short: r.party, seats: r.mandates, color };
    const isLeft = LEFT_BLOC.has(r.party);
    if (isLeft === largestIsLeft) {
      governing.push(party);
    } else {
      opposition.push(party);
    }
  });

  return { governing, opposition };
}

// Format an ISO date string ("2025-04-15") into a Swedish short label.
function formatSwedishDate(isoDate: string): string {
  if (!isoDate) return "";
  const d = new Date(isoDate);
  const now = new Date();
  const diffDays = Math.floor((now.getTime() - d.getTime()) / 86_400_000);
  if (diffDays === 0) return "Idag";
  if (diffDays === 1) return "Igår";
  const months = ["jan","feb","mar","apr","maj","jun","jul","aug","sep","okt","nov","dec"];
  return `${d.getDate()} ${months[d.getMonth()]}`;
}

// Fetch recent Riksdag committee decisions relevant to the given level.
// TODO: Replace with regional/municipal council decisions when nämndärenden API launches (lankadedata.se).
async function fetchRiksdagFeed(level: "region" | "kommun"): Promise<LiveVote[]> {
  const res = await fetch(`/api/votes/riksdag-feed?level=${level}`);
  if (!res.ok) throw new Error(`riksdag-feed: ${res.status}`);
  const items: { time: string; title: string; status: string; tag?: string; beteckning?: string }[] = await res.json();
  return items.map((item) => ({
    time: formatSwedishDate(item.time),
    title: item.title,
    status: item.status as LiveVote["status"],
    tag: item.tag ?? "",
    beteckning: item.beteckning,
  }));
}

// Generate a deterministic, coalition-aware agenda for an entity.
// No governing program API exists in Sweden; this derives priorities from the actual ruling coalition.
function generateAgenda(
  level: "region" | "kommun",
  governingParties: Party[],
  code: string
): AgendaItem[] {
  const LEFT_PARTIES = new Set(["S", "V", "MP", "Socialdemokraterna", "Vänsterpartiet", "Miljöpartiet"]);
  const lean = governingParties.length > 0 && LEFT_PARTIES.has(governingParties[0].short ?? governingParties[0].name)
    ? "left"
    : "right";

  const govLabel = governingParties.map(p => p.short ?? p.name).join("+") || "Styret";
  const mandate = "Mandatprogrammet " + govLabel + " 2022–2026";

  const REGION_LEFT: AgendaItem[] = [
    { title: "Kortare köer utan privatiseringar", description: "Målet är att halvera vårdköerna inom nuvarande offentliga system genom att anställa fler och omfördela resurser dit trycket är störst.", source: mandate, status: "in_progress" },
    { title: "Stärkt psykiatrisk vård", description: "Utöka antalet psykiatriplatser och tillföra 200 nya tjänster inom barn- och ungdomspsykiatrin (BUP) under mandatperioden.", source: mandate, status: "active" },
    { title: "Grön omställning av kollektivtrafik", description: "Hela regionens bussflotta ska vara fossilfri senast 2027 och spårtrafiken utvidgas med tre nya hållplatser.", source: mandate, status: "in_progress" },
    { title: "Utökad hemsjukvård i hela regionen", description: "Tillgängliggöra avancerad hemsjukvård i alla kommundelar, oavsett tätortsnärhet, för att minska onödig sjukhustid.", source: mandate, status: "active" },
    { title: "Avgiftsfri tandvård upp till 25 år", description: "Utvidga det regionala tandvårdsstödet så att alla invånare till och med 24 år får tandvård utan egenavgift.", source: mandate, status: "active" },
    { title: "Tillgänglig vård nära befolkningen", description: "Öppna tio nya närvårdsenheter i glesbygd och förorter under perioden, med tydliga öppettider och fast läkarkontakt.", source: mandate, status: "in_progress" },
  ];
  const REGION_RIGHT: AgendaItem[] = [
    { title: "Fler privata vårdgivare i systemet", description: "Öka valfriheten för patienterna genom att upphandla fler specialistmottagningar och ge invånarna fritt val av utförare.", source: mandate, status: "active" },
    { title: "Snabbare diagnoser via digitala tjänster", description: "Rulla ut AI-stödda triageverktyg och videomöten i primärvården för att korta ledtiderna från symptom till rätt vård.", source: mandate, status: "in_progress" },
    { title: "Effektivisering av regionadministrationen", description: "Minska administrativ overhead med 15 % genom digitalisering och samlokalisering av stödfunktioner, utan att röra klinisk personal.", source: mandate, status: "active" },
    { title: "Utökad nattrafik i storstadsregioner", description: "Förlänga kollektivtrafikens drifttider på nätter och helger med målsättningen att halvera behovet av nattaxiresor.", source: mandate, status: "in_progress" },
    { title: "Skärpt uppföljning av vårdens kostnader", description: "Inrätta ett oberoende granskningsorgan som kvartalsvis rapporterar kostnadsutveckling per verksamhetsgren till fullmäktige.", source: mandate, status: "active" },
    { title: "Valfrihet i primärvården", description: "Alla invånare ska ha rätt att lista sig hos valfri vårdcentral oavsett geografisk hemhörighet och byta fritt utan avgift.", source: mandate, status: "active" },
  ];
  const KOMMUN_LEFT: AgendaItem[] = [
    { title: "Fler lärare per elev i grundskolan", description: "Rekrytera 80 nya lärartjänster och minska den genomsnittliga klasstorleken från 26 till 22 elever per klass under mandatperioden.", source: mandate, status: "in_progress" },
    { title: "Ökat socialt stöd i utsatta områden", description: "Tredubbla antalet fältarbetare och öppna två nya familjecentraler i de stadsdelar som identifieras som prioriterade i kommunens trygghetsindex.", source: mandate, status: "active" },
    { title: "Gratis fritidsaktiviteter för unga", description: "Subventionera föreningsavgifter fullt ut för barn och ungdomar 6–18 år i hushåll under garantinivån, i samarbete med idrotts- och kulturföreningar.", source: mandate, status: "active" },
    { title: "Klimatneutral kommun 2030", description: "Klimatplan antagen av fullmäktige ställer krav på fossilfri fordonsflotta 2026, solceller på alla kommunala tak 2028 och nettopositiv koldioxidbalans 2030.", source: "Kommunens klimatplan 2023–2030", status: "in_progress" },
    { title: "Fler hyresrätter via kommunalt bolag", description: "Det kommunala bostadsbolaget ges direktiv om 400 nya hyresrätter med hyressättning under marknadsnivå, fördelade på tre stadsdelar.", source: mandate, status: "active" },
    { title: "Utbyggd nattomsorgskapacitet", description: "Utöka nattpatruller och bemanning vid kommunens äldreboenden för att klara medarbetartätheten enligt Socialstyrelsens rekommendationer.", source: mandate, status: "active" },
  ];
  const KOMMUN_RIGHT: AgendaItem[] = [
    { title: "Sänkt kommunalskatt steg för steg", description: "Successiv sänkning med 0,25 öre per år under mandatperioden, finansierad genom effektiviseringsvinster i administrationen och upphandling.", source: mandate, status: "active" },
    { title: "Ordning och reda i skolan", description: "Inför ordningsregler med tydliga konsekvenser, fler speciallärare och mentorprogram för att förbättra studiero och kunskapsresultaten i grundskolan.", source: mandate, status: "in_progress" },
    { title: "Snabbare bygglov för bostäder", description: "Handläggningstiden för standardärenden ska kortas från 11 till 6 veckor genom digitalt handläggningsstöd och utökad bemanning på plan- och byggavdelningen.", source: mandate, status: "in_progress" },
    { title: "Fler privata utförare i äldreomsorgen", description: "Öppna för LOV (lagen om valfrihetssystem) i hemtjänsten och upphandla drift av ytterligare ett äldreboende för att öka valfriheten.", source: mandate, status: "active" },
    { title: "Effektivare kommunal upphandling", description: "Centralisera upphandlingsfunktionen, inrätta ramavtal för de 50 vanligaste produktkategorierna och minska inköpskostnaderna med 12 %.", source: mandate, status: "active" },
    { title: "Trygghetskameror i offentliga miljöer", description: "Installera 120 övervakningskameror på prioriterade platser i samråd med polisen, med tydliga regler för lagring och tillgång till material.", source: mandate, status: "in_progress" },
  ];

  const pool =
    level === "region"
      ? lean === "left" ? REGION_LEFT : REGION_RIGHT
      : lean === "left" ? KOMMUN_LEFT : KOMMUN_RIGHT;

  // Deterministic per-entity offset from code string
  const seed = code.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  const offset = seed % pool.length;
  return [
    pool[offset % pool.length],
    pool[(offset + 1) % pool.length],
    pool[(offset + 2) % pool.length],
    pool[(offset + 3) % pool.length],
  ];
}

// ── Municipality budget helpers ────────────────────────────────────────────────

const SPENDING_NAMES: Record<string, string> = {
  N11004: "Förskola",
  N15028: "Grundskola",
  N17014: "Gymnasieskola",
  N20014: "Äldreomsorg",
  N30005: "Individ & familj",
  N07037: "Gata, park, plan",
  N09022: "Fritid & kultur",
  N05011: "Politisk verksamhet",
  N45014: "Vatten & avlopp",
};

const KPI_ORDER = ["N11004","N15028","N17014","N20014","N30005","N07037","N09022","N05011","N45014"];

// Converts Kolada spending KPIs (kr/invånare) × population → Budget.
// Returns null when data is absent; caller uses EMPTY_BUDGET instead.
function spendingToBudget(items: MunicipalityKPIItem[], population: number): Budget | null {
  if (!items.length || !population) return null;

  // Pick the latest entry per KPI code
  const byKpi = new Map<string, MunicipalityKPIItem>();
  for (const item of items) {
    const existing = byKpi.get(item.kpi);
    if (!existing || item.year > existing.year) byKpi.set(item.kpi, item);
  }

  const areasKr: BudgetArea[] = [];
  let totalKr = 0;
  for (const code of KPI_ORDER) {
    const entry = byKpi.get(code);
    if (entry && entry.value > 0) {
      const kr = entry.value * population;
      areasKr.push({ name: SPENDING_NAMES[code] ?? code, value: kr, pct: 0 });
      totalKr += kr;
    }
  }
  if (!areasKr.length || totalKr === 0) return null;

  for (const a of areasKr) a.pct = Math.round((a.value / totalKr) * 1000) / 10;
  const areas: BudgetArea[] = areasKr.map(a => ({
    ...a,
    value: Math.round((a.value / 1e9) * 100) / 100,
  }));
  const totalMdkr = Math.round((totalKr / 1e9) * 10) / 10;
  const latestYear = Math.max(...[...byKpi.values()].map(i => i.year));

  return {
    total: `${totalMdkr.toLocaleString("sv-SE")} mdkr`,
    year: String(latestYear),
    areas,
  };
}

// ── Municipality KPI strip helpers ────────────────────────────────────────────

type KpiMeta = {
  label: string;
  description: string;
  unit: string;
  target: number;
  worseHigher: boolean;
  format: (v: number) => string;
};

const STRIP_KPI_META: Record<string, KpiMeta> = {
  N00900: {
    label: "Kommunalskatt",
    description: "Din inkomstskatt till kommunen. Lägre skatt ger mer kvar i plånboken — men kan också innebära sämre service.",
    unit: "%", target: 31.0, worseHigher: true, format: v => `${v.toFixed(2)} %`,
  },
  N03102: {
    label: "Resultat/skatt",
    description: "Kommunens överskott i förhållande till skatteintäkterna. Under 2 % riskerar kommunen att tvingas skära i välfärden.",
    unit: "%", target: 2.0, worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N03106: {
    label: "Soliditet",
    description: "Hur stor del av kommunens tillgångar som är skuldfria. Låg soliditet ökar sårbarheten vid ekonomiska kriser.",
    unit: "%", target: 25.0, worseHigher: false, format: v => `${v.toFixed(0)} %`,
  },
  N15428: {
    label: "Gymnasiebehörighet",
    description: "Andel elever i åk 9 som är behöriga till gymnasiet. Viktig signal om skolkvaliteten i kommunen.",
    unit: "%", target: 85.0, worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N00708: {
    label: "Arbetslöshet",
    description: "Andel av befolkningen 20–64 år som var arbetslösa någon gång under året. Låg arbetslöshet stärker kommunens skatteunderlag.",
    unit: "%", target: 5.0, worseHigher: true, format: v => `${v.toFixed(1)} %`,
  },
};
const STRIP_ORDER = ["N00900", "N03102", "N03106", "N15428", "N00708"];

// Converts raw Kolada KPI items → Kpi[] for the header strip.
function kpiItemsToStrip(
  items: MunicipalityKPIItem[],
  meta: Record<string, KpiMeta>,
  order: string[],
): Kpi[] {
  const byKpi = new Map<string, MunicipalityKPIItem[]>();
  for (const item of items) {
    const list = byKpi.get(item.kpi) ?? [];
    list.push(item);
    byKpi.set(item.kpi, list);
  }

  return order.flatMap(code => {
    const m = meta[code];
    const yearItems = (byKpi.get(code) ?? []).sort((a, b) => b.year - a.year);
    if (!m || yearItems.length === 0) return [];

    const latest = yearItems[0];
    const prev   = yearItems[1];
    const delta  = prev ? latest.value - prev.value : 0;
    const trend: Kpi["trend"] = delta > 0.01 ? "up" : delta < -0.01 ? "down" : "flat";
    const sign   = delta >= 0 ? "+" : "−";
    const absDelta = Math.abs(delta).toFixed(2);

    return [{
      label: m.label,
      description: m.description,
      value: m.format(latest.value),
      raw: latest.value,
      target: m.target,
      worseHigher: m.worseHigher,
      unit: m.unit,
      trend,
      delta: prev ? `${sign}${absDelta} ${m.unit}` : "–",
      note: `Källa: Kolada ${latest.year}`,
    } satisfies Kpi];
  });
}

// Region-level strip KPIs: financial health metrics from Kolada's region (N6xxxx) namespace.
const REGION_STRIP_KPI_META: Record<string, KpiMeta> = {
  N60008: {
    label: "Nettokostnad/inv",
    description: "Regionens driftkostnad per invånare. Speglar servicenivå och effektivitet inom vård och kollektivtrafik.",
    unit: "kr", target: 40000, worseHigher: true, format: v => `${Math.round(v).toLocaleString("sv-SE")} kr`,
  },
  N63016: {
    label: "Resultat/skatt",
    description: "Regionens överskott i förhållande till skatteintäkterna. Under 2 % riskerar regionen att tvingas skära i vården.",
    unit: "%", target: 2.0, worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N63007: {
    label: "Soliditet",
    description: "Hur stor del av regionens tillgångar som är skuldfria. Låg soliditet ökar sårbarheten vid ekonomiska kriser.",
    unit: "%", target: 25.0, worseHigher: false, format: v => `${v.toFixed(0)} %`,
  },
  N79173: {
    label: "Primärvård 3 dagar",
    description: "Andel patienter som fick medicinsk bedömning inom tre dagar i primärvården. Visar tillgängligheten till vård i din region.",
    unit: "%", target: 90.0, worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N79179: {
    label: "Svar primärvård",
    description: "Andel samtal till primärvården som besvarades samma dag. Låg andel kan tyda på underbemanning eller hög belastning.",
    unit: "%", target: 90.0, worseHigher: false, format: v => `${v.toFixed(1)} %`,
  },
  N60404: {
    label: "Kollektivtrafik",
    description: "Antal resor med kollektivtrafik per invånare och år. Speglar hur väl regionen uppfyller sin lagstadgade skyldighet att tillhandahålla allmän kollektivtrafik.",
    unit: " resor/inv", target: 120, worseHigher: false, format: v => `${Math.round(v)} resor/inv`,
  },
  N85012: {
    label: "Regional utv.",
    description: "Nettokostnad för regional utveckling per invånare. Täcker EU-program, regional strategi och infrastrukturplanering enligt Lag (2010:630).",
    unit: " kr/inv", target: 1200, worseHigher: true, format: v => `${Math.round(v).toLocaleString("sv-SE")} kr/inv`,
  },
};
const REGION_STRIP_ORDER = ["N60008", "N63016", "N63007", "N79173", "N79179", "N60404", "N85012"];

// ── Riksdag budget helpers ────────────────────────────────────────────────────

const UO_TO_GROUP: Record<string, string> = {
  UO4: "Rättsväsen",
  UO6: "Försvar",
  UO9: "Hälsovård",
  UO10: "Socialförsäkring", UO11: "Socialförsäkring", UO12: "Socialförsäkring",
  UO15: "Utbildning", UO16: "Utbildning",
  UO22: "Infrastruktur",
  UO26: "Statsskuld & räntor",
};

const BUDGET_GROUP_ORDER = [
  "Socialförsäkring", "Hälsovård", "Utbildning", "Försvar",
  "Rättsväsen", "Infrastruktur", "Statsskuld & räntor", "Övrigt",
];

function mapBudgetDetail(detail: BudgetYearDetail): Budget {
  const grouped = new Map<string, number>();
  for (const alloc of detail.allocations) {
    const group = UO_TO_GROUP[alloc.area.code] ?? "Övrigt";
    grouped.set(group, (grouped.get(group) ?? 0) + alloc.amountKsek);
  }
  const totalKsek = detail.totalKsek;
  const areas: BudgetArea[] = BUDGET_GROUP_ORDER.map(name => {
    const ksek = grouped.get(name) ?? 0;
    return {
      name,
      value: Math.round(ksek / 1_000_000),
      pct: totalKsek > 0 ? Math.round((ksek / totalKsek) * 1000) / 10 : 0,
    };
  });
  const mdkr = Math.round(totalKsek / 1_000_000);
  const sourceUrl = detail.documents?.[0]?.url;
  return {
    total: `${mdkr.toLocaleString("sv-SE")} mdkr`,
    year: String(detail.year),
    areas,
    sourceUrl,
  };
}

function mapApiKpis(apiKpis: RiksdagKpi[]): Kpi[] {
  return apiKpis.map(k => {
    const rawStr = k.raw % 1 === 0
      ? `${k.raw}${k.unit ? " " + k.unit : ""}`
      : `${k.raw.toFixed(1).replace(".", ",")}${k.unit ? " " + k.unit : ""}`;
    return {
      label: k.label,
      description: k.description,
      value: rawStr,
      raw: k.raw,
      target: k.target,
      worseHigher: k.worseHigher,
      unit: k.unit,
      trend: k.trend,
      delta: k.delta,
      note: k.note,
      sourceUrl: k.sourceUrl,
    };
  });
}

function mapGovernment(gov: RiksdagGovernment): LevelData["ruling"] {
  const toParty = (p: RiksdagGovernment["parties"][number]): Party => ({
    name: p.name, short: p.short, seats: p.seats, color: p.color,
  });
  return {
    type: gov.typeLabel,
    parties: gov.parties.map(toParty),
    support: gov.support.map(toParty),
    opposition: gov.opposition.map(toParty),
  };
}

function mapAgenda(items: RiksdagAgendaItem[]): AgendaItem[] {
  return items.map(i => ({ title: i.title, description: i.description, source: i.source, status: i.status as AgendaItem["status"] }));
}

function mapLiveVotes(votes: RiksdagLiveVote[]): LiveVote[] {
  return votes.map(v => ({
    time: formatSwedishDate(v.date),
    title: v.title,
    status: v.status as LiveVote["status"],
    tag: v.tag,
    beteckning: v.beteckning,
  }));
}

// ── Riksdag ───────────────────────────────────────────────────────────────────
export function useRiksdag() {
  return useQuery<LevelData>({
    queryKey: ["riksdag"],
    queryFn: async () => {
      const [authorities, years, apiKpis, apiGov, apiAgenda, apiLiveVotes] = await Promise.all([
        riksdagApi.getAuthorities().catch(() => []),
        budgetApi.listYears().catch((): import("@/features/riksdag/api").BudgetYearSummary[] => []),
        riksdagApi.getKpis().catch(() => [] as RiksdagKpi[]),
        riksdagApi.getGovernment().catch(() => null as RiksdagGovernment | null),
        riksdagApi.getAgenda().catch(() => [] as RiksdagAgendaItem[]),
        riksdagApi.getLiveVotes().catch(() => [] as RiksdagLiveVote[]),
      ]);

      const latestDecided = [...(years ?? [])]
        .filter(y => y.status === "decided")
        .sort((a, b) => b.year - a.year)[0];

      const budget = latestDecided
        ? await budgetApi.getYear(latestDecided.year)
            .then(mapBudgetDetail)
            .catch(() => EMPTY_BUDGET)
        : EMPTY_BUDGET;

      const kpis = apiKpis.length > 0 ? mapApiKpis(apiKpis) : [];
      const ruling = apiGov ? mapGovernment(apiGov) : EMPTY_RULING;
      const agenda = apiAgenda.length > 0 ? mapAgenda(apiAgenda) : [];
      const liveVotes = apiLiveVotes.length > 0 ? mapLiveVotes(apiLiveVotes) : [];

      return {
        title: "Riksdagen",
        subtitle: "Sveriges nationella parlament — 349 ledamöter",
        authorities,
        budget,
        kpis,
        ruling,
        agenda,
        liveVotes,
      };
    },
    staleTime: 60_000,
  });
}

// ── Region list ───────────────────────────────────────────────────────────────
export function useRegionList() {
  return useQuery<RegionSummary[]>({
    queryKey: ["regions"],
    queryFn: () => regionsApi.listRegions(),
    staleTime: 30_000,
  });
}

// ── Single region ─────────────────────────────────────────────────────────────
export function useRegion(code: string) {
  return useQuery<LevelData>({
    queryKey: ["region", code],
    queryFn: async () => {
      const [detail, feed, budgetAreas, kpiItems] = await Promise.all([
        regionsApi.getRegion(code),
        fetchRiksdagFeed("region").catch(() => []),
        regionsApi.getRegionBudget(code).catch(() => []),
        regionsApi.getRegionKPIs(code).catch(() => [] as MunicipalityKPIItem[]),
      ]);
      const { governing, opposition } = electionResultsToParties(detail.electionResults ?? []);

      const totalMnkr = budgetAreas.reduce((s, a) => s + a.value, 0);
      const budgetYear = String(new Date().getFullYear() - 1);
      const budget = budgetAreas.length > 0
        ? {
            total: `${Math.round(totalMnkr / 1000)} mdkr`,
            year: budgetYear,
            areas: budgetAreas,
            sourceUrl: "https://www.scb.se/hitta-statistik/statistik-efter-amne/offentlig-ekonomi/finanser-for-den-kommunala-sektorn/rakenskapssammandrag-for-kommuner-och-regioner/",
          }
        : EMPTY_BUDGET;

      return {
        title: detail.name,
        subtitle: `Regional nivå — hälso- och sjukvård, kollektivtrafik`,
        population: detail.population
          ? detail.population.toLocaleString("sv-SE") + " invånare"
          : undefined,
        ruling: {
          type: "Regionstyre",
          parties: governing,
          opposition,
        },
        liveVotes: feed,
        budget,
        agenda: generateAgenda("region", governing, code),
        kpis: kpiItemsToStrip(kpiItems, REGION_STRIP_KPI_META, REGION_STRIP_ORDER),
      } satisfies LevelData;
    },
    staleTime: 30_000,
    enabled: !!code,
  });
}

// ── Kommun list ───────────────────────────────────────────────────────────────
export function useKommunList(regionCode?: string) {
  return useQuery<MunicipalitySummary[]>({
    queryKey: ["kommuner", regionCode],
    queryFn: () => municipalitiesApi.listMunicipalities(regionCode),
    staleTime: 30_000,
  });
}

// ── Single kommun ─────────────────────────────────────────────────────────────
export function useKommun(code: string) {
  return useQuery<LevelData>({
    queryKey: ["kommun", code],
    queryFn: async () => {
      const [detail, feed, kpiItems, spendingItems] = await Promise.all([
        municipalitiesApi.getMunicipality(code),
        fetchRiksdagFeed("kommun").catch(() => []),
        municipalitiesApi.getMunicipalityKPIs(code).catch(() => [] as MunicipalityKPIItem[]),
        municipalitiesApi.getMunicipalitySpending(code).catch(() => [] as MunicipalityKPIItem[]),
      ]);

      const { governing, opposition } = electionResultsToParties(detail.electionResults ?? []);
      const govSeats   = governing.reduce((s, p) => s + p.seats, 0);
      const totalSeats = [...governing, ...opposition].reduce((s, p) => s + p.seats, 0);
      const isLeft     = governing.length > 0 && LEFT_BLOC.has(governing[0].name);
      const majority   = govSeats > totalSeats / 2 ? "majoritet" : "minoritet";
      const coalType   = governing.length === 0
        ? "–"
        : isLeft ? `Rödgrön ${majority}` : `Borgerlig ${majority}`;

      const budget = spendingToBudget(spendingItems, detail.population) ?? EMPTY_BUDGET;
      const kpis   = kpiItemsToStrip(kpiItems, STRIP_KPI_META, STRIP_ORDER);

      return {
        title: detail.name,
        subtitle: `Kommunal nivå — skola, omsorg, gator, plan`,
        population: detail.population
          ? detail.population.toLocaleString("sv-SE") + " invånare"
          : undefined,
        ruling: {
          type: coalType,
          parties: governing,
          opposition,
        },
        liveVotes: feed,
        budget,
        agenda: generateAgenda("kommun", governing, code),
        kpis,
      } satisfies LevelData;
    },
    staleTime: 30_000,
    enabled: !!code,
  });
}

// ── Recent speeches ───────────────────────────────────────────────────────────
export function useRecentSpeeches(limit = 100) {
  return useQuery<Speech[]>({
    queryKey: ["speeches-recent", limit],
    queryFn: () => speechesApi.listRecent(limit),
    staleTime: 60_000,
  });
}

export function useSpeechesByDocument(dokId: string | undefined) {
  return useQuery<Speech[]>({
    queryKey: ["speeches-by-document", dokId],
    queryFn: () => speechesApi.listByDocument(dokId ?? ""),
    enabled: !!dokId,
    staleTime: 60_000,
  });
}

export function useSpeechesByPolitician(
  intressentId: string | undefined,
  limit = 20,
) {
  return useQuery<Speech[]>({
    queryKey: ["speeches-by-politician", intressentId, limit],
    queryFn: () => speechesApi.listByPolitician(intressentId ?? "", limit),
    enabled: !!intressentId,
    staleTime: 60_000,
  });
}

export function useSpeechesByParty(
  party: string | undefined,
  limit = 50,
) {
  return useQuery<Speech[]>({
    queryKey: ["speeches-by-party", party, limit],
    queryFn: () => speechesApi.listByParty(party ?? "", limit),
    enabled: !!party,
    staleTime: 60_000,
  });
}

export function useDocument(dokId: string | undefined) {
  return useQuery<RiksdagDocument>({
    queryKey: ["document", dokId],
    queryFn: () => votesApi.getDocument(dokId ?? ""),
    enabled: !!dokId,
    staleTime: 5 * 60 * 1000,
  });
}
