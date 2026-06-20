// Shared domain types — keep in sync with api/openapi.yaml
// The generated api-contract.ts (from openapi-typescript) is the authoritative source.
// These types are hand-written mirrors for convenience; run `npm run generate:api` to regenerate.

export type PartyCode = "S" | "M" | "SD" | "C" | "V" | "KD" | "L" | "MP" | "unknown";
export type VoteResult = "Ja" | "Nej" | "Avstår" | "Frånvarande";
export type ProposalType = "prop" | "mot" | "bet";
export type Topic =
  | "sjukvard"
  | "forsvar"
  | "invandring"
  | "skatt"
  | "skola"
  | "klimat"
  | "brott"
  | "bostad"
  | "arbete"
  | "energi"
  | "other";
export type Specificity = "concrete" | "directional" | "rhetorical";

export interface Politician {
  intressentId: string;
  firstName: string;
  lastName: string;
  party: PartyCode;
  constituency?: string;
  imageUrl?: string;
  isActive: boolean;
}

export interface PoliticianSummary extends Politician {
  totalVotes?: number;
  promiseCount?: number;
  alignmentPct?: number;
}

export interface PoliticianListResponse {
  data: PoliticianSummary[];
  total: number;
  page: number;
  pageSize: number;
}

export interface Vote {
  id: number;
  voteringId: string;
  politicianId: string;
  party: PartyCode;
  voteResult: VoteResult;
  beteckning: string;
  forslagspunkt: string;
  session: string;
  dokId?: string;
  proposedByParty?: PartyCode;
  proposalType?: ProposalType;
  documentTitle?: string;
  date?: string;
  createdAt: string;
}

export interface VoteListResponse {
  data: Vote[];
  total: number;
  page: number;
  pageSize: number;
}

export interface VoteSummary {
  beteckning: string;
  forslagspunkt: string;
  documentTitle: string;
  proposedByParty?: string;
  proposalType?: string;
}

export interface VoteSummaryListResponse {
  data: VoteSummary[];
  total: number;
  page: number;
  pageSize: number;
}

export interface PartyVotePosition {
  party: PartyCode;
  jaCount: number;
  nejCount: number;
  avstarCount: number;
  franvarandeCount: number;
  dominantVote?: VoteResult;
}

export interface VoteDetail {
  beteckning: string;
  forslagspunkt: string;
  dokId?: string;
  documentTitle: string;
  proposedByParty?: PartyCode;
  proposalType?: ProposalType;
  session: string;
  partyBreakdown: PartyVotePosition[];
  contextNote?: string;
  date?: string;
  status?: string;
  subtitle?: string;
  summary?: string;
  /** Full Riksdagen document body (HTML). Rendered inline on the
   *  beslut detail page so users can read the proposal, motivation
   *  and debate transcript without leaving the site. */
  bodyHtml?: string;
  /** Riksdagen dokuppgift — raw scheduled debate/decision datetimes
   *  ("YYYY-MM-DD HH:MM:SS") and lifecycle text. Fact layer, shown verbatim. */
  debattDate?: string;
  beslutDate?: string;
  statusText?: string;
  /** "Beslut i korthet" — plain-Swedish summary (HTML) from Riksdagen. */
  notis?: string;
}

export interface Goal {
  id: number;
  party: PartyCode;
  goalText: string;
  topic: Topic;
  specificity: Specificity;
  sourceDocument: string;
  keywords?: string[];
  relevantCommittees?: string[];
}

export interface GoalWithAlignment extends Goal {
  relevantVotes?: number;
  alignmentPct?: number;
}

export interface GoalVoteMatch {
  beteckning: string;
  forslagspunkt: string;
  relevanceScore: number;
  alignedDirection: VoteResult;
  explanation?: string;
  proposedByParty?: PartyCode;
  proposalType?: ProposalType;
  contextNote?: string;
  documentTitle?: string;
  partyAlignmentPct?: number;
}

export interface GoalVoteBreakdown {
  goal: GoalWithAlignment;
  matches: GoalVoteMatch[];
}

export interface TopicAlignment {
  topic: Topic;
  goalCount: number;
  alignmentPct: number;
}

export interface PartySummary {
  party: PartyCode;
  totalGoals: number;
  avgAlignmentPct: number;
  topicBreakdown?: TopicAlignment[];
}

// ── Promise Types ─────────────────────────────────────────────────────

export interface PoliticianPromise {
  id: number;
  politicianId: string;
  speechId?: number;
  speechDate?: string;
  speechTopic?: string;
  promiseText: string;
  topic: Topic;
  specificity: Specificity;
  keywords?: string[];
}

export interface PromiseVoteMatch {
  voteId: number;
  beteckning?: string;
  forslagspunkt?: string;
  documentTitle?: string;
  voteResult?: VoteResult;
  relevanceScore: number;
  alignment: "supports" | "contradicts" | "unclear";
  explanation?: string;
  proposedByParty?: PartyCode;
  proposalType?: ProposalType;
}

export interface PromiseWithMatches extends PoliticianPromise {
  voteMatches?: PromiseVoteMatch[];
}

// ── Budget Types ──────────────────────────────────────────────────────

export interface ExpenditureArea {
  code: string;
  name: string;
  description?: string;
  sortOrder: number;
}

export interface BudgetYear {
  year: number;
  status: "decided" | "proposed";
  totalKsek: number;
}

export interface BudgetAllocation {
  area: ExpenditureArea;
  amountKsek: number;
  amountFormatted?: string;
}

export interface BudgetDocumentRef {
  type: "proposition" | "committee_report";
  title: string;
  url: string;
  description: string;
}

export interface BudgetYearDetail {
  year: number;
  status: "decided" | "proposed";
  totalKsek: number;
  allocations: BudgetAllocation[];
  documents?: BudgetDocumentRef[];
}

export interface BudgetComparisonRow {
  area: ExpenditureArea;
  baseAmountKsek: number;
  compareAmountKsek: number;
  deltaKsek: number;
  deltaPct: number;
}

export interface BudgetComparison {
  baseYear: number;
  compareYear: number;
  baseTotalKsek: number;
  compareTotalKsek: number;
  totalDeltaKsek: number;
  totalDeltaPct: number;
  rows: BudgetComparisonRow[];
}

export interface AreaTimeSeriesEntry {
  year: number;
  amountKsek: number;
  documents?: BudgetDocumentRef[];
}

export interface AreaTimeSeries {
  area: ExpenditureArea;
  entries: AreaTimeSeriesEntry[];
}

export type BudgetTier = "national" | "regional" | "municipality";

export interface BudgetTrendEntry {
  year: number;
  amountKsek: number;
  amountFormatted?: string;
}

export interface RelatedBudgetArea {
  code: string;
  name: string;
  trend: BudgetTrendEntry[];
  latestDeltaPct: number;
  shareOfBudgetPct: number;
}

export interface FundingRow {
  code: string;
  name: string;
  deltaKsek: number;
  deltaPct: number;
}

export interface FundingContext {
  year: number;
  topIncreases: FundingRow[];
  topDecreases: FundingRow[];
  totalBudgetDeltaKsek: number;
  totalBudgetDeltaPct: number;
}

export interface TopicContext {
  topic: string;
  description: string;
  committees: string[];
  relatedAreas: RelatedBudgetArea[];
  fundingContext?: FundingContext;
}

// ── Region & Municipality Types ───────────────────────────────────────

export interface ElectionResult {
  party: string;
  mandates: number;
  votePct: number;
  totalMandates: number;
}

export interface RegionSummary {
  code: string;
  name: string;
  capital: string;
  population: number;
  electionYear: number;
  governingParties: string[];
  totalMandates: number;
}

export interface RegionDetail extends RegionSummary {
  electionResults: ElectionResult[];
}

export interface RegionPlanGoal {
  title: string;
  description?: string;
}

export interface RegionPlan {
  code: string;
  label: string;
  url: string;
  goalsLabel?: string;
  period?: string;
  goals: RegionPlanGoal[];
}

export interface RegionBudgetArea {
  name: string;
  value: number;
  pct: number;
}

export interface RegionBudgetSnapshot {
  region_code: string;
  area_name: string;
  year: number;
  value_mnkr: number;
  total_mnkr: number;
  pct: number;
}

export interface MunicipalityBudgetSnapshot {
  mun_code: string;
  area_name: string;
  year: number;
  value_mnkr: number;
  total_mnkr: number;
  pct: number;
}

export interface BudgetSnapshot {
  area_name: string;
  year: number;
  value_mnkr: number;
  total_mnkr: number;
  pct: number;
}

export interface RegionAreaDataPoint {
  region_code: string;
  value_mnkr: number;
  total_mnkr: number;
  pct: number;
}

export interface MunicipalityAreaDataPoint {
  mun_code: string;
  value_mnkr: number;
  total_mnkr: number;
  pct: number;
}

export interface MunicipalitySummary {
  code: string;
  name: string;
  regionCode: string;
  regionName: string;
  population: number;
  electionYear: number;
  governingParties: string[];
  totalMandates: number;
}

export interface MunicipalityDetail extends MunicipalitySummary {
  electionResults: ElectionResult[];
}

export interface MunicipalityKPIItem {
  kpi: string;
  year: number;
  value: number;
  status: string;
}

export interface RegionKPIRankEntry {
  region_code: string;
  name: string;
  value: number;
  year: number;
  rank: number;
  total: number;
}

export interface KPIRank {
  kpi: string;
  rank: number;
  total: number;
  mean: number;
}

export interface PopulationTrendEntry {
  year: number;
  population: number;
}

export interface ProcurementCategorySummary {
  cpv_division: string;
  label: string;
  total_value_sek: number;
  pct: number;
  count: number;
}

export interface RiksdagDocument {
  dokId: string;
  type: string;
  title: string;
  subtitle?: string;
  summary?: string;
  date?: string;
  beteckning?: string;
}

export interface Intressent {
  intressentId: string;
  name: string;
  party: string;
  role?: string;
}

export interface RiksdagDocumentFull extends RiksdagDocument {
  bodyHtml?: string;
  intressenter: Intressent[];
}

