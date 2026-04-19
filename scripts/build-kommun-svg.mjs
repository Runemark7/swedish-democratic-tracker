// Convert sweden-kommuner.svg (Wikimedia Commons, CC-BY-SA, by Lokal_Profil)
// into a typed TS module.
//
// Also precomputes a per-region viewBox so the embedded "kommuner of region X"
// map can zoom-to-fit without parsing path data at runtime.
import { readFileSync, writeFileSync } from "node:fs";

const SVG_PATH = "frontend/src/features/municipalities/data/sweden-kommuner.svg";
const OUT_PATH = "frontend/src/features/municipalities/data/sweden-kommuner-svg.ts";

const svg = readFileSync(SVG_PATH, "utf8");

let viewBox;
const vbMatch = svg.match(/viewBox="([^"]+)"/);
if (vbMatch) {
  viewBox = vbMatch[1];
} else {
  const w = svg.match(/\bwidth="(\d+)"/)?.[1] ?? "290";
  const h = svg.match(/\bheight="(\d+)"/)?.[1] ?? "660";
  viewBox = `0 0 ${w} ${h}`;
}

const pathRe = /<path\b([^/>]*?)\/>/gs;
const idRe = /\bid="(\d{4})"/;
const dRe = /\bd="([^"]+)"/;

const entries = [];
let m;
while ((m = pathRe.exec(svg)) !== null) {
  const attrs = m[1];
  const idM = attrs.match(idRe);
  const dM = attrs.match(dRe);
  if (!idM || !dM) continue;
  const code = idM[1];
  const regionCode = code.slice(0, 2);
  entries.push({ code, regionCode, d: dM[1] });
}

if (entries.length !== 290) {
  console.warn(`Expected 290 kommun paths, got ${entries.length}`);
}

entries.sort((a, b) => a.code.localeCompare(b.code));

// ---- SVG path bounding box (handles M m L l H h V v C c S s Q q T t A a Z z)
function pathBounds(d) {
  const tokens = d.match(/[a-zA-Z]|-?\d*\.?\d+(?:e[-+]?\d+)?/gi) ?? [];
  let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
  let x = 0, y = 0, startX = 0, startY = 0;
  let cmd = "";
  let i = 0;
  const num = () => Number(tokens[i++]);
  const upd = () => {
    if (x < minX) minX = x;
    if (x > maxX) maxX = x;
    if (y < minY) minY = y;
    if (y > maxY) maxY = y;
  };
  while (i < tokens.length) {
    const t = tokens[i];
    if (/^[a-zA-Z]$/.test(t)) {
      cmd = t;
      i++;
      if (cmd === "Z" || cmd === "z") {
        x = startX;
        y = startY;
        upd();
        cmd = "";
        continue;
      }
    }
    switch (cmd) {
      case "M":
        x = num(); y = num(); startX = x; startY = y; upd();
        cmd = "L"; break;
      case "m":
        x += num(); y += num(); startX = x; startY = y; upd();
        cmd = "l"; break;
      case "L":
        x = num(); y = num(); upd(); break;
      case "l":
        x += num(); y += num(); upd(); break;
      case "H": x = num(); upd(); break;
      case "h": x += num(); upd(); break;
      case "V": y = num(); upd(); break;
      case "v": y += num(); upd(); break;
      case "C":
        num(); num(); num(); num(); x = num(); y = num(); upd(); break;
      case "c":
        num(); num(); num(); num(); x += num(); y += num(); upd(); break;
      case "S":
      case "Q":
        num(); num(); x = num(); y = num(); upd(); break;
      case "s":
      case "q":
        num(); num(); x += num(); y += num(); upd(); break;
      case "T": x = num(); y = num(); upd(); break;
      case "t": x += num(); y += num(); upd(); break;
      case "A":
        num(); num(); num(); num(); num(); x = num(); y = num(); upd(); break;
      case "a":
        num(); num(); num(); num(); num(); x += num(); y += num(); upd(); break;
      default:
        i++;
    }
  }
  if (!Number.isFinite(minX)) return null;
  return { minX, minY, maxX, maxY };
}

// Compute per-region bbox (union of all kommun paths in that region).
const regionBoxes = {};
for (const e of entries) {
  const b = pathBounds(e.d);
  if (!b) continue;
  const acc = regionBoxes[e.regionCode];
  if (!acc) {
    regionBoxes[e.regionCode] = { ...b };
  } else {
    if (b.minX < acc.minX) acc.minX = b.minX;
    if (b.minY < acc.minY) acc.minY = b.minY;
    if (b.maxX > acc.maxX) acc.maxX = b.maxX;
    if (b.maxY > acc.maxY) acc.maxY = b.maxY;
  }
}

const regionViewBoxes = {};
const pad = 6;
for (const [rc, b] of Object.entries(regionBoxes)) {
  const x = Math.floor(b.minX - pad);
  const y = Math.floor(b.minY - pad);
  const w = Math.ceil(b.maxX - b.minX + 2 * pad);
  const h = Math.ceil(b.maxY - b.minY + 2 * pad);
  regionViewBoxes[rc] = `${x} ${y} ${w} ${h}`;
}

const out = `// Auto-generated from src/features/municipalities/data/sweden-kommuner.svg
// Source: Wikimedia Commons "SWE-Map Kommuner.svg" by Lokal_Profil
//   https://commons.wikimedia.org/wiki/File:SWE-Map_Kommuner.svg
// License: CC-BY-SA 2.5 — see ./LICENSING.md
// Original data: Statistics Sweden (SCB)
// Regenerate with: node scripts/build-kommun-svg.mjs

export type KommunSvgPath = {
  code: string;       // 4-digit SCB kommunkod
  regionCode: string; // 2-digit SCB länskod (first two digits of code)
  d: string;
};

export const SWEDEN_KOMMUN_VIEWBOX = "${viewBox}";

/** Tight viewBox per region, precomputed from path bounds. */
export const SWEDEN_KOMMUN_REGION_VIEWBOX: Readonly<Record<string, string>> = {
${Object.entries(regionViewBoxes)
  .sort(([a], [b]) => a.localeCompare(b))
  .map(([rc, vb]) => `  ${JSON.stringify(rc)}: ${JSON.stringify(vb)},`)
  .join("\n")}
};

export const SWEDEN_KOMMUN_PATHS: readonly KommunSvgPath[] = [
${entries
  .map(
    (e) =>
      `  { code: ${JSON.stringify(e.code)}, regionCode: ${JSON.stringify(
        e.regionCode
      )}, d: ${JSON.stringify(e.d)} },`
  )
  .join("\n")}
];
`;

writeFileSync(OUT_PATH, out);
console.log(`Wrote ${entries.length} kommun paths.`);
console.log(`Full viewBox: ${viewBox}`);
console.log(`Region viewBoxes: ${Object.keys(regionViewBoxes).length}`);
