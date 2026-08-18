import { useMemo } from "react";
import { useNavigate } from "react-router-dom";
import type { components } from "@/shared/api-contract";

type MunicipalitySummary = components["schemas"]["MunicipalitySummary"];
import {
  SWEDEN_KOMMUN_PATHS,
  SWEDEN_KOMMUN_VIEWBOX,
  SWEDEN_KOMMUN_REGION_VIEWBOX,
} from "../data/sweden-kommuner-svg";
import { SourceMarker } from "@/components/sources/SourceMarker";

type Props = {
  municipalities: MunicipalitySummary[];
  /** When set, only this region's kommuner are clickable; the rest fade. */
  regionCode?: string;
  className?: string;
  highlightedCode?: string;
};

// Distinct hue per region — visually distinguishes neighbours without
// implying a political/data meaning.
const REGION_HUES: Record<string, string> = {
  "01": "#003461", "03": "#2a5d8f", "04": "#3d7ab0", "05": "#0e3b5b",
  "06": "#7a4f9a", "07": "#9a6ba8", "08": "#5b3a7c", "09": "#b97e2c",
  "10": "#d6a13a", "12": "#a83a3a", "13": "#c45a5a", "14": "#1f6f4a",
  "17": "#2f8b66", "18": "#4d9a7c", "19": "#5fa68e", "20": "#84794d",
  "21": "#9c8b3c", "22": "#3e8d9c", "23": "#5fa6b2", "24": "#7d4848",
  "25": "#5b3030",
};

export function SwedenKommunMap({
  municipalities,
  regionCode,
  className,
  highlightedCode,
}: Props) {
  const navigate = useNavigate();

  const knownCodes = useMemo(
    () => new Set(municipalities.map((m) => m.code)),
    [municipalities]
  );

  const nameByCode = useMemo(() => {
    const map = new Map<string, string>();
    for (const m of municipalities) map.set(m.code, m.name);
    return map;
  }, [municipalities]);

  const { viewBox, paths, aspect } = useMemo(() => {
    const vb =
      (regionCode && SWEDEN_KOMMUN_REGION_VIEWBOX[regionCode]) ??
      SWEDEN_KOMMUN_VIEWBOX;
    const ps = regionCode
      ? SWEDEN_KOMMUN_PATHS.filter((p) => p.regionCode === regionCode)
      : SWEDEN_KOMMUN_PATHS;
    const [, , w, h] = vb.split(/\s+/).map(Number);
    return { viewBox: vb, paths: ps, aspect: `${w} / ${h}` };
  }, [regionCode]);

  return (
    <>
      <div
        className={"relative w-full mx-auto " + (className ?? "")}
        style={{ aspectRatio: aspect }}
      >
        <svg
          viewBox={viewBox}
          width="100%"
          height="100%"
          preserveAspectRatio="xMidYMid meet"
          aria-label={
            regionCode
              ? `Karta över kommunerna i region ${regionCode}`
              : "Karta över Sveriges kommuner"
          }
          role="img"
        >
          {paths.map((p) => {
            const known = knownCodes.has(p.code);
            const active = highlightedCode === p.code;
            const fill = REGION_HUES[p.regionCode] ?? "#003461";
            const opacity = active ? 0.95 : known ? 0.65 : 0.3;
            const name = nameByCode.get(p.code);
            return (
              <path
                key={p.code}
                d={p.d}
                fill={fill}
                fillOpacity={opacity}
                stroke="#ffffff"
                strokeWidth={0.4}
                vectorEffect="non-scaling-stroke"
                style={{
                  cursor: known ? "pointer" : "default",
                  transition: "fill-opacity 120ms ease",
                }}
                onClick={() => {
                  if (known) navigate(`/kommun/${p.code}`);
                }}
                onMouseEnter={(e) => {
                  if (known) e.currentTarget.setAttribute("fill-opacity", "0.95");
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.setAttribute("fill-opacity", String(opacity));
                }}
                role={known ? "link" : undefined}
                aria-label={name ?? `Kommun ${p.code}`}
              >
                <title>{name ?? `Kommun ${p.code}`}</title>
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
