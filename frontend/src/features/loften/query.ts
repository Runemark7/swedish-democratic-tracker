/** Filters of the Löften list. Empty or missing means "any". */
export interface LoftenFilter {
  party?: string;
  utskott?: string;
  status?: string;
  /** riksmöte:beteckning, e.g. 2026/27:SkU2 */
  betankande?: string;
  /** mandate period code, e.g. 2026-2030; empty = every period */
  mandatperiod?: string;
}

/** Query string (with leading "?", or "") shared by the JSON list and the CSV download. */
export function loftenQuery(f: LoftenFilter): string {
  const q = new URLSearchParams();
  if (f.party) q.set("party", f.party);
  if (f.utskott) q.set("utskott", f.utskott);
  if (f.status) q.set("status", f.status);
  if (f.betankande) q.set("betankande", f.betankande);
  if (f.mandatperiod) q.set("mandatperiod", f.mandatperiod);
  const qs = q.toString();
  return qs ? `?${qs}` : "";
}

/** Where the current selection downloads as a CSV file. */
export function loftenCsvHref(f: LoftenFilter): string {
  return `/api/loften.csv${loftenQuery(f)}`;
}
