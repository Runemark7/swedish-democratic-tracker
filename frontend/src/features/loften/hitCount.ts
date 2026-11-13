import { MANDATE_LABEL } from "./labels.ts";

/**
 * Träffantal for the current Löften selection, e.g. "12 löften" / "1 löfte" /
 * "Inga löften". A chosen party and mandatperiod are named, so the reader sees
 * what the number counts.
 */
export function loftenHitCount(n: number, party = "", mandatperiod = ""): string {
  const base = n === 0 ? "Inga löften" : n === 1 ? "1 löfte" : `${n} löften`;
  const forParty = party ? ` för ${party}` : "";
  const inPeriod = mandatperiod ? `, mandatperiod ${MANDATE_LABEL[mandatperiod] ?? mandatperiod}` : "";
  return `${base}${forParty}${inPeriod}`;
}
