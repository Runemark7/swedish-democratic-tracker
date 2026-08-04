const MONTHS_SV = [
  "januari", "februari", "mars", "april", "maj", "juni",
  "juli", "augusti", "september", "oktober", "november", "december",
];

/**
 * Format an ISO date as Swedish "17 juni 2026".
 *
 * Lives here rather than in a page so that every date the site states reads the
 * same way. It replaced a helper that returned today's date and rendered it as
 * "LIVE · <today>", which claimed a currency the data never had.
 *
 * Returns the input unchanged if it is not an ISO date — a malformed value is
 * shown as-is rather than silently becoming a plausible-looking wrong date.
 */
export function swedishDate(iso: string): string {
  const [y, m, d] = iso.split("-").map(Number);
  if (!y || !m || !d || m < 1 || m > 12) return iso;
  return `${d} ${MONTHS_SV[m - 1]} ${y}`;
}
