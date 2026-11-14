/** What the page needs of a förslagspunkt to state its outcome. */
export interface OutcomeInput {
  mode: string;
  /** Riksdagen's own word for the winning alternative: "utskottet",
   *  "reservation 5", or empty when Riksdagen has not stated one. */
  vinnare: string;
  voteringStored: boolean;
  parties: { party: string; partilinje: string; ja: number; nej: number }[];
}

export interface PunktOutcome {
  /** The outcome in plain words. */
  label: string;
  /** "281 ja, 66 nej" when the punkt was decided by a stored votering. */
  tally: string;
  /** Parties whose Partilinje was the winning / the losing side. Empty when
   *  there was no votering or the sides cannot be told apart. */
  winningParties: string[];
  losingParties: string[];
}

const reservationRe = /^reservation\s*(\d+)$/i;

/** The outcome of one förslagspunkt, or null when it is not known.
 *
 *  Riksdagen's `vinnare` decides it. In a votering Ja is the utskott's
 *  proposal and Nej the reservation it was put against, so the winner also
 *  tells which parties stood behind the winning proposal. Without a `vinnare`
 *  only the tally of a stored votering is reported, worded as a tally. A punkt
 *  with neither, a tie, or a betänkande that is not decided gets no label
 *  rather than a guess. */
export function punktOutcome(k: OutcomeInput, decided: boolean | null | undefined): PunktOutcome | null {
  if (decided === false || k.mode === "ej_beslutad") return null;

  const counted = k.mode === "votering" && k.voteringStored && k.parties.length > 0;
  const ja = counted ? k.parties.reduce((s, p) => s + p.ja, 0) : 0;
  const nej = counted ? k.parties.reduce((s, p) => s + p.nej, 0) : 0;
  const tally = counted ? `${ja} ja, ${nej} nej` : "";
  const line = (side: string) => (counted ? k.parties.filter((p) => p.partilinje === side).map((p) => p.party) : []);
  const sides = (winner: "Ja" | "Nej") => ({
    winningParties: line(winner),
    losingParties: line(winner === "Ja" ? "Nej" : "Ja"),
  });

  const vinnare = k.vinnare.trim().replace(/\s+/g, " ");
  if (vinnare.toLowerCase() === "utskottet") {
    return { label: "Riksdagen biföll utskottets förslag", tally, ...sides("Ja") };
  }
  const reservation = reservationRe.exec(vinnare);
  if (reservation) {
    return { label: `Riksdagen biföll reservation ${reservation[1]} – utskottets förslag avslogs`, tally, ...sides("Nej") };
  }
  if (vinnare) {
    return { label: `Vinnande förslag enligt Riksdagen: ${vinnare}`, tally, winningParties: [], losingParties: [] };
  }
  if (!counted || ja === nej) return null;
  return { label: `Majoriteten röstade ${ja > nej ? "ja" : "nej"}`, tally, ...sides(ja > nej ? "Ja" : "Nej") };
}
