// frontend/scripts/verify-sources.mjs
import { readdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import matter from "gray-matter";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "..", "..");
const SRC_DIR = path.join(REPO_ROOT, "docs", "data-sources");
const REPORT_PATH = path.join(REPO_ROOT, "frontend", "verify-report.json");
const TIMEOUT_MS = 30_000;
const MAX_ATTEMPTS = 2;

async function attemptFetch(entry, fetchImpl) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), TIMEOUT_MS);
  try {
    let res = await fetchImpl(entry.upstream, { method: "HEAD", signal: controller.signal });
    const headStatus = res.status;
    if (res.status === 405) {
      res = await fetchImpl(entry.upstream, { method: "GET", signal: controller.signal });
    }
    return { res, headStatus };
  } finally {
    clearTimeout(timer);
  }
}

export async function checkSource(entry, { fetch: fetchImpl = globalThis.fetch } = {}) {
  if (entry.kind === "seed" || entry.kind === "synthesized") {
    return { id: entry.id, status: "skipped", reason: `kind=${entry.kind}` };
  }
  if (!entry.upstream) {
    return { id: entry.id, status: "skipped", reason: "no upstream URL" };
  }

  let headStatus;
  let res;
  let lastErr;
  for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
    try {
      ({ res, headStatus } = await attemptFetch(entry, fetchImpl));
      lastErr = null;
      break;
    } catch (err) {
      lastErr = err;
      // Only retry timeouts; permanent errors fail fast.
      if (err.name !== "AbortError") break;
    }
  }

  if (lastErr) {
    if (lastErr.name === "AbortError") {
      return { id: entry.id, status: "timeout" };
    }
    return { id: entry.id, status: "error", reason: String(lastErr.message ?? lastErr) };
  }

  const contentType = (
    res.headers.get?.("content-type") ??
    res.headers.get?.("Content-Type") ??
    ""
  )
    .toString()
    .toLowerCase();

  // SCB PxWeb endpoints are POST-only: they return 400 (bad request — no body)
  // on both HEAD and GET. Treat 400 from api.scb.se as reachable.
  if ((res.status === 400 || res.status === 405) && entry.upstream.includes("api.scb.se")) {
    return { id: entry.id, status: "ok", httpCode: res.status, contentType: "pxweb-post-only" };
  }

  // Generic POST-only APIs: HEAD→405 and GET→405 means the endpoint exists but
  // only accepts POST. Treat as reachable.
  if (headStatus === 405 && res.status === 405) {
    return { id: entry.id, status: "ok", httpCode: 405, contentType: "post-only" };
  }

  if (!(res.status >= 200 && res.status < 400)) {
    return { id: entry.id, status: "unreachable", httpCode: res.status };
  }

  if (entry.kind === "csv") {
    // If the upstream URL has no file extension at all, it is a landing page
    // (the page itself is the pointer to the actual file). Returning HTML at
    // 200 is expected behaviour — skip the content-type check.
    // If the URL *does* have an extension (any extension), we validate that
    // the content-type looks like a binary data file.
    const urlPath = new URL(entry.upstream).pathname;
    const hasAnyExt = /\.[a-z]{2,5}$/i.test(urlPath);
    if (hasAnyExt) {
      const okType =
        contentType.includes("csv") ||
        contentType.includes("zip") ||
        contentType.includes("octet-stream") ||
        contentType.includes("spreadsheet") ||
        contentType.includes("excel");
      if (!okType) {
        return {
          id: entry.id,
          status: "wrong-content-type",
          httpCode: res.status,
          contentType,
        };
      }
    }
  }

  return { id: entry.id, status: "ok", httpCode: res.status, contentType };
}

async function loadEntries() {
  const files = (await readdir(SRC_DIR)).filter(
    (f) => f.endsWith(".md") && !f.startsWith("_") && f !== "INDEX.md"
  );
  const entries = [];
  for (const f of files) {
    const raw = await readFile(path.join(SRC_DIR, f), "utf8");
    const { data: fm } = matter(raw);
    entries.push({
      id: fm.id,
      kind: fm.kind,
      upstream: fm.upstream && fm.upstream !== "~" ? fm.upstream : null,
    });
  }
  return entries;
}

async function main() {
  const entries = await loadEntries();
  const results = await Promise.all(entries.map((e) => checkSource(e)));
  const failures = results.filter(
    (r) => r.status !== "ok" && r.status !== "skipped"
  );

  await writeFile(REPORT_PATH, JSON.stringify({ results }, null, 2));

  for (const r of results) {
    const line =
      r.status === "ok"
        ? `  ✓ ${r.id} — HTTP ${r.httpCode} (${r.contentType})`
        : r.status === "skipped"
        ? `  · ${r.id} — skipped (${r.reason})`
        : `  ✗ ${r.id} — ${r.status}${r.httpCode ? ` (HTTP ${r.httpCode})` : ""}${r.reason ? ` — ${r.reason}` : ""}`;
    console.log(line);
  }

  if (failures.length > 0) {
    console.error(`\nverify-sources: ${failures.length} source(s) failed verification`);
    process.exit(1);
  }
  console.log(`\nverify-sources: ${results.length} source(s) checked, all healthy`);
}

if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch((err) => {
    console.error("verify-sources failed:", err.message);
    process.exit(1);
  });
}
