import { useState } from "react";
import { useParams, useSearchParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { votesApi } from "./api";
import { PARTY_COLORS, committeeFromBeteckning } from "@/shared/design";
import { useSpeechesByDocument } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
import type { components } from "@/shared/api-contract";

type PartyVotePosition = components["schemas"]["PartyVotePosition"];

function committeeLabel(beteckning: string): string {
  return committeeFromBeteckning(beteckning) ?? beteckning;
}

/** Riksdagen's outcome words. Anything else (including a missing value) is not an outcome. */
function knownOutcome(value: string | null | undefined): string {
  const text = value?.trim() ?? "";
  if (text === "Bifall" || text === "Avslag" || text === "Återremiss") return text;
  return "";
}

// Parse a Riksdagen "YYYY-MM-DD HH:MM:SS" datetime to a Date (date part only).
// Returns null for empty/unparseable input.
function parseRiksdagDate(raw?: string): Date | null {
  if (!raw) return null;
  const datePart = raw.slice(0, 10);
  const d = new Date(`${datePart}T00:00:00`);
  return isNaN(d.getTime()) ? null : d;
}

const SV_DATE = new Intl.DateTimeFormat("sv-SE", {
  day: "numeric",
  month: "long",
  year: "numeric",
});

// Label for the empty Debatten section, derived purely from sourced dates.
// Never asserts "no debate happened" — states the date + that no anföranden
// are registered, leaving interpretation to the reader.
function debattFallbackLabel(debattDate?: string): string {
  const d = parseRiksdagDate(debattDate);
  if (!d) return "Inget debattdatum registrerat";
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  // Today counts as planned/ongoing — a debate dated today may still lie ahead,
  // so we don't assert it already happened until the date is strictly past.
  if (d.getTime() >= today.getTime()) {
    return `Debatt planerad: ${SV_DATE.format(d)}`;
  }
  return `Debatten hölls ${SV_DATE.format(d)} · inga anföranden registrerade här`;
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
  // Beteckning numbering restarts each riksmöte. Home links pass `rm`.
  const riksmote = searchParams.get("rm") ?? "";

  // The vote API is one förslagspunkt at a time. Punkt 1 is what this page
  // can show without a new "list every punkt" read; the heading says so.
  const punkt = "1";

  const { data, isLoading, isError } = useQuery({
    queryKey: ["beslut", beteckning, riksmote, punkt],
    queryFn: () => votesApi.getDetail(beteckning!, punkt, riksmote || undefined),
    enabled: !!beteckning,
    retry: false,
  });

  const dokIdForSpeeches = data?.dokId;
  const { data: speeches } = useSpeechesByDocument(dokIdForSpeeches);
  const [expandSpeakers, setExpandSpeakers] = useState(false);

  const title    = data?.documentTitle ?? fallbackTitle;
  const tallySession = data?.session ?? "";
  // The link's riksmöte wins in the header. A tally from another year is
  // named, not silently presented as the betänkande the reader opened.
  const session = riksmote || tallySession;
  const sessionMismatch = Boolean(riksmote && tallySession && riksmote !== tallySession);
  const committee = committeeLabel(beteckning ?? "");
  const dokId    = data?.dokId;

  // Never invent Bifall. Prefer Riksdagen's own outcome word, then the
  // ballot tally for this punkt, then a status the linker already knew.
  // While the request is in flight, data is absent, so the query-string
  // hint can show; once data has arrived, an empty tally stays empty.
  const statusText = data
    ? knownOutcome(data.statusText) || knownOutcome(data.status)
    : knownOutcome(fallbackStatus);
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

        {sessionMismatch && (
          <p
            style={{
              fontFamily: "var(--font-body)",
              fontSize: 13,
              color: "var(--color-fg)",
              margin: "0 0 16px",
              lineHeight: 1.5,
            }}
          >
            Länken gäller riksmötet {riksmote}. Svaret som hämtades är från {tallySession}.
            Beteckningen {beteckning} finns i båda, och den här läsningen avgränsades inte
            till riksmötet i länken.
          </p>
        )}

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
              color: statusText ? statusColor : "var(--color-fg-muted)",
            }}
          >
            {statusText || (isLoading ? "Hämtar utfall…" : "Utfall ej publicerat")}
          </span>
        </div>

        {/* ── SAMMANFATTNING ──────────────────────────────────────── */}
        <section
          style={{
            background: "var(--color-sdt-surface)",
            border: "1px solid var(--color-border)",
            borderRadius: 4,
            padding: "16px 20px",
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
              marginBottom: 8,
            }}
          >
            Sammanfattning
          </div>
          {data?.notis ? (
            // Riksdagen "Beslut i korthet" is trusted government HTML. Render it
            // flush inside the card (matching the summary <p>) — not via
            // DocumentBody, which adds its own surface/border and would nest a
            // card inside this card.
            <div
              className="riksdagen-doc"
              style={{
                fontFamily: "var(--font-body)",
                fontSize: 14,
                lineHeight: 1.6,
                color: "var(--color-fg)",
              }}
              dangerouslySetInnerHTML={{ __html: data.notis }}
            />
          ) : data?.summary ? (
            <p
              style={{
                fontFamily: "var(--font-body)",
                fontSize: 14,
                lineHeight: 1.6,
                color: "var(--color-fg)",
                margin: 0,
              }}
            >
              {data.summary}
            </p>
          ) : (
            <p
              style={{
                fontFamily: "var(--font-body)",
                fontSize: 13,
                fontStyle: "italic",
                color: "var(--color-fg-muted)",
                margin: 0,
              }}
            >
              Riksdagen har inte publicerat någon sammanfattning för det här beslutet.
            </p>
          )}
        </section>

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
              marginBottom: 6,
            }}
          >
            Partiernas röster · förslagspunkt {punkt}
          </div>
          <p
            style={{
              fontFamily: "var(--font-body)",
              fontSize: 12,
              color: "var(--color-fg-muted)",
              margin: "0 0 16px",
              lineHeight: 1.5,
            }}
          >
            Siffrorna gäller förslagspunkt {punkt}
            {(tallySession || session) ? ` i riksmötet ${tallySession || session}` : ""}. Ett betänkande kan ha fler punkter.
          </p>

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

        {/* ── DEBATTEN — speakers ─────────────────────────────────── */}
        <section
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
              marginBottom: 12,
            }}
          >
            Debatten {speeches && speeches.length > 0 ? `· ${speeches.length} talare` : ""}
          </div>
          {(!speeches || speeches.length === 0) && (
            <p
              style={{
                fontFamily: "var(--font-body)",
                fontStyle: "italic",
                fontSize: 13,
                color: "var(--color-fg-muted)",
                margin: 0,
              }}
            >
              {debattFallbackLabel(data?.debattDate)}
            </p>
          )}
          {speeches && speeches.length > 0 && (
            <>
              {(expandSpeakers ? speeches : speeches.slice(0, 5)).map((s) => (
                <SpeechRow key={s.id} speech={s} />
              ))}
              {speeches.length > 5 && !expandSpeakers && (
                <button
                  onClick={() => setExpandSpeakers(true)}
                  style={{
                    marginTop: 12,
                    fontFamily: "var(--font-mono)",
                    fontSize: 11,
                    color: "var(--color-accent)",
                    background: "none",
                    border: "none",
                    cursor: "pointer",
                    letterSpacing: "0.05em",
                    padding: 0,
                  }}
                >
                  Visa alla {speeches.length} anföranden →
                </button>
              )}
            </>
          )}
        </section>

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
