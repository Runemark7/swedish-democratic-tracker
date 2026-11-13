const MONTHS_SV = [
  "januari", "februari", "mars", "april", "maj", "juni",
  "juli", "augusti", "september", "oktober", "november", "december",
];

const ISO_DATE = /^(\d{4})-(\d{2})-(\d{2})$/;
const ISO_TIMESTAMP = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}/;
// What String(new Date()) produces: "Mon Jun 08 2026 00:00:00 GMT+0000 (…)".
const JS_DATE_STRING = /^[A-Z][a-z]{2} [A-Z][a-z]{2} \d{2} \d{4} \d{2}:\d{2}:\d{2} GMT[+-]\d{4}/;

const STOCKHOLM_DAY = new Intl.DateTimeFormat("sv-SE", {
  timeZone: "Europe/Stockholm",
  year: "numeric",
  month: "numeric",
  day: "numeric",
});

/** The Swedish calendar day an instant falls on, as [year, month, day]. */
function stockholmDay(instant: Date): [number, number, number] {
  const parts = STOCKHOLM_DAY.formatToParts(instant);
  const part = (type: string) => Number(parts.find((p) => p.type === type)?.value);
  return [part("year"), part("month"), part("day")];
}

/**
 * Format a date as Swedish "17 juni 2026".
 *
 * Lives here rather than in a page so that every date the site states reads the
 * same way. It replaced a helper that returned today's date and rendered it as
 * "LIVE · <today>", which claimed a currency the data never had.
 *
 * Accepts an ISO date ("2026-06-17"), which is formatted as written, and an
 * instant — an ISO timestamp or the string JavaScript makes of a Date — which
 * is dated by the Swedish calendar day it falls on. The latter exists because
 * a Date once leaked into the source registry and was shown to readers as
 * "Mon Jun 08 2026 00:00:00 GMT+0000 (Coordinated Universal Time)".
 *
 * Returns the input unchanged if it is neither — a malformed value is shown
 * as-is rather than silently becoming a plausible-looking wrong date.
 */
export function swedishDate(value: string): string {
  let ymd: [number, number, number];
  const plain = ISO_DATE.exec(value);
  if (plain) {
    ymd = [Number(plain[1]), Number(plain[2]), Number(plain[3])];
  } else if (ISO_TIMESTAMP.test(value) || JS_DATE_STRING.test(value)) {
    const instant = new Date(value);
    if (Number.isNaN(instant.getTime())) return value;
    ymd = stockholmDay(instant);
  } else {
    return value;
  }
  const [y, m, d] = ymd;
  if (!y || !m || !d || m < 1 || m > 12 || d > 31) return value;
  return `${d} ${MONTHS_SV[m - 1]} ${y}`;
}
