import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { HBars } from "@/components/charts";
import type { YearlyHeadcount } from "@/types/democracy";
import { riksdagApi } from "./api";

function Skeleton() {
  return (
    <div className="sdt-page" style={{ padding: "40px 32px" }}>
      {[120, 200, 160, 120].map((h, i) => (
        <div
          key={i}
          style={{
            height: h,
            background: "var(--color-track)",
            borderRadius: 4,
            marginBottom: 16,
          }}
        />
      ))}
    </div>
  );
}

export function AuthorityDetailPage() {
  const { slug } = useParams<{ slug: string }>();

  const { data: authority, isLoading, isError } = useQuery({
    queryKey: ["riksdag-authority", slug],
    queryFn: () => riksdagApi.getAuthority(slug ?? ""),
    enabled: !!slug,
    staleTime: 60_000,
  });

  if (isLoading) return <Skeleton />;

  if (isError || !authority) {
    return (
      <div className="sdt-page" style={{ padding: "40px 32px" }}>
        <Link to="/riksdag" style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-accent)", textDecoration: "none" }}>
          ← Tillbaka till riksdag
        </Link>
        <div style={{ marginTop: 40, fontFamily: "var(--font-body)", fontSize: 16, color: "var(--color-fg-muted)" }}>
          Myndigheten hittades inte.
        </div>
      </div>
    );
  }

  const histMax = authority.history.length > 0
    ? Math.max(...authority.history.map((h) => h.expenditureMdkr))
    : authority.expenditureMdkr;

  return (
    <div className="sdt-page">
      {/* ── Back link ─────────────────────────────────────────────────── */}
      <div style={{ padding: "24px 32px 0" }}>
        <Link
          to="/riksdag"
          style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-accent)", textDecoration: "none", letterSpacing: "0.05em" }}
        >
          ← RIKSDAG
        </Link>
      </div>

      {/* ── Hero ──────────────────────────────────────────────────────── */}
      <div
        style={{
          padding: "20px 32px 28px",
          borderBottom: "1px solid var(--color-border)",
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
            marginBottom: 8,
          }}
        >
          {authority.ministry}
        </div>
        <h1
          style={{
            fontFamily: "var(--font-serif)",
            fontSize: 32,
            fontWeight: 700,
            margin: "0 0 6px",
            color: "var(--color-fg)",
            lineHeight: 1.1,
          }}
        >
          {authority.name}
        </h1>
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 12,
            color: "var(--color-fg-muted)",
            marginBottom: authority.description ? 16 : 0,
          }}
        >
          {authority.role}
        </div>
        {authority.description && (
          <p
            style={{
              fontFamily: "var(--font-body)",
              fontSize: 14,
              color: "var(--color-fg-muted)",
              margin: 0,
              maxWidth: 600,
              lineHeight: 1.6,
            }}
          >
            {authority.description}
          </p>
        )}
      </div>

      {/* ── Main grid ─────────────────────────────────────────────────── */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: 0,
          borderBottom: "1px solid var(--color-border)",
        }}
      >
        {/* Expenditure history */}
        <div
          style={{
            padding: "24px 28px",
            borderRight: "1px solid var(--color-border)",
          }}
        >
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              letterSpacing: "0.15em",
              color: "var(--color-fg-muted)",
              marginBottom: 20,
              textTransform: "uppercase",
            }}
          >
            Kostnadsutveckling {authority.history.length > 0 ? `${authority.history[0].year}–${authority.history[authority.history.length - 1].year}` : authority.year}
          </div>
          {authority.history.length > 0 ? (
            <HBars
              items={authority.history.map((h) => ({
                name: String(h.year),
                value: Math.round(h.expenditureMdkr * 10) / 10,
                color: "#2d6fa8",
              }))}
              max={histMax}
              unit=" mdkr"
              height={5}
              gap={8}
            />
          ) : (
            <div style={{ fontSize: 12, color: "var(--color-fg-muted)", fontFamily: "var(--font-mono)" }}>
              {authority.expenditureMdkr.toFixed(1)} mdkr ({authority.year})
            </div>
          )}
          <div
            style={{
              marginTop: 16,
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              opacity: 0.7,
            }}
          >
            Källa: Statskontoret årsutfall · driftkostnader exkl. transfereringar
          </div>
        </div>

        {/* Employees */}
        <div style={{ padding: "24px 28px" }}>
          <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, letterSpacing: "0.15em", color: "var(--color-fg-muted)", marginBottom: 20, textTransform: "uppercase" }}>
            Anställda
          </div>

          {/* Current headcount hero */}
          <div style={{ display: "flex", alignItems: "baseline", gap: 8, marginBottom: 4 }}>
            <span style={{ fontFamily: "var(--font-serif)", fontSize: 48, fontWeight: 700, color: "var(--color-fg)", lineHeight: 1 }}>
              {authority.headcountInt > 0 ? authority.headcountInt.toLocaleString("sv-SE") : authority.headcount}
            </span>
          </div>
          <div style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-fg-muted)", marginBottom: 20 }}>
            heltidsanställda · {authority.year}
          </div>

          {/* Headcount history chart */}
          {authority.headcountHistory && authority.headcountHistory.length > 1 ? (
            <>
              <div style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", letterSpacing: "1px", marginBottom: 10, textTransform: "uppercase" }}>
                Personalutveckling {authority.headcountHistory[0].year}–{authority.headcountHistory[authority.headcountHistory.length - 1].year}
              </div>
              <HBars
                items={(authority.headcountHistory as YearlyHeadcount[]).map((h) => ({
                  name: String(h.year),
                  value: h.headcountInt,
                  color: "#5a9fd0",
                }))}
                max={Math.max(...authority.headcountHistory.map((h) => h.headcountInt))}
                unit=""
                height={5}
                gap={8}
                formatValue={(v) => v.toLocaleString("sv-SE")}
              />
              <div style={{ marginTop: 10, fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", opacity: 0.7 }}>
                Källa: SCB OE0108 · heltidsanställda årsgenomsnitt
              </div>
            </>
          ) : null}

          {/* Meta */}
          <div style={{ display: "flex", flexDirection: "column", gap: 4, marginTop: 16 }}>
            <div style={{ display: "flex", alignItems: "baseline", gap: 8 }}>
              <span style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)", textTransform: "uppercase", letterSpacing: "0.1em" }}>Departement</span>
              <span style={{ fontFamily: "var(--font-body)", fontSize: 13, color: "var(--color-fg)" }}>{authority.ministry}</span>
            </div>
            <div style={{ display: "flex", alignItems: "baseline", gap: 8 }}>
              <span style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)", textTransform: "uppercase", letterSpacing: "0.1em" }}>Kostnad {authority.year}</span>
              <span style={{ fontFamily: "var(--font-mono)", fontSize: 13, color: "var(--color-fg)" }}>{authority.expenditureMdkr.toFixed(1)} mdkr</span>
            </div>
          </div>
        </div>
      </div>

      {/* ── Kostnad & datakälla ────────────────────────────────────────── */}
      <div
        style={{
          margin: "0 32px",
          border: "1px solid var(--color-border)",
          borderTop: "none",
          background: "var(--color-sdt-surface)",
          padding: "20px 24px",
        }}
      >
        <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, letterSpacing: "0.15em", color: "var(--color-fg-muted)", textTransform: "uppercase", marginBottom: 10 }}>
          Om kostnaderna · Hur beräknas {authority.expenditureMdkr.toFixed(1)} mdkr?
        </div>
        <p style={{ fontFamily: "var(--font-body)", fontSize: 13, color: "var(--color-fg-muted)", margin: "0 0 10px", lineHeight: 1.65, maxWidth: 620 }}>
          Siffran visar <strong style={{ color: "var(--color-fg)" }}>driftkostnader</strong> — vad myndigheten faktiskt kostade att driva under {authority.year}. Det inkluderar löner, lokaler, IT och övriga förvaltningskostnader. Det exkluderar <em>transfereringar</em> (t.ex. bidrag och ersättningar som myndigheten betalar ut till hushåll eller företag).
        </p>
        <p style={{ fontFamily: "var(--font-body)", fontSize: 13, color: "var(--color-fg-muted)", margin: "0 0 12px", lineHeight: 1.65, maxWidth: 620 }}>
          Källa är <strong style={{ color: "var(--color-fg)" }}>Statskontoret årsutfall definitiv</strong> — den slutgiltiga redovisningen av statens faktiska utgifter per anslag. För att verifiera siffran: jämför med det anslag som Riksdagen beslutade i regleringsbrevet nedan. Skillnaden är myndighetens eventuella under- eller överskridande av tilldelat anslag.
        </p>
        <a
          href="https://www.statskontoret.se/psidata/arsutfall"
          target="_blank"
          rel="noopener noreferrer"
          style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: "var(--color-accent)", textDecoration: "none", letterSpacing: "0.05em" }}
        >
          ↗ Öppna Statskontoret årsutfall
        </a>
      </div>

      {/* ── Regleringsbrev ────────────────────────────────────────────── */}
      <div
        style={{
          margin: "0 32px",
          border: "1px solid var(--color-border)",
          borderTop: "none",
          background: "var(--color-sdt-surface)",
          padding: "20px 24px",
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
            marginBottom: 10,
          }}
        >
          Krav från riksdag · Regleringsbrev
        </div>
        <p
          style={{
            fontFamily: "var(--font-body)",
            fontSize: 13,
            color: "var(--color-fg-muted)",
            margin: "0 0 12px",
            lineHeight: 1.6,
            maxWidth: 560,
          }}
        >
          Riksdagen styr myndigheter genom regleringsbrev — årliga beslut som anger uppdrag, mål och ekonomiska ramar. Sök bland gällande och historiska regleringsbrev för {authority.name} i Riksdagens öppna dokumentdatabas.
        </p>
        <a
          href={authority.regleringsbrevUrl}
          target="_blank"
          rel="noopener noreferrer"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-accent)",
            textDecoration: "none",
            letterSpacing: "0.05em",
          }}
        >
          ↗ Sök regleringsbrev för {authority.name}
        </a>
      </div>

      {/* ── Links ─────────────────────────────────────────────────────── */}
      {(authority.websiteUrl || authority.annualReportUrl) && (
        <div
          style={{
            margin: "0 32px 32px",
            border: "1px solid var(--color-border)",
            borderTop: "none",
            background: "var(--color-sdt-surface)",
          }}
        >
          {authority.websiteUrl && (
            <a
              href={authority.websiteUrl}
              target="_blank"
              rel="noopener noreferrer"
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                padding: "14px 24px",
                borderBottom: authority.annualReportUrl ? "1px solid var(--color-border)" : "none",
                fontFamily: "var(--font-mono)",
                fontSize: 12,
                color: "var(--color-fg)",
                textDecoration: "none",
              }}
            >
              <span>Officiell webbplats</span>
              <span style={{ color: "var(--color-accent)" }}>↗</span>
            </a>
          )}
          {authority.annualReportUrl && (
            <a
              href={authority.annualReportUrl}
              target="_blank"
              rel="noopener noreferrer"
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                padding: "14px 24px",
                fontFamily: "var(--font-mono)",
                fontSize: 12,
                color: "var(--color-fg)",
                textDecoration: "none",
              }}
            >
              <span>Årsredovisning</span>
              <span style={{ color: "var(--color-accent)" }}>↗</span>
            </a>
          )}
        </div>
      )}
    </div>
  );
}
