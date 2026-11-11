// Fails when the UI says a Löfte is "hållet" or "brutet" (ADR-002): Löftesstatus
// is vote-based and never a verdict. Scans the frontend source and the public
// data-source pages. Only whole words count, so "innehållet" is fine.
import { existsSync } from "node:fs";
import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "..", "..");
// frontend/src is resolved from this script so it also works in the Docker build,
// where the context is frontend/ only and docs/ is absent (missing roots are skipped).
const ROOTS = [path.resolve(__dirname, "..", "src"), path.join(REPO_ROOT, "docs", "data-sources")];
const EXT = new Set([".ts", ".tsx", ".md", ".html"]);
export const FORBIDDEN = /(?<![\p{L}])(hållet|brutet)(?![\p{L}])/giu;

export function findForbidden(text) {
  const hits = [];
  text.split("\n").forEach((line, i) => {
    for (const m of line.matchAll(FORBIDDEN)) hits.push({ line: i + 1, word: m[0] });
  });
  return hits;
}

async function* walk(dir) {
  for (const e of await readdir(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) {
      if (e.name !== "node_modules") yield* walk(p);
    } else if (EXT.has(path.extname(e.name))) yield p;
  }
}

export async function main() {
  const failures = [];
  for (const root of ROOTS) {
    if (!existsSync(root)) continue;
    for await (const file of walk(root)) {
      for (const h of findForbidden(await readFile(file, "utf8"))) {
        failures.push(`${path.relative(REPO_ROOT, file)}:${h.line}: "${h.word}"`);
      }
    }
  }
  if (failures.length) {
    console.error('The site must never say "hållet" or "brutet" (adr/002-loftesstatus-is-vote-based.md):\n' + failures.join("\n"));
    process.exit(1);
  }
  console.log("check-forbidden-words: ok");
}

if (process.argv[1] === fileURLToPath(import.meta.url)) await main();
