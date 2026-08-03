import { Link } from "react-router-dom";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { useRecordCoverage } from "@/hooks/useDemocracy";
import { SOURCES } from "@/components/sources/SourceRegistry";

/** A section with a mono eyebrow, matching /data. */
function Section({ eyebrow, title, children }: {
  eyebrow: string;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section style={{ marginBottom: 40, maxWidth: 640 }}>
      <div
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 10,
          letterSpacing: "2px",
          color: "var(--color-fg-muted)",
          textTransform: "uppercase",
          marginBottom: 6,
        }}
      >
        {eyebrow}
      </div>
      <h2
        style={{
          fontFamily: "var(--font-serif)",
          fontSize: 22,
          fontWeight: 400,
          margin: "0 0 10px",
          lineHeight: 1.2,
        }}
      >
        {title}
      </h2>
      <div style={{ fontSize: 14, lineHeight: 1.65, color: "var(--color-fg)" }}>
        {children}
      </div>
    </section>
  );
}

/**
 * Om sajten — the method, stated plainly.
 *
 * For a site built for readers who distrust the institutions that report on
 * politics, the method IS the product. Every claim here is one the reader can
 * check, and the page says what the site cannot do as prominently as what it
 * can. Nothing on this page asks to be believed on trust.
 */
export function OmSajtenPage() {
  const isMobile = useMediaQuery("(max-width: 640px)");
  const { data: coverage } = useRecordCoverage();
  const sourceCount = Object.keys(SOURCES).length;

  const pad = isMobile ? "20px 14px" : "40px 32px";

  return (
    <div className="sdt-page">
      <header style={{ padding: isMobile ? "20px 14px 16px" : "40px 32px 24px" }}>
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "2px",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
            marginBottom: 8,
          }}
        >
          OM SAJTEN
        </div>
        <h1
          style={{
            fontFamily: "var(--font-serif)",
            fontSize: isMobile ? 28 : 44,
            margin: 0,
            fontWeight: 400,
            lineHeight: 1.05,
          }}
        >
          Hur sajten fungerar — och vad den inte gör
        </h1>
        <p style={{ fontSize: 13, color: "var(--color-fg-muted)", margin: "8px 0 0", maxWidth: 560 }}>
          Du ska inte behöva lita på den här sajten. Du ska kunna kontrollera den.
        </p>
      </header>

      <main style={{ padding: pad }}>
        <Section eyebrow="Syfte" title="Vad det här är">
          <p style={{ margin: "0 0 10px" }}>
            Sverige publicerar nästan allt staten gör — varje omröstning, varje
            proposition. Materialet är offentligt, men det är skrivet på ett
            språk och i en mängd som gör det praktiskt oläsbart för den som inte
            har det som yrke.
          </p>
          <p style={{ margin: 0 }}>
            Riksdagskollen försöker inte tolka det åt dig. Den lägger partiets
            egna uttalade mål bredvid hur partiet faktiskt röstade, med källa på
            båda, och låter dig dra slutsatsen själv.
          </p>
        </Section>

        <Section eyebrow="Gränser" title="Vad sajten aldrig gör">
          <ul style={{ margin: 0, paddingLeft: 18, listStyle: "disc" }}>
            <li style={{ marginBottom: 6 }}>
              <strong>Inga betyg.</strong> Det finns inget &rdquo;populismindex&rdquo;,
              ingen trovärdighetspoäng, ingen rankning av partier. Ett sådant tal
              vore en åsikt förklädd till mätning.
            </li>
            <li style={{ marginBottom: 6 }}>
              <strong>Inga omdömen.</strong> Sajten skriver aldrig att ett parti
              &rdquo;bröt sitt löfte&rdquo;. Den visar löftet, visar rösten, och
              slutar där.
            </li>
            <li style={{ marginBottom: 6 }}>
              <strong>Inget urval.</strong> Sajten plockar inte ut vilka
              omröstningar som ska synas. Hela mandatperiodens röstning finns
              med — urvalet är där partiskhet smyger sig in.
            </li>
            <li>
              <strong>Ingen färgkodning av rätt och fel.</strong> Siffror visas i
              neutral ton. Talet är fakta; värderingen är din.
            </li>
          </ul>
        </Section>

        <Section eyebrow="Metod" title="Varför sajten inte räknar ut någon procent">
          <p style={{ margin: "0 0 10px" }}>
            Sajten visade tidigare en siffra av typen &rdquo;4 av 5 relevanta
            röster i linje&rdquo;. Den är borttagen, och det är värt att förklara
            varför — för det säger något om hur sajten är tänkt att fungera.
          </p>
          <p style={{ margin: "0 0 10px" }}>
            För att räkna ut en sådan siffra måste man veta åt vilket håll ett ja
            respektive ett nej pekar i varje omröstning. Sajten härledde det ur
            vem som hade lagt förslaget. Det blev systematiskt fel.
          </p>
          <p style={{ margin: "0 0 10px" }}>
            Ett oppositionsparti lägger en motion. Utskottet föreslår att den
            avslås. Omröstningen gäller utskottets förslag — så partiet röstar
            nej, och härledningen, som väntade sig ja, bokförde det som att
            partiet svek sitt eget löfte. Mätt över hela mandatperioden gav det{" "}
            <strong>68,6 % för regeringspartierna mot 43,2 % för oppositionen</strong>
            {" "}— en skillnad som inte mätte löftesuppfyllelse alls, utan bara
            hur ofta en grov tumregel råkade stämma.
          </p>
          <p style={{ margin: 0, color: "var(--color-fg-muted)", fontSize: 13 }}>
            Siffran gick inte att justera fram till rätthet: riktningen beror på
            vad ett betänkande faktiskt föreslår, vilket inte går att härleda ur
            vem som lagt förslaget. Därför räknar sajten inte längre ut något
            tal. Löftet och rösterna visas bredvid varandra, med källa på båda,
            och du drar slutsatsen.
          </p>
        </Section>

        <Section eyebrow="Täckning" title="Vad sajten har — och vad som saknas">
          {coverage ? (
            <>
              <p style={{ margin: "0 0 10px" }}>
                Sajten visar {coverage.mandate.label.toLowerCase()} och innehåller{" "}
                <strong>
                  {coverage.ingested.toLocaleString("sv-SE")} av{" "}
                  {coverage.expected.toLocaleString("sv-SE")}
                </strong>{" "}
                omröstningar. Nämnaren är Riksdagens egen räkning, inte vår.
              </p>
              {coverage.ingested < coverage.expected && (
                <p style={{ margin: "0 0 10px" }}>
                  De {coverage.expected - coverage.ingested} som fattas är kända
                  och redovisade: någon enstaka omröstning gäller en motion utan
                  betänkandebeteckning och går inte att hämta, och för ett par
                  betänkanden publicerar Riksdagen ingen röstdata alls.
                </p>
              )}
              <p style={{ margin: 0, fontSize: 13, color: "var(--color-fg-muted)" }}>
                Uppdelning per riksmöte finns under{" "}
                <Link to="/data" style={{ color: "var(--color-accent)" }}>
                  Datakällor
                </Link>
                .
              </p>
            </>
          ) : (
            <p style={{ margin: 0 }}>
              Täckningen redovisas på{" "}
              <Link to="/data" style={{ color: "var(--color-accent)" }}>
                Datakällor
              </Link>
              .
            </p>
          )}
        </Section>

        <Section eyebrow="Källor" title="Hur du kontrollerar en siffra">
          <p style={{ margin: "0 0 10px" }}>
            Varje värde på sajten bär en källmarkering. Klicka på den så ser du
            varifrån talet kommer och när det senast verifierades.
          </p>
          <p style={{ margin: 0 }}>
            Alla {sourceCount} källor listas på{" "}
            <Link to="/data" style={{ color: "var(--color-accent)" }}>
              Datakällor
            </Link>
            , med länk till originalet hos Riksdagen, SCB, Kolada eller
            Statskontoret. Hittar du ett fel är det källan som gäller, inte
            sajten.
          </p>
        </Section>

        <Section eyebrow="Avsändare" title="Vem ligger bakom">
          <p style={{ margin: "0 0 10px" }}>
            Riksdagskollen byggs och underhålls av Alexander Runemark som ett
            privat projekt.
          </p>
          <ul style={{ margin: "0 0 10px", paddingLeft: 18, listStyle: "disc" }}>
            <li style={{ marginBottom: 6 }}>Ingen finansiering.</li>
            <li style={{ marginBottom: 6 }}>
              Ingen koppling till något parti, någon myndighet eller något
              mediehus.
            </li>
            <li>Inga annonser.</li>
          </ul>
          <p style={{ margin: 0, fontSize: 13, color: "var(--color-fg-muted)" }}>
            Besöksstatistik samlas in med en egen-hostad, kakfri installation av
            Umami. Den räknar sidvisningar, inte personer, och inga uppgifter
            lämnas till tredje part.
          </p>
        </Section>
      </main>
    </div>
  );
}
