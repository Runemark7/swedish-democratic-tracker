// Frontend-only types — shapes the API does not describe and that are
// computed client-side. Keep nothing here that has a counterpart in
// api/openapi.yaml; those belong in @/shared/api-contract (generated).
// If a type here ever gains an API counterpart, move it there and delete
// this one.

/** National / regional / municipal budget tier, rendered by TierNav. */
export type BudgetTier = "national" | "regional" | "municipality";

/**
 * A single area's budget snapshot. The API exposes the region- and
 * municipality-typed variants (RegionBudgetSnapshot, MunicipalityBudgetSnapshot);
 * this is the untyped shape assembled client-side in useDemocracy.
 */
export interface BudgetSnapshot {
  area_name: string;
  year: number;
  value_mnkr: number;
  total_mnkr: number;
  pct: number;
}

/**
 * A municipality area data point. The API defines RegionAreaDataPoint but
 * has no municipal twin; the municipal query returns this shape.
 */
export interface MunicipalityAreaDataPoint {
  mun_code: string;
  value_mnkr: number;
  total_mnkr: number;
  pct: number;
}