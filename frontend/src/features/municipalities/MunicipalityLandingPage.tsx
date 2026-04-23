import { useState } from "react";
import { NavLink } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { municipalitiesApi } from "./api";
import { SwedenKommunMap } from "./components/SwedenKommunMap";
import type { MunicipalitySummary } from "@/shared/types";

function MunicipalityCard({ mun }: { mun: MunicipalitySummary }) {
  return (
    <NavLink
      to={`/kommun/${mun.code}`}
      style={({ isActive }) => ({
        display: "block",
        padding: "14px 16px",
        border: "1px solid var(--color-border)",
        background: isActive ? "var(--color-track)" : "var(--color-sdt-surface)",
        color: "var(--color-fg)",
        textDecoration: "none",
        transition: "background 0.15s",
      })}
    >
      <div
        style={{
          display: "flex",
          alignItems: "flex-start",
          justifyContent: "space-between",
          marginBottom: 6,
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
            {mun.name}
          </div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              color: "var(--color-fg-muted)",
              marginTop: 2,
            }}
          >
            {mun.regionName}
          </div>
        </div>
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            color: "var(--color-fg-muted)",
            background: "var(--color-track)",
            borderRadius: 2,
            padding: "2px 6px",
          }}
        >
          {mun.code}
        </span>
      </div>

      {/* Governing parties */}
      {mun.governingParties.length > 0 && (
        <div
          style={{
            display: "flex",
            gap: 4,
            flexWrap: "wrap",
            marginBottom: 8,
          }}
        >
          {mun.governingParties.map((p) => (
            <span
              key={p}
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                letterSpacing: "0.5px",
                color: "var(--color-accent)",
                background: "color-mix(in oklch, var(--color-accent) 12%, transparent)",
                borderRadius: 2,
                padding: "1px 6px",
              }}
            >
              {p}
            </span>
          ))}
        </div>
      )}

      <div style={{ display: "flex", gap: 16 }}>
        <div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 12,
              fontWeight: 700,
              color: "var(--color-fg)",
            }}
          >
            {mun.population.toLocaleString("sv-SE")}
          </div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              textTransform: "uppercase",
              letterSpacing: "1px",
            }}
          >
            inv.
          </div>
        </div>
        <div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 12,
              fontWeight: 700,
              color: "var(--color-fg)",
            }}
          >
            {mun.totalMandates}
          </div>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 9,
              color: "var(--color-fg-muted)",
              textTransform: "uppercase",
              letterSpacing: "1px",
            }}
          >
            mandat
          </div>
        </div>
      </div>
    </NavLink>
  );
}

export function MunicipalityLandingPage() {
  const [search, setSearch] = useState("");

  const { data: municipalities, isLoading, error } = useQuery({
    queryKey: ["municipalities"],
    queryFn: () => municipalitiesApi.listMunicipalities(),
    staleTime: 60 * 60 * 1000,
  });

  const filtered = municipalities?.filter(
    (m) =>
      search === "" ||
      m.name.toLowerCase().includes(search.toLowerCase()) ||
      m.regionName.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div
      style={{
        background: "var(--color-bg)",
        color: "var(--color-fg)",
        minHeight: "100vh",
        fontFamily: "var(--font-body)",
      }}
    >
      {/* Header */}
      <div style={{ padding: "40px 32px 0" }}>
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
          KAMMARE TRE
        </div>
        <h1
          style={{
            fontFamily: "var(--font-serif)",
            fontSize: 44,
            fontWeight: 400,
            letterSpacing: "-1.5px",
            lineHeight: 1,
            margin: "0 0 8px",
            color: "var(--color-fg)",
          }}
        >
          Välj kommun
        </h1>
        <p
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-fg-muted)",
            margin: "0 0 24px",
          }}
        >
          Utforska valresultat och styrande partier i Sveriges kommuner.
          Data: Valmyndigheten, val 2022.
        </p>

        {/* Search */}
        <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 28 }}>
          <input
            type="search"
            placeholder="Sök kommun eller region…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            style={{
              background: "var(--color-sdt-surface)",
              border: "1px solid var(--color-border)",
              color: "var(--color-fg)",
              fontFamily: "var(--font-body)",
              fontSize: 13,
              padding: "8px 12px",
              borderRadius: 4,
              outline: "none",
              width: 280,
            }}
          />
          {municipalities && (
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                color: "var(--color-fg-muted)",
              }}
            >
              {filtered?.length ?? 0} av {municipalities.length} kommuner
            </span>
          )}
        </div>
      </div>

      {/* Loading / error */}
      {isLoading && (
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 12,
            color: "var(--color-fg-muted)",
            padding: "64px 32px",
            textAlign: "center",
          }}
        >
          Laddar kommuner…
        </div>
      )}
      {error && (
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 12,
            color: "var(--color-pulse)",
            padding: "32px 32px",
            textAlign: "center",
          }}
        >
          Kunde inte hämta kommundata.
        </div>
      )}

      {/* Map + list */}
      {filtered && municipalities && (
        <div
          style={{
            display: "flex",
            flexDirection: "row",
            gap: 32,
            padding: "0 32px 40px",
            alignItems: "flex-start",
          }}
        >
          {/* Map column */}
          <div style={{ flexShrink: 0, width: 260 }}>
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 10,
                color: "var(--color-fg-muted)",
                textTransform: "uppercase",
                letterSpacing: "1.5px",
                marginBottom: 10,
              }}
            >
              Klicka på en kommun
            </div>
            <SwedenKommunMap municipalities={municipalities} />
            <p
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 9,
                color: "var(--color-fg-muted)",
                marginTop: 8,
                lineHeight: 1.5,
              }}
            >
              Karta: Lokal_Profil / Wikimedia Commons (CC BY-SA 2.5),
              grunddata från SCB.
            </p>
          </div>

          {/* List grid */}
          <div
            style={{
              flex: 1,
              display: "grid",
              gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))",
              gap: 1,
              background: "var(--color-border)",
              border: "1px solid var(--color-border)",
              alignContent: "start",
            }}
          >
            {filtered.map((mun) => (
              <MunicipalityCard key={mun.code} mun={mun} />
            ))}
            {filtered.length === 0 && (
              <div
                style={{
                  gridColumn: "1 / -1",
                  fontFamily: "var(--font-mono)",
                  fontSize: 12,
                  color: "var(--color-fg-muted)",
                  padding: "32px",
                  textAlign: "center",
                }}
              >
                Inga kommuner matchade sökningen.
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
