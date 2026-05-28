import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useKpiRanking, useKommunList } from "@/hooks/useDemocracy";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { BarsWithMean } from "@/features/budget/components/BarsWithMean";
import { STRIP_KPI_META } from "./kpiMeta";

const INITIAL_LIMIT = 20;

function Skeleton() {
  return (
    <div className="sdt-page" style={{ padding: "40px 32px" }}>
      {[80, 200, 320].map((h, i) => (
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

export function KommunKpiRankingPage() {
  const { code = "", kpiId = "" } = useParams<{ code: string; kpiId: string }>();
  const isMobile = useMediaQuery("(max-width: 640px)");
  const [showAll, setShowAll] = useState(false);

  const { data: kommunList } = useKommunList();
  const { data: rankEntries = [], isLoading } = useKpiRanking(kpiId);

  const kommunName = kommunList?.find((k) => k.code === code)?.name ?? code;
  const meta = STRIP_KPI_META[kpiId];
  const kpiLabel = meta?.label ?? kpiId;
  const kpiUnit = meta?.unit ?? "";

  // Sort descending (already sorted by backend, but ensure)
  const sorted = [...rankEntries].sort((a, b) => b.value - a.value);

  const mean =
    sorted.length > 0
      ? sorted.reduce((s, e) => s + e.value, 0) / sorted.length
      : 0;
  const maxValue = sorted[0]?.value ?? 1;

  const visible = showAll ? sorted : sorted.slice(0, INITIAL_LIMIT);

  if (isLoading) return <Skeleton />;

  const pad = isMobile ? "14px" : "32px";

  return (
    <div className="sdt-page">
      {/* Breadcrumb */}
      <div
        style={{
          padding: isMobile ? "14px 14px 0" : "28px 32px 0",
          display: "flex",
          alignItems: "center",
          gap: 8,
        }}
      >
        <Link
          to={`/kommun/${code}`}
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            letterSpacing: "0.1em",
            color: "var(--color-accent)",
            textDecoration: "none",
            textTransform: "uppercase",
          }}
        >
          ← {kommunName}
        </Link>
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-fg-muted)",
            letterSpacing: "0.08em",
          }}
        >
          / {kpiLabel}
        </span>
      </div>

      {/* Comparison section */}
      <div
        style={{
          border: "1px solid var(--color-border)",
          margin: isMobile ? "14px 14px 28px" : "20px 32px 28px",
          background: "var(--color-sdt-surface)",
          padding: isMobile ? "14px" : "20px 24px",
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
            marginBottom: 4,
            textTransform: "uppercase",
          }}
        >
          Jämförelse med alla kommuner
        </div>
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 9,
            letterSpacing: "0.08em",
            color: "var(--color-fg-muted)",
            marginBottom: 20,
            opacity: 0.7,
          }}
        >
          Råvärde ({kpiUnit}), senaste år · {kommunName} markerad
        </div>

        {sorted.length === 0 ? (
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg-muted)",
              padding: "20px 0",
              textAlign: "center",
            }}
          >
            Data saknas
          </div>
        ) : (
          <>
            <BarsWithMean
              points={visible.map((e) => ({ region_code: e.mun_code, pct: e.value }))}
              mean={mean}
              maxValue={maxValue}
              highlightCode={code}
              codeToName={new Map(rankEntries.map((e) => [e.mun_code, e.name]))}
              isMobile={isMobile}
            />
            {!showAll && sorted.length > INITIAL_LIMIT && (
              <button
                onClick={() => setShowAll(true)}
                style={{
                  marginTop: 16,
                  padding: "7px 14px",
                  fontFamily: "var(--font-mono)",
                  fontSize: 10,
                  letterSpacing: "0.1em",
                  color: "var(--color-accent)",
                  background: "transparent",
                  border: "1px solid var(--color-border)",
                  cursor: "pointer",
                  display: "block",
                }}
              >
                VISA FLER KOMMUNER ({sorted.length - INITIAL_LIMIT} ST)
              </button>
            )}
          </>
        )}
      </div>

      {/* Legend */}
      <div
        style={{
          padding: `0 ${pad} 16px`,
          display: "flex",
          gap: 16,
          flexWrap: "wrap",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
          <span
            style={{
              display: "inline-block",
              width: 12,
              height: 4,
              borderRadius: 2,
              background: "var(--color-accent, #0b3d7a)",
            }}
          />
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.08em",
            }}
          >
            {kommunName.toUpperCase()}
          </span>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
          <span
            style={{
              display: "inline-block",
              width: 12,
              height: 4,
              borderRadius: 2,
              background: "var(--color-fg-muted, #7a8390)",
            }}
          />
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.08em",
            }}
          >
            ÖVRIGA KOMMUNER
          </span>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
          <span
            style={{
              display: "inline-block",
              width: 1.5,
              height: 12,
              borderLeft: "1.5px dashed #d97706",
            }}
          />
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.08em",
            }}
          >
            MEDELVÄRDE
          </span>
        </div>
      </div>
    </div>
  );
}
