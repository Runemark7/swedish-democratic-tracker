import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { HBars } from "@/components/charts";
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
            Anställda
          </div>
          <div
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: 48,
              fontWeight: 700,
              color: "var(--color-fg)",
              lineHeight: 1,
              marginBottom: 6,
            }}
          >
            {authority.headcountInt > 0
              ? authority.headcountInt.toLocaleString("sv-SE")
              : authority.headcount}
          </div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg-muted)",
              marginBottom: 16,
            }}
          >
            heltidsanställda · {authority.year}
          </div>
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: 4,
              marginTop: 8,
            }}
          >
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
