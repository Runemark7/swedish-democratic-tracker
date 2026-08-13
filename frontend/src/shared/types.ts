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
  | "pension"
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
  /** Vote points matched to this goal — a fact about our matching.
   *  The alignment percentage was retired: it systematically disadvantaged
   *  opposition parties, scoring a party that voted Nej to the rejection of its
   *  own motion as having broken its own promise. */
  relevantVotes?: number;
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
}

export interface GoalVoteBreakdown {
  goal: GoalWithAlignment;
  matches: GoalVoteMatch[];
}

export interface TopicAlignment {
  topic: Topic;
  /** All goals in this topic. */
  goalCount: number;
}

export interface PartySummary {
  party: PartyCode;
  totalGoals: number;
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


// ── Committees (utskott) ──────────────────────────────────────────────
// Mirrors components["schemas"]["Committee"] in api-contract.ts, which this
// file hand-duplicates. A field that differs in name or nullability fails tsc,
// and that failure is the only check there is.

export interface CommitteeExpenditureArea {
  code: string;
  name: string;
  amountKsek: number;
}

export interface Committee {
  code: string;
  name: string;
  /** Voteringar decided in the period, counted as the record identifies one
   *  (votering_id). A count of the record, never a ranking. */
  voteringar: number;
  /** Populated only by GET /committees/{code}; the list endpoint returns it
   *  empty rather than issuing one query per committee. */
  expenditureAreas: CommitteeExpenditureArea[];
  /** The budget year `expenditureAreas` amounts were allocated for. 0 from the
   *  list endpoint, which populates no amounts. The caller cannot derive it:
   *  the detail endpoint resolves an unspecified year to the newest decided
   *  one, which moves as budgets are seeded. */
  budgetYear: number;
}

/**
 * How a party's members voted on one förslagspunkt.
 *
 * "Delad" means they split evenly between two or more positions. It exists so
 * the record does not resolve a tie to whichever position happens to sort
 * first — that rule once published "L: Frånvarande" for a votering where eight
 * L members voted Ja.
 */
export interface PartyPosition {
  party: PartyCode;
  position: "Ja" | "Nej" | "Avstår" | "Frånvarande" | "Delad";
}

/**
 * One decided förslagspunkt with every party's position.
 *
 * Carries no date. Votes hold only `system_datum`, which is when Riksdagen last
 * touched the row, so rendering it as a decision date would repeat a bug the
 * site already removed.
 */
export interface CommitteeVotering {
  /** Distinguishes the förslagspunkter decided by two separate voteringar —
   *  same beteckning and förslagspunkt, different votering, possibly
   *  contradictory party positions. Without it they read as one decision
   *  contradicting itself. */
  voteringId: string;
  beteckning: string;
  forslagspunkt: string;
  documentTitle: string;
  riksmote: string;
  /** True when the record holds more than one votering for this (riksmöte,
   *  beteckning, förslagspunkt). Computed server-side over the committee's
   *  whole record: a pair can straddle a page boundary, so a page counting its
   *  own duplicates marks neither half. */
  decidedByMultipleVoteringar: boolean;
  partyPositions: PartyPosition[];
}

export interface CommitteeVoteringPage {
  items: CommitteeVotering[];
  /** Total for the committee, independent of paging. */
  total: number;
}

// ── Record coverage ───────────────────────────────────────────────────
export interface MandatePeriod {
  code: string;
  label: string;
  /** Once the period has ended the record is final. */
  ended: boolean;
}

export interface RiksmoteRecord {
  riksmote: string;
  expected: number;
  ingested: number;
  unreachable: number;
  lastDecisionDate: string | null;
  /** When this riksmöte's `expected` was last read from Riksdagen. */
  denominatorCheckedAt: string | null;
}

export interface RecordCoverage {
  mandate: MandatePeriod;
  /** Riksdagen's own count of voteringar in the period, not ours. */
  expected: number;
  ingested: number;
  /** Listed upstream but carrying no beteckning, so not fetchable. */
  unreachable: number;
  /** Date of the newest decision held. Never an ingest timestamp. */
  lastDecisionDate: string | null;
  /**
   * When `expected` was last read from Riksdagen — the oldest such date across
   * the period's riksmöten, since the figure is only as fresh as its stalest
   * part. `ingested` is counted live and always current. Null means the
   * denominator has never been established, in which case no share of the
   * total can be stated.
   */
  denominatorCheckedAt: string | null;
  byRiksmote: RiksmoteRecord[];
}
