import { Link } from "react-router-dom";
import { SourceMarker } from "@/components/sources/SourceMarker";

/** How Löftesstatus is computed, and where the method stops (ADR-002). */
export function MetodPage() {
  return (
    <div className="sdt-page" style={{ padding: "24px 32px 64px", maxWidth: 760 }}>
      <Link to="/loften" className="text-xs text-link">← Löften</Link>
      <h1 className="font-display text-3xl mt-2 mb-4">Metod: så räknas Löftesstatus</h1>
      <div className="text-sm leading-relaxed space-y-4" data-testid="metod">
        <p>
          Ett <strong>Löfte</strong> är ett konkret uttalande i ett partis valmanifest, med citat och
          länk till originalet. Löftesstatus säger bara vad partiet <em>gjorde i riksdagen</em>. Den
          säger inte om ett löfte infriats, och den avgör inte om partiet gjort rätt.
        </p>
        <h2 className="font-display text-lg">Steg för steg</h2>
        <ol className="list-decimal pl-5 space-y-1">
          <li>Redaktören kopplar ett Löfte till en omröstning och anger vilken röst, Ja eller Nej, som går i Löftets riktning. Först när kopplingen är <em>bekräftad</em> räknas den.</li>
          <li>För varje bekräftad omröstning tittar vi på partiets <strong>Partilinje</strong>: den röst som flest av partiets ledamöter avgav. Frånvarande räknas inte.</li>
          <li>Är Partilinjen samma som Löftets riktning räknas omröstningen som <strong>för</strong>. Är den motsatt räknas den som <strong>emot</strong>.</li>
          <li>Är Partilinjen Avstår, eller är partiet delat, anger den ingen riktning. Omröstningen visas men räknas varken för eller emot.</li>
          <li>Bara omröstningar inom Löftets mandatperiod räknas.</li>
        </ol>
        <h2 className="font-display text-lg">De fyra statusarna</h2>
        <ul className="list-disc pl-5 space-y-1">
          <li><strong>Röstat för</strong>: alla räknade omröstningar går i Löftets riktning.</li>
          <li><strong>Röstat emot</strong>: alla räknade omröstningar går mot Löftets riktning.</li>
          <li><strong>Blandat</strong>: både för och emot förekommer.</li>
          <li><strong>Inte prövat än</strong>: ingen omröstning har kunnat räknas för eller emot. Det gäller också när alla bekräftade omröstningar slutade i Avstår eller Delad. De visas ändå på Löftets sida.</li>
        </ul>
        <p>
          Gamla löften från valet 2022 ligger i ett arkiv. De har ingen Löftesstatus, eftersom vi inte
          kopplat dem till omröstningar och en status då skulle beskriva vår egen lucka, inte partiet.
        </p>
        <h2 className="font-display text-lg">Vad metoden inte kan säga</h2>
        <ul className="list-disc pl-5 space-y-1">
          <li>En röst för eller emot är inte detsamma som att löftet genomförts. Oppositionen kan inte genomföra något på egen hand.</li>
          <li>En omröstning avgör ofta flera frågor på en gång. Redaktören avgör vilken omröstning som prövar Löftet, och vi har kanske inte hittat alla.</li>
          <li>Partilinjen är en majoritet. Ledamöter som röstat annorlunda, Avvikare, visas men ändrar aldrig partiets status.</li>
          <li>Ett parti kan ha goda skäl till en röst som inte syns i omröstningen. Tolkningen är din.</li>
        </ul>
        <h2 className="font-display text-lg" id="vem-bedomer">Vem bedömer</h2>
        <p data-testid="vem-bedomer">
          Boten föreslår Löften och kopplingar. En människa, Redaktören, bekräftar eller avvisar dem i
          chatten. Först då blir de offentliga. Varje beslut loggas med en referens i den öppna{" "}
          <Link to="/andringslogg" className="text-link underline">Ändringslogg</Link>.
          Den här sidan säger inte om fler än en person granskar bedömningarna.
        </p>
        <h2 className="font-display text-lg" id="rattelser">Rättelser</h2>
        <div data-testid="rattelser" className="space-y-2">
          <p>
            Har du hittat ett fel eller en föråldrad uppgift? Rapportera det som ett ärende på{" "}
            <a
              href="https://github.com/Runemark7/swedish-democratic-tracker/issues/new"
              className="text-link underline"
              rel="noopener noreferrer"
            >
              GitHub
            </a>
            . Ange vilken sida det gäller och gärna en länk till källan. Ärendet är öppet för alla att läsa.
          </p>
          <p>
            Beslutar Redaktören att ändra, stryka eller ångra något registreras ändringen i{" "}
            <Link to="/andringslogg" className="text-link underline">Ändringsloggen</Link>.
            Ett ångrat beslut tas inte bort ur loggen. Ångrandet blir en egen post som pekar på det ursprungliga.
          </p>
        </div>
        <p>
          Varje omröstning på ett Löfte har en länk till Riksdagens egen uppgift, så att du kan kontrollera
          den själv. <SourceMarker sourceId="loften-valmanifest" /> Läs mer under{" "}
          <Link to="/data/loften-valmanifest" className="text-link underline">Data</Link> om hur datan
          bearbetas.
        </p>
      </div>
    </div>
  );
}
