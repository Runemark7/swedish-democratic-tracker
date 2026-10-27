import { Link } from "react-router-dom";
import { useMediaQuery } from "@/hooks/useMediaQuery";

/**
 * Landing for the side product: kommuner and regioner. The routes /kommun
 * and /region stay; this page is the single nav entry that reaches both.
 */
export function LokaltPage() {
  const isMobile = useMediaQuery("(max-width: 640px)");

  const cards = [
    {
      to: "/kommun",
      kicker: "Kommun",
      title: "Kommuner",
      text: "Nyckeltal, utgifter och befolkning för de 290 kommunerna.",
    },
    {
      to: "/region",
      kicker: "Region",
      title: "Regioner",
      text: "Nyckeltal och budget för de 21 regionerna.",
    },
  ];

  return (
    <div className="sdt-page">
      <header
        style={{
          padding: isMobile ? "20px 14px 16px" : "8px 0 24px",
          display: "flex",
          flexDirection: "column",
          gap: 8,
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "2px",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
          }}
        >
          Lokalt
        </div>
        <h1
          style={{
            fontFamily: "var(--font-serif)",
            fontSize: isMobile ? 28 : 40,
            fontWeight: 400,
            letterSpacing: "-0.8px",
            color: "var(--color-fg)",
            margin: 0,
            lineHeight: 1.1,
          }}
        >
          Kommuner och regioner
        </h1>
        <p
          style={{
            fontFamily: "var(--font-body)",
            fontSize: 14,
            color: "var(--color-fg-muted)",
            margin: 0,
            maxWidth: 520,
            lineHeight: 1.6,
          }}
        >
          Siffror för kommun och region, vid sidan av vad riksdagen beslutar.
        </p>
      </header>

      <div
        style={{
          margin: isMobile ? "0 14px 24px" : "0 0 24px",
          display: "grid",
          gridTemplateColumns: isMobile ? "1fr" : "1fr 1fr",
          gap: 1,
          background: "var(--color-border)",
          border: "1px solid var(--color-border)",
        }}
      >
        {cards.map((card) => (
          <Link
            key={card.to}
            to={card.to}
            style={{
              background: "var(--color-sdt-surface)",
              padding: isMobile ? "18px 16px" : "24px 22px",
              textDecoration: "none",
              color: "inherit",
              display: "flex",
              flexDirection: "column",
              gap: 8,
              minWidth: 0,
            }}
          >
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.14em",
                textTransform: "uppercase",
                color: "var(--color-fg-muted)",
              }}
            >
              {card.kicker}
            </span>
            <span
              style={{
                fontFamily: "var(--font-serif)",
                fontSize: 26,
                color: "var(--color-fg)",
              }}
            >
              {card.title}
            </span>
            <span
              style={{
                fontFamily: "var(--font-body)",
                fontSize: 13,
                color: "var(--color-fg-muted)",
                lineHeight: 1.5,
              }}
            >
              {card.text}
            </span>
          </Link>
        ))}
      </div>
    </div>
  );
}
