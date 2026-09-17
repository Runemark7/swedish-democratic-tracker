// One-shot: convert a swedenMap.js export (e.g. from arbetsProvEvry) into a
// typed TS module keyed by SCB 2-digit region code. The generated output is
// committed, so this only needs to run when regenerating.
// Run with: node scripts/build-region-svg.mjs <path-to-swedenMap.js>
import { writeFileSync } from "node:fs";
import { pathToFileURL } from "node:url";

const source = process.argv[2];
if (!source) {
  console.error("Usage: node scripts/build-region-svg.mjs <path-to-swedenMap.js>");
  process.exit(1);
}
const { default: map } = await import(pathToFileURL(source).href);

// ISO 3166-2:SE → SCB 2-digit län code
const ISO_TO_SCB = {
  "SE-K":  "10", // Blekinge
  "SE-W":  "20", // Dalarna
  "SE-I":  "09", // Gotland
  "SE-X":  "21", // Gävleborg
  "SE-N":  "13", // Halland
  "SE-Z":  "23", // Jämtland
  "SE-F":  "06", // Jönköping
  "SE-H":  "08", // Kalmar
  "SE-G":  "07", // Kronoberg
  "SE-BD": "25", // Norrbotten
  "SE-M":  "12", // Skåne
  "SE-AB": "01", // Stockholm
  "SE-D":  "04", // Södermanland
  "SE-C":  "03", // Uppsala
  "SE-S":  "17", // Värmland
  "SE-AC": "24", // Västerbotten
  "SE-Y":  "22", // Västernorrland
  "SE-U":  "19", // Västmanland
  "SE-O":  "14", // Västra Götaland
  "SE-T":  "18", // Örebro
  "SE-E":  "05", // Östergötland
};

const paths = map.g.path;

let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
for (const p of paths) {
  const nums = p["-d"].match(/-?\d+(?:\.\d+)?/g).map(Number);
  for (let i = 0; i < nums.length; i += 2) {
    if (nums[i]   < minX) minX = nums[i];
    if (nums[i]   > maxX) maxX = nums[i];
    if (nums[i+1] < minY) minY = nums[i+1];
    if (nums[i+1] > maxY) maxY = nums[i+1];
  }
}
const pad = 5;
const vbX = Math.floor(minX - pad);
const vbY = Math.floor(minY - pad);
const vbW = Math.ceil(maxX - minX + 2 * pad);
const vbH = Math.ceil(maxY - minY + 2 * pad);

const entries = paths
  .map((p) => {
    const iso = p["-id"];
    const scb = ISO_TO_SCB[iso];
    if (!scb) throw new Error(`No SCB code for ISO ${iso}`);
    return { scbCode: scb, isoCode: iso, name: p["-title"], d: p["-d"] };
  })
  .sort((a, b) => a.scbCode.localeCompare(b.scbCode));

const out = `// Auto-generated from arbetsProvEvry/src/assets/swedenMap.js
// Source: see ./LICENSING.md (license unverified — TODO before public launch)
// Regenerate with: node scripts/build-region-svg.mjs

export type RegionSvgPath = {
  scbCode: string;
  isoCode: string;
  name: string;
  d: string;
};

export const SWEDEN_REGION_VIEWBOX = "${vbX} ${vbY} ${vbW} ${vbH}";

export const SWEDEN_REGION_PATHS: readonly RegionSvgPath[] = [
${entries
  .map(
    (e) =>
      `  { scbCode: ${JSON.stringify(e.scbCode)}, isoCode: ${JSON.stringify(
        e.isoCode
      )}, name: ${JSON.stringify(e.name)}, d: ${JSON.stringify(e.d)} },`
  )
  .join("\n")}
];
`;

writeFileSync("frontend/src/features/regions/data/sweden-regions-svg.ts", out);
console.log(`Wrote ${entries.length} region paths.`);
console.log(`viewBox: ${vbX} ${vbY} ${vbW} ${vbH}`);
