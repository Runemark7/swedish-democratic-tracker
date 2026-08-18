import { useNavigate } from "react-router-dom";
import type { components } from "@/shared/api-contract";

type RegionSummary = components["schemas"]["RegionSummary"];
import {
  SWEDEN_REGION_PATHS,
  SWEDEN_REGION_VIEWBOX,
} from "../data/sweden-regions-svg";
import { SourceMarker } from "@/components/sources/SourceMarker";

type Props = {
  regions: RegionSummary[];
  className?: string;
  highlightedCode?: string;
};

export function SwedenRegionMap({ regions, className, highlightedCode }: Props) {
  const navigate = useNavigate();
  const knownCodes = new Set(regions.map((r) => r.code));

  return (
    <>
      <div
        className={
          "relative w-full max-w-[280px] mx-auto " + (className ?? "")
        }
        style={{ aspectRatio: "810 / 1822" }}
      >
        <svg
          viewBox={SWEDEN_REGION_VIEWBOX}
          width="100%"
          height="100%"
          preserveAspectRatio="xMidYMid meet"
          aria-label="Karta över Sveriges regioner"
          role="img"
        >
          {SWEDEN_REGION_PATHS.map((p) => {
            const known = knownCodes.has(p.scbCode);
            const active = highlightedCode === p.scbCode;
            return (
              <path
                key={p.scbCode}
                d={p.d}
                fill="var(--color-primary)"
                fillOpacity={active ? 0.95 : known ? 0.55 : 0.25}
                stroke="var(--color-surface-lowest)"
                strokeWidth={1.2}
                vectorEffect="non-scaling-stroke"
                style={{
                  cursor: known ? "pointer" : "not-allowed",
                  transition: "fill-opacity 120ms ease",
                }}
                onClick={() => {
                  if (known) navigate(`/region/${p.scbCode}`);
                }}
                onMouseEnter={(e) => {
                  if (known) e.currentTarget.setAttribute("fill-opacity", "0.9");
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.setAttribute(
                    "fill-opacity",
                    String(active ? 0.95 : known ? 0.55 : 0.25)
                  );
                }}
                role={known ? "link" : undefined}
                aria-label={p.name}
              >
                <title>{p.name}</title>
              </path>
            );
          })}
        </svg>
      </div>
      <div
        style={{
          fontSize: 9,
          color: "var(--color-fg-muted)",
          textAlign: "center",
          marginTop: 4,
        }}
      >
        Karta: Wikimedia Commons <SourceMarker sourceId="wikimedia-svg" />
      </div>
    </>
  );
}
