import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
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

  const hasHeadcount = (authority.headcountHistory?.length ?? 0) > 1;
  const hcByYear = new Map((authority.headcountHistory ?? []).map((h) => [h.year, h.headcountInt]));

  const mergedHistory = authority.history.map((h) => ({
    year: h.year,
    expenditureMdkr: h.expenditureMdkr,
    headcountInt: hcByYear.get(h.year) ?? null,
  }));

  const maxCost = mergedHistory.length > 0
    ? Math.max(...mergedHistory.map((h) => h.expenditureMdkr))
    : authority.expenditureMdkr;

  const maxHead = hasHeadcount
    ? Math.max(...mergedHistory.map((h) => h.headcountInt ?? 0))
    : 0;

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

      {/* ── Combined historik ─────────────────────────────────────────── */}
      <div style={{ padding: "24px 32px", borderBottom: "1px solid var(--color-border)" }}>
        {/* Header row */}
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", marginBottom: 20 }}>
          <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, letterSpacing: "0.15em", color: "var(--color-fg-muted)", textTransform: "uppercase" }}>
            Historik {mergedHistory.length > 0 ? `${mergedHistory[0].year}–${mergedHistory[mergedHistory.length - 1].year}` : authority.year}
          </div>
          {/* Current stats */}
          <div style={{ display: "flex", gap: 24 }}>
            <div style={{ textAlign: "right" }}>
              <div style={{ fontFamily: "var(--font-mono)", fontSize: 18, fontWeight: 600, color: "var(--color-fg)" }}>
                {authority.expenditureMdkr.toFixed(1)} mdkr
              </div>
              <div style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", letterSpacing: "0.05em" }}>
                kostnad {authority.year}
              </div>
            </div>
            {authority.headcountInt > 0 && (
              <div style={{ textAlign: "right" }}>
                <div style={{ fontFamily: "var(--font-mono)", fontSize: 18, fontWeight: 600, color: "var(--color-fg)" }}>
                  {authority.headcountInt.toLocaleString("sv-SE")}
                </div>
                <div style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", letterSpacing: "0.05em" }}>
                  anställda {authority.year}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Year rows */}
        <div style={{ display: "flex", flexDirection: "column", gap: 7 }}>
          {mergedHistory.map((row) => (
            <div key={row.year} style={{ display: "grid", gridTemplateColumns: `3rem 1fr${hasHeadcount ? " 1fr" : ""}`, gap: "0 16px", alignItems: "center" }}>
              <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)" }}>{row.year}</div>

              {/* Cost bar */}
              <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                <div style={{ flex: 1, height: 5, background: "var(--color-track)", borderRadius: 2, overflow: "hidden" }}>
                  <div style={{ height: "100%", width: `${(row.expenditureMdkr / maxCost) * 100}%`, background: "#2d6fa8", borderRadius: 2 }} />
                </div>
                <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg)", minWidth: "5.5rem", textAlign: "right" }}>
                  {row.expenditureMdkr.toFixed(1)} mdkr
                </div>
              </div>

              {/* Headcount bar */}
              {hasHeadcount && (
                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                  <div style={{ flex: 1, height: 5, background: "var(--color-track)", borderRadius: 2, overflow: "hidden" }}>
                    <div style={{ height: "100%", width: row.headcountInt ? `${(row.headcountInt / maxHead) * 100}%` : "0%", background: "#5a8f6e", borderRadius: 2 }} />
                  </div>
                  <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg)", minWidth: "4.5rem", textAlign: "right" }}>
                    {row.headcountInt ? row.headcountInt.toLocaleString("sv-SE") : "–"}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Legend */}
        <div style={{ display: "flex", gap: 20, marginTop: 14 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 5 }}>
            <div style={{ width: 12, height: 4, background: "#2d6fa8", borderRadius: 2 }} />
            <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", opacity: 0.8 }}>Kostnad · Statskontoret årsutfall</span>
          </div>
          {hasHeadcount && (
            <div style={{ display: "flex", alignItems: "center", gap: 5 }}>
              <div style={{ width: 12, height: 4, background: "#5a8f6e", borderRadius: 2 }} />
              <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", opacity: 0.8 }}>Anställda dec. · SCB KLS AM0102</span>
            </div>
          )}
          <div style={{ display: "flex", alignItems: "center", gap: 5, marginLeft: "auto" }}>
            <div style={{ display: "flex", alignItems: "baseline", gap: 6 }}>
              <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg-muted)", textTransform: "uppercase", letterSpacing: "0.1em" }}>Departement</span>
              <span style={{ fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--color-fg)" }}>{authority.ministry}</span>
            </div>
          </div>
        </div>
      </div>

      {/* ── Budget vs Utfall ──────────────────────────────────────────── */}
      {authority.history.length > 0 && authority.history.some(h => (h.budgetMdkr ?? 0) > 0) && (() => {
        type RowWithBudget = { year: number; expenditureMdkr: number; budgetMdkr: number };
        const rows = ([...authority.history].reverse() as RowWithBudget[]).filter(h => h.budgetMdkr > 0);
        const deviations = rows
          .map(h => (h.expenditureMdkr - h.budgetMdkr) / h.budgetMdkr * 100);
        const avgDev = deviations.reduce((s, d) => s + d, 0) / deviations.length;
        const devColor = (d: number) => Math.abs(d) <= 2 ? "var(--color-fg-muted)" : d > 0 ? "#d0533f" : "#2d9e6b";
        return (
          <div style={{ margin: "0 32px", border: "1px solid var(--color-border)", borderTop: "none", background: "var(--color-sdt-surface)", padding: "20px 24px" }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", marginBottom: 16 }}>
              <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, letterSpacing: "0.15em", color: "var(--color-fg-muted)", textTransform: "uppercase" }}>
                Budget vs Utfall · Anslag jämfört med faktiska kostnader
              </div>
              <div style={{ fontFamily: "var(--font-mono)", fontSize: 11, color: devColor(avgDev) }}>
                Genomsnitt {avgDev > 0 ? "+" : ""}{avgDev.toFixed(1)}% {Math.abs(avgDev) <= 2 ? "· Håller budgeten" : avgDev > 0 ? "· Överskrider budget" : "· Under budget"}
              </div>
            </div>
            <table style={{ width: "100%", borderCollapse: "collapse", fontFamily: "var(--font-mono)", fontSize: 12 }}>
              <thead>
                <tr style={{ borderBottom: "1px solid var(--color-border)" }}>
                  {["År", "Anslag", "Utfall", "Avvikelse"].map(h => (
                    <th key={h} style={{ textAlign: h === "År" ? "left" : "right", padding: "4px 8px", fontSize: 9, letterSpacing: "1px", textTransform: "uppercase", color: "var(--color-fg-muted)", fontWeight: 400 }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rows.map(h => {
                  const dev = (h.expenditureMdkr - h.budgetMdkr) / h.budgetMdkr * 100;
                  const color = devColor(dev);
                  return (
                    <tr key={h.year} style={{ borderBottom: "1px solid var(--color-border)" }}>
                      <td style={{ padding: "6px 8px", color: "var(--color-fg-muted)" }}>{h.year}</td>
                      <td style={{ padding: "6px 8px", textAlign: "right", color: "var(--color-fg-muted)" }}>{h.budgetMdkr.toFixed(2)} mdkr</td>
                      <td style={{ padding: "6px 8px", textAlign: "right", color: "var(--color-fg)" }}>{h.expenditureMdkr.toFixed(2)} mdkr</td>
                      <td style={{ padding: "6px 8px", textAlign: "right", color, fontWeight: Math.abs(dev) > 5 ? 600 : 400 }}>
                        {dev > 0 ? "+" : ""}{dev.toFixed(1)}%
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            <div style={{ marginTop: 10, fontSize: 9, fontFamily: "var(--font-mono)", color: "var(--color-fg-muted)", opacity: 0.7 }}>
              Anslag = Statens budget + Ändringsbudgetar · Källa: Statskontoret årsutfall definitiv
            </div>
          </div>
        );
      })()}

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
