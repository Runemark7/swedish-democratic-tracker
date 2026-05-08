import { useParams, useSearchParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { votesApi } from "./api";
import { PARTY_COLORS } from "@/shared/design";
import type { PartyVotePosition } from "@/shared/types";

// Derive a human-readable committee name from the beteckning prefix.
function committeeLabel(beteckning: string): string {
  const labels: Record<string, string> = {
    SoU: "Socialutskottet",
    TU:  "Trafikutskottet",
    UbU: "Utbildningsutskottet",
    CU:  "Civilutskottet",
    FiU: "Finansutskottet",
    JuU: "Justitieutskottet",
    SkU: "Skatteutskottet",
    FöU: "Försvarsutskottet",
    AU:  "Arbetsmarknadsutskottet",
    MJU: "Miljö- och jordbruksutskottet",
    KU:  "Konstitutionsutskottet",
    NU:  "Näringsutskottet",
    SfU: "Socialförsäkringsutskottet",
    KrU: "Kulturutskottet",
    UU:  "Utrikesutskottet",
  };
  const prefix = beteckning.replace(/\d.*$/, "").replace(/[^A-Za-zÅÄÖåäö]/g, "");
  return labels[prefix] ?? prefix;
}

function VoteBar({ pos }: { pos: PartyVotePosition }) {
  const total = pos.jaCount + pos.nejCount + pos.avstarCount + pos.franvarandeCount;
  const pctJa   = total > 0 ? (pos.jaCount  / total) * 100 : 0;
  const pctNej  = total > 0 ? (pos.nejCount / total) * 100 : 0;
  const style = PARTY_COLORS[pos.party];
  const barColor = style?.bg ?? "var(--color-accent)";

  return (
    <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
      <span
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 11,
          fontWeight: 700,
          color: style?.bg ?? "var(--color-fg)",
          minWidth: 28,
          letterSpacing: "0.05em",
        }}
      >
        {pos.party}
      </span>

      {/* stacked bar */}
      <div
        style={{
          flex: 1,
          height: 8,
          background: "var(--color-track)",
          borderRadius: 4,
          overflow: "hidden",
          display: "flex",
        }}
      >
        <div style={{ width: `${pctJa}%`, background: barColor, transition: "width 0.3s" }} />
        <div style={{ width: `${pctNej}%`, background: "var(--color-fg-muted)", opacity: 0.35, transition: "width 0.3s" }} />
      </div>

      <span
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 10,
          color: "var(--color-fg-muted)",
          minWidth: 80,
          textAlign: "right",
        }}
      >
        {pos.jaCount > 0 && <span style={{ color: barColor }}>{pos.jaCount} Ja</span>}
        {pos.jaCount > 0 && pos.nejCount > 0 && " · "}
        {pos.nejCount > 0 && <span>{pos.nejCount} Nej</span>}
        {pos.avstarCount > 0 && ` · ${pos.avstarCount} Av`}
      </span>
    </div>
  );
}

function DocumentBody({ html }: { html: string }) {
  return (
    <div
      style={{
        background: "var(--color-sdt-surface)",
        border: "1px solid var(--color-border)",
        borderRadius: 4,
        padding: "24px 28px",
        marginBottom: 16,
        fontFamily: "var(--font-body)",
        fontSize: 14,
        lineHeight: 1.7,
        color: "var(--color-fg)",
        // Force readable colors over Riksdagen's bundled <style> block.
        // The HTML ships with inline width/colors tuned for their site;
        // we override via wrapper rules.
        overflowX: "auto",
      }}
      className="riksdagen-doc"
      // Riksdagen serves trusted government markup; sanitisation overhead
      // is overkill here. Wrapper class lets us scope CSS overrides.
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}

export function BeslutDetailPage() {
  const { beteckning } = useParams<{ beteckning: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const fallbackTitle  = searchParams.get("title") ?? beteckning ?? "";
  const fallbackStatus = searchParams.get("status") ?? "";
  const fallbackTag    = searchParams.get("tag") ?? "";
  const fallbackTime   = searchParams.get("time") ?? "";

  const { data, isLoading, isError } = useQuery({
    queryKey: ["beslut", beteckning],
    queryFn: () => votesApi.getDetail(beteckning!, "1"),
    enabled: !!beteckning,
    retry: false,
  });

  const title    = data?.documentTitle ?? fallbackTitle;
  const session  = data?.session ?? "";
  const committee = committeeLabel(beteckning ?? "");
  const dokId    = data?.dokId;

  const statusText = data?.status ?? fallbackStatus ?? "Bifall";
  const beslutDate = data?.date ?? fallbackTime;
  const statusColor =
    statusText === "Bifall"    ? "#16a34a" :
    statusText === "Avslag"    ? "#dc2626" :
    statusText === "Återremiss"? "#d97706" : "var(--color-fg-muted)";

  return (
    <div
      style={{
        background: "var(--color-bg)",
        color: "var(--color-fg)",
        minHeight: "100vh",
        fontFamily: "var(--font-body)",
        paddingBottom: 64,
      }}
    >
      <div style={{ maxWidth: 780, margin: "0 auto", padding: "40px 32px 0" }}>

        {/* ── Back ─────────────────────────────────────────────────── */}
        <button
          onClick={() => navigate(-1)}
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            letterSpacing: "0.1em",
            color: "var(--color-fg-muted)",
            background: "none",
            border: "none",
            cursor: "pointer",
            padding: 0,
            marginBottom: 32,
            textTransform: "uppercase",
          }}
        >
          ← Tillbaka
        </button>

        {/* ── Meta row ─────────────────────────────────────────────── */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 12,
            marginBottom: 12,
            flexWrap: "wrap",
          }}
        >
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "0.15em",
              color: "var(--color-fg-muted)",
              textTransform: "uppercase",
            }}
          >
            Betänkande
          </span>
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              color: "var(--color-accent)",
              background: "color-mix(in srgb, var(--color-accent) 10%, transparent)",
              borderRadius: 3,
              padding: "2px 8px",
              letterSpacing: "0.05em",
            }}
          >
            {beteckning}
          </span>
          {session && (
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                color: "var(--color-fg-muted)",
                letterSpacing: "0.1em",
              }}
            >
              · {session}
            </span>
          )}
          {fallbackTag && (
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                color: "var(--color-fg-muted)",
                background: "var(--color-track)",
                borderRadius: 3,
                padding: "2px 8px",
              }}
            >
              {fallbackTag}
            </span>
          )}
        </div>

        {/* ── Title ────────────────────────────────────────────────── */}
        <h1
          style={{
            fontFamily: "var(--font-serif)",
            fontSize: 36,
            fontWeight: 400,
            letterSpacing: "-0.5px",
            lineHeight: 1.15,
            margin: "0 0 8px",
            color: "var(--color-fg)",
          }}
        >
          {title}
        </h1>
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 12,
            color: "var(--color-fg-muted)",
            marginBottom: data?.subtitle ? 8 : 32,
          }}
        >
          {committee}{beslutDate ? ` · ${beslutDate}` : ""}
        </div>

        {data?.subtitle && (
          <div style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-fg-muted)", marginBottom: 32, letterSpacing: "0.05em" }}>
            {data.subtitle}
          </div>
        )}

        {/* ── Outcome strip ────────────────────────────────────────── */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 16,
            padding: "16px 20px",
            background: "var(--color-sdt-surface)",
            border: "1px solid var(--color-border)",
            borderRadius: 4,
            marginBottom: 32,
          }}
        >
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "0.15em",
              textTransform: "uppercase",
              color: "var(--color-fg-muted)",
            }}
          >
            Utfall
          </span>
          <span
            style={{
              fontFamily: "var(--font-serif)",
              fontStyle: "italic",
              fontSize: 20,
              color: statusColor,
            }}
          >
            {statusText}
          </span>
        </div>

        {data?.summary && (
          <div style={{
            background: "var(--color-sdt-surface)",
            border: "1px solid var(--color-border)",
            borderRadius: 4,
            padding: "16px 20px",
            marginBottom: 16,
            fontFamily: "var(--font-body)",
            fontSize: 14,
            lineHeight: 1.6,
            color: "var(--color-fg-muted)",
          }}>
            {data.summary}
          </div>
        )}

        {/* ── Party breakdown ──────────────────────────────────────── */}
        <div
          style={{
            background: "var(--color-sdt-surface)",
            border: "1px solid var(--color-border)",
            borderRadius: 4,
            padding: "20px 24px",
            marginBottom: 16,
          }}
        >
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "0.15em",
              textTransform: "uppercase",
              color: "var(--color-fg-muted)",
              marginBottom: 16,
            }}
          >
            Partiernas röster
          </div>

          {isLoading && (
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {[1, 2, 3, 4, 5].map((i) => (
                <div
                  key={i}
                  style={{
                    height: 16,
                    background: "var(--color-track)",
                    borderRadius: 4,
                    animation: "d3pulse 2s ease infinite",
                  }}
                />
              ))}
            </div>
          )}

          {isError && (
            <p
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 12,
                color: "var(--color-fg-muted)",
                margin: 0,
                opacity: 0.6,
              }}
            >
              Kunde inte hämta röstdata.
            </p>
          )}

          {data && data.partyBreakdown.length === 0 && (
            <p
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 12,
                color: "var(--color-fg-muted)",
                margin: 0,
                opacity: 0.6,
              }}
            >
              Fullständig röstdata ej tillgänglig för detta beslut.
            </p>
          )}

          {data && data.partyBreakdown.length > 0 && (
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {data.partyBreakdown.map((pos) => (
                <VoteBar key={pos.party} pos={pos} />
              ))}
            </div>
          )}
        </div>

        {/* ── Riksdagen link ───────────────────────────────────────── */}
        {dokId && (
          <a
            href={`https://www.riksdagen.se/sv/dokument-och-lagar/dokument/betankande/_${dokId}/`}
            target="_blank"
            rel="noopener noreferrer"
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 6,
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              letterSpacing: "0.1em",
              color: "var(--color-accent)",
              textDecoration: "none",
              marginBottom: 40,
            }}
          >
            Läs debatt på riksdagen.se →
          </a>
        )}

        {/* ── Document body (proposal, motivation, debate transcript) ── */}
        {data?.bodyHtml && (
          <div
            style={{
              borderTop: "1px solid var(--color-border)",
              paddingTop: 32,
              display: "flex",
              flexDirection: "column",
              gap: 12,
            }}
          >
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.15em",
                textTransform: "uppercase",
                color: "var(--color-fg-muted)",
                marginBottom: 4,
              }}
            >
              Hela betänkandet
            </div>
            <DocumentBody html={data.bodyHtml} />
          </div>
        )}

      </div>
    </div>
  );
}
