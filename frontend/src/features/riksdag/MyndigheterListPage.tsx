import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useMyndigheterList, useMyndigheterStats } from "@/hooks/useDemocracy";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { SourceMarker } from "@/components/sources/SourceMarker";
import type { SourceId } from "@/components/sources/SourceRegistry";

type UnderGovChoice = "all" | "true" | "false";

function Stat({ label, value, caption, missing, sourceId }: { label: string; value: string; caption: string; missing?: number; sourceId?: SourceId }) {
  return (
    <div>
      <div style={{ fontFamily: "var(--font-mono)", fontSize: 9, letterSpacing: "0.15em", color: "var(--color-fg-muted)", textTransform: "uppercase", marginBottom: 6 }}>
        {label}
      </div>
      <div style={{ fontFamily: "var(--font-serif)", fontSize: 26, fontWeight: 400, color: "var(--color-fg)", lineHeight: 1.05 }}>
        {value}
        {sourceId && <SourceMarker sourceId={sourceId} />}
      </div>
      <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)", marginTop: 4, lineHeight: 1.4 }}>
        {caption}
        {missing !== undefined && missing > 0 && (
          <>
            {" · "}
            <span style={{ color: "var(--color-down, #b8391c)" }}>{missing} saknar data</span>
          </>
        )}
      </div>
    </div>
  );
}

export function MyndigheterListPage() {
  const isMobile = useMediaQuery("(max-width: 640px)");
  const [q, setQ] = useState("");
  const [underGov, setUnderGov] = useState<UnderGovChoice>("all");
  const [page, setPage] = useState(1);
  const pageSize = 50;

  const filter = useMemo(
    () => ({
      q: q.trim() || undefined,
      underGovernment: underGov === "all" ? undefined : underGov === "true",
      page,
      pageSize,
    }),
    [q, underGov, page]
  );

  const { data, isLoading } = useMyndigheterList(filter);
  const { data: stats } = useMyndigheterStats();
  const items = data?.items ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <div className="sdt-page" style={{ background: "var(--color-sdt-surface)", minHeight: "100vh" }}>
      <div style={{ padding: isMobile ? "18px 14px 0" : "40px 32px 0" }}>
        <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, letterSpacing: "2px", color: "var(--color-fg-muted)", marginBottom: 6 }}>
          MYNDIGHETSREGISTER
        </div>
        <div style={{ fontFamily: "var(--font-serif)", fontSize: isMobile ? 28 : 44, fontWeight: 400, letterSpacing: "-1px", color: "var(--color-fg)", lineHeight: 1, marginBottom: 6 }}>
          Sveriges myndigheter
        </div>
        <div style={{ fontSize: 13, color: "var(--color-fg-muted)" }}>
          {total} myndigheter · Källa: SCB Myndighetsregistret + Statskontorets Myndighetsförteckning
        </div>
      </div>

      {stats && (
        <div
          style={{
            margin: isMobile ? "14px 14px 0" : "22px 32px 0",
            border: "1px solid var(--color-border)",
            background: "var(--color-sdt-surface)",
            padding: isMobile ? 16 : "18px 22px",
          }}
        >
          <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, letterSpacing: "0.15em", color: "var(--color-fg-muted)", marginBottom: 14 }}>
            ÖVERSIKT
          </div>
          <div style={{ display: "grid", gridTemplateColumns: isMobile ? "repeat(2, 1fr)" : "repeat(4, 1fr)", gap: isMobile ? 14 : 24 }}>
            <Stat
              label="Totalsumma utgift"
              value={`${stats.totalExpenditureMdkr.toFixed(0)} mdkr`}
              caption={`${stats.withExpenditure} av ${stats.total} myndigheter har utgiftsdata`}
              missing={stats.total - stats.withExpenditure}
              sourceId="statskontoret-arsutfall"
            />
            <Stat
              label="Totalt anställda"
              value={stats.totalHeadcount.toLocaleString("sv-SE")}
              caption={`${stats.withHeadcount} av ${stats.total} har anställdadata`}
              missing={stats.total - stats.withHeadcount}
            />
            <Stat
              label="Under regeringen"
              value={`${stats.underGovernment}`}
              caption={`av ${stats.total} totalt (resten under riksdagen, domstolar, AP-fonder, utlandsmyndigheter)`}
            />
            <Stat
              label="Senast uppdaterad"
              value={new Date(stats.latestUpdatedAt).toLocaleDateString("sv-SE")}
              caption={`Veckovis ingestion från SCB + Statskontoret`}
            />
          </div>
        </div>
      )}

      <div style={{ display: "flex", gap: 12, flexWrap: "wrap", padding: isMobile ? "14px 14px 0" : "22px 32px 0", alignItems: "center" }}>
        <input
          type="search"
          value={q}
          onChange={(e) => { setQ(e.target.value); setPage(1); }}
          placeholder="Sök myndighet..."
          style={{
            flex: 1,
            minWidth: 220,
            fontFamily: "var(--font-mono)",
            fontSize: 13,
            padding: "8px 12px",
            background: "var(--color-sdt-surface)",
            color: "var(--color-fg)",
            border: "1px solid var(--color-border)",
            borderRadius: 4,
          }}
        />
        <div style={{ display: "flex", gap: 2, fontFamily: "var(--font-mono)", fontSize: 11 }}>
          {(["all", "true", "false"] as UnderGovChoice[]).map((v) => (
            <button
              key={v}
              onClick={() => { setUnderGov(v); setPage(1); }}
              style={{
                padding: "6px 10px",
                background: underGov === v ? "var(--color-accent)" : "var(--color-sdt-surface)",
                color: underGov === v ? "var(--color-on-primary, #fff)" : "var(--color-fg)",
                border: "1px solid var(--color-border)",
                cursor: "pointer",
                letterSpacing: "0.05em",
              }}
            >
              {v === "all" ? "ALLA" : v === "true" ? "UNDER REGERINGEN" : "UNDER RIKSDAGEN"}
            </button>
          ))}
        </div>
      </div>

      <div style={{ margin: isMobile ? "14px 14px 28px" : "22px 32px 32px", border: "1px solid var(--color-border)", borderRadius: 8, overflow: "hidden" }}>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 90px 100px 110px 130px 14px", gap: 8, padding: "8px 14px", background: "var(--color-surface-low)", fontFamily: "var(--font-mono)", fontSize: 9, letterSpacing: "0.12em", color: "var(--color-on-surface-variant)", textTransform: "uppercase" }}>
          <span>Myndighet</span>
          <span style={{ textAlign: "right" }}>Anställda</span>
          <span style={{ textAlign: "right" }}>Utgift</span>
          <span style={{ textAlign: "right" }}>Departement</span>
          <span style={{ textAlign: "right" }}>Typ</span>
          <span />
        </div>

        {isLoading && <div style={{ padding: 24, textAlign: "center", color: "var(--color-fg-muted)", fontFamily: "var(--font-mono)", fontSize: 12 }}>Laddar…</div>}

        {!isLoading && items.length === 0 && (
          <div style={{ padding: 24, textAlign: "center", color: "var(--color-fg-muted)", fontFamily: "var(--font-mono)", fontSize: 12 }}>
            Inga träffar.
          </div>
        )}

        {items.map((a) => (
          <Link
            key={a.orgNumber}
            to={`/riksdag/myndigheter/${a.slug}`}
            style={{ display: "grid", gridTemplateColumns: "1fr 90px 100px 110px 130px 14px", gap: 8, padding: "10px 14px", alignItems: "center", borderTop: "1px solid var(--color-border)", textDecoration: "none", color: "inherit" }}
          >
            <span style={{ fontSize: 13, fontWeight: 500, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
              {a.name}
            </span>
            <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, textAlign: "right", color: a.headcountInt ? "var(--color-fg)" : "var(--color-fg-muted)" }}>
              {a.headcountInt ? a.headcountInt.toLocaleString("sv-SE") : "saknas"}
            </span>
            <span style={{ fontFamily: "var(--font-mono)", fontSize: 11, textAlign: "right", color: a.expenditureMdkr != null ? "var(--color-fg)" : "var(--color-fg-muted)" }}>
              {a.expenditureMdkr != null ? `${a.expenditureMdkr.toFixed(1)} mdkr` : "saknas"}
              {a.expenditureMdkr != null && <SourceMarker sourceId="statskontoret-arsutfall" />}
            </span>
            <span style={{ fontFamily: "var(--font-mono)", fontSize: 10, textAlign: "right", color: "var(--color-fg-muted)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
              {a.department || "—"}
            </span>
            <span style={{ fontFamily: "var(--font-mono)", fontSize: 10, textAlign: "right", color: "var(--color-fg-muted)" }}>
              {a.type}
            </span>
            <span style={{ fontFamily: "var(--font-mono)", fontSize: 13, color: "var(--color-fg-muted)" }}>›</span>
          </Link>
        ))}

        {total > pageSize && (
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "10px 14px", borderTop: "1px solid var(--color-border)", background: "var(--color-surface-low)", fontFamily: "var(--font-mono)", fontSize: 11 }}>
            <span style={{ color: "var(--color-fg-muted)" }}>Sida {page} av {totalPages}</span>
            <div style={{ display: "flex", gap: 6 }}>
              <button disabled={page <= 1} onClick={() => setPage((p) => Math.max(1, p - 1))} style={{ padding: "4px 10px", fontFamily: "inherit", background: "var(--color-sdt-surface)", color: page <= 1 ? "var(--color-fg-muted)" : "var(--color-fg)", border: "1px solid var(--color-border)", cursor: page <= 1 ? "default" : "pointer" }}>← Föregående</button>
              <button disabled={page >= totalPages} onClick={() => setPage((p) => Math.min(totalPages, p + 1))} style={{ padding: "4px 10px", fontFamily: "inherit", background: "var(--color-sdt-surface)", color: page >= totalPages ? "var(--color-fg-muted)" : "var(--color-fg)", border: "1px solid var(--color-border)", cursor: page >= totalPages ? "default" : "pointer" }}>Nästa →</button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
