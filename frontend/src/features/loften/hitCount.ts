/** Träffantal for the current Löften selection, e.g. "12 löften" / "1 löfte" / "Inga löften". */
export function loftenHitCount(n: number, party = ""): string {
  if (n === 0) return party ? `Inga löften för ${party}` : "Inga löften";
  const base = n === 1 ? "1 löfte" : `${n} löften`;
  return party ? `${base} för ${party}` : base;
}
