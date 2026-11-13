export const MANDATE_LABEL: Record<string, string> = {
  "2022-2026": "2022–2026",
  "2026-2030": "2026–2030",
};

export const STATUS_FILTERS: { slug: string; label: string }[] = [
  { slug: "for", label: "Röstat för" },
  { slug: "emot", label: "Röstat emot" },
  { slug: "blandat", label: "Blandat" },
  { slug: "inte-provat", label: "Inte prövat än" },
  { slug: "arkiv", label: "Arkiv (ingen status)" },
];

/** Mandate periods with Löften, newest first. The codes are mandate_periods.code. */
export const MANDATPERIODER: string[] = ["2026-2030", "2022-2026"];

/** A ?mandatperiod= value the API accepts, or "" (every period). */
export function mandatperiodFromQuery(raw: string | null): string {
  return raw && /^\d{4}-\d{4}$/.test(raw) ? raw : "";
}
