/** Plain-language meaning of each procedural term, shown where the term is
 *  first used on the motion page. Definitions describe the procedure only. */
const TERMS: Record<string, string> = {
  yrkande: "Ett yrkande är ett konkret krav i en motion, till exempel att riksdagen ska be regeringen utreda något.",
  betänkande: "Ett betänkande är utskottets skriftliga förslag till riksdagen om hur motionerna ska avgöras.",
  förslagspunkt: "En förslagspunkt är en numrerad punkt i betänkandet; kammaren beslutar punkt för punkt.",
  acklamation: "Acklamation betyder att kammaren avgjorde frågan utan rösträkning, så det finns ingen uppgift om hur varje parti röstade.",
  votering: "En votering är en omröstning där varje ledamots röst registreras.",
  reservation: "En reservation är när ledamöter i utskottet tycker annorlunda än utskottets majoritet och skriver ner det.",
  "återkommande krav": "Återkommande krav är samma yrkande, med identisk ordalydelse, som lämnats in under flera riksmöten.",
};

export function Explain({ terms }: { terms: string[] }) {
  return (
    <div className="text-xs text-on-surface-variant my-1" data-testid="explain">
      {terms.map((t) => (
        <p key={t}>{TERMS[t]}</p>
      ))}
    </div>
  );
}
