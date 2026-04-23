import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { regionsApi } from "./api";
import { SwedenRegionMap } from "./components/SwedenRegionMap";
import type { RegionSummary } from "@/shared/types";

function RegionCard({ region }: { region: RegionSummary }) {
  return (
    <Link
      to={`/region/${region.code}`}
      style={{
        display: "block",
        background: "var(--color-sdt-surface)",
        border: "1px solid var(--color-border)",
        borderRadius: 4,
        padding: "14px 16px",
        textDecoration: "none",
        transition: "border-color 0.15s",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          marginBottom: 8,
        }}
      >
        <div>
          <div
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: 15,
              fontWeight: 400,
              color: "var(--color-fg)",
              lineHeight: 1.2,
            }}
          >
            {region.name}
          </div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              color: "var(--color-fg-muted)",
              marginTop: 2,
            }}
          >
            {region.capital}
          </div>
        </div>
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 9,
            color: "var(--color-fg-muted)",
            background: "var(--color-track)",
            borderRadius: 2,
            padding: "2px 5px",
            letterSpacing: "0.05em",
          }}
        >
          {region.code}
        </span>
      </div>

      <div
        style={{
          display: "flex",
          gap: 4,
          flexWrap: "wrap",
          marginBottom: 10,
        }}
      >
        {region.governingParties.map((p) => (
          <span
            key={p}
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              background: "var(--color-track)",
              borderRadius: 2,
              padding: "2px 5px",
              letterSpacing: "0.05em",
            }}
          >
            {p}
          </span>
        ))}
      </div>

      <div style={{ display: "flex", gap: 16 }}>
        <div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg)",
              fontVariantNumeric: "tabular-nums",
            }}
          >
            {region.population.toLocaleString("sv-SE")}
          </div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.1em",
              textTransform: "uppercase",
              marginTop: 1,
            }}
          >
            invånare
          </div>
        </div>
        <div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg)",
              fontVariantNumeric: "tabular-nums",
            }}
          >
            {region.totalMandates}
          </div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.1em",
              textTransform: "uppercase",
              marginTop: 1,
            }}
          >
            mandat
          </div>
        </div>
        <div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg)",
              fontVariantNumeric: "tabular-nums",
            }}
          >
            {region.electionYear}
          </div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.1em",
              textTransform: "uppercase",
              marginTop: 1,
            }}
          >
            val
          </div>
        </div>
      </div>
    </Link>
  );
}

export function RegionLandingPage() {
  const { data: regions, isLoading, error } = useQuery({
    queryKey: ["regions"],
    queryFn: regionsApi.listRegions,
    staleTime: 60 * 60 * 1000,
  });

  return (
    <div
      style={{
        background: "var(--color-bg)",
        color: "var(--color-fg)",
        minHeight: "100%",
      }}
    >
      {/* Header */}
      <div style={{ padding: "40px 32px 0" }}>
        <div
          style={{
            display: "flex",
            alignItems: "flex-end",
            gap: 20,
            marginBottom: 6,
          }}
        >
          <span
            style={{
              fontFamily: "var(--font-serif)",
              fontStyle: "italic",
              color: "var(--color-accent-2)",
              fontSize: 88,
              fontWeight: 400,
              letterSpacing: "-3px",
              lineHeight: 0.85,
              flexShrink: 0,
            }}
          >
            II
          </span>
          <div>
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "2px",
                color: "var(--color-fg-muted)",
                marginBottom: 6,
              }}
            >
              KAMMARE TVÅ
            </div>
            <div
              style={{
                fontFamily: "var(--font-serif)",
                fontSize: 52,
                fontWeight: 400,
                letterSpacing: "-1.5px",
                color: "var(--color-fg)",
                lineHeight: 1,
                marginBottom: 8,
              }}
            >
              Välj region
            </div>
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                color: "var(--color-fg-muted)",
                letterSpacing: "0.05em",
              }}
            >
              Utforska valresultat och styrande koalitioner i Sveriges 21
              regioner. Data: Valmyndigheten, val 2022.
            </div>
          </div>
        </div>
      </div>

      {isLoading && (
        <div style={{ padding: "40px 32px" }}>
          {[80, 80, 80].map((h, i) => (
            <div
              key={i}
              style={{
                height: h,
                background: "var(--color-track)",
                borderRadius: 4,
                marginBottom: 12,
              }}
            />
          ))}
        </div>
      )}

      {error && (
        <p
          style={{
            padding: "32px",
            color: "var(--color-pulse)",
            fontFamily: "var(--font-mono)",
            fontSize: 12,
          }}
        >
          Kunde inte hämta regiondata.
        </p>
      )}

      {regions && (
        <div
          style={{
            display: "flex",
            flexDirection: "row",
            gap: 32,
            padding: "28px 32px 40px",
            alignItems: "flex-start",
          }}
        >
          {/* Map sidebar */}
          <div style={{ flexShrink: 0, width: 280 }}>
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 9,
                letterSpacing: "0.15em",
                color: "var(--color-fg-muted)",
                textTransform: "uppercase",
                marginBottom: 12,
                textAlign: "center",
              }}
            >
              Klicka på en region
            </div>
            <SwedenRegionMap regions={regions} />
          </div>

          {/* Region cards grid */}
          <div
            style={{
              flex: 1,
              display: "grid",
              gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))",
              gap: 8,
            }}
          >
            {regions.map((region) => (
              <div key={region.code} id={`region-${region.code}`}>
                <RegionCard region={region} />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
