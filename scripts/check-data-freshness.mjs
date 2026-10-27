// Compares the live site's freshness report with Riksdagen.
//
// A quiet chamber makes our newest speech old. The failure that hid for
// months was the opposite: Riksdagen had new speeches and we did not.
// This fails when our date is more than LAG_DAYS behind the source, or when
// a job has not succeeded inside its schedule's window.
//
// Usage: node scripts/check-data-freshness.mjs
//   SITE   default https://sdt.runevibe.se

const LAG_DAYS = 7;
const DAILY_JOB_HOURS = 48;
const WEEKLY_JOB_HOURS = 9 * 24;
const SITE = (process.env.SITE || "https://sdt.runevibe.se").replace(/\/$/, "");

export function currentRiksmote(date) {
  const y0 = date.getUTCFullYear();
  const month = date.getUTCMonth() + 1;
  const y = month < 9 ? y0 - 1 : y0;
  return `${y}/${String((y + 1) % 100).padStart(2, "0")}`;
}

export function daysBetween(earlier, later) {
  const a = Date.parse(`${earlier}T00:00:00Z`);
  const b = Date.parse(`${later}T00:00:00Z`);
  if (Number.isNaN(a) || Number.isNaN(b)) return null;
  return Math.floor((b - a) / 86_400_000);
}

// Calendar day, compared below as midnight UTC. The API's latestDate is the
// stored DATE (speeches) or the UTC day of system_datum (votes), not a
// session-local midnight. Taking the first ten characters keeps a date-only
// value on that day; parsing it in local time would move it.
function dateOnly(value) {
  if (!value || typeof value !== "string") return null;
  const day = value.slice(0, 10);
  return /^\d{4}-\d{2}-\d{2}$/.test(day) ? day : null;
}

function hoursSince(iso, now) {
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return null;
  return (now.getTime() - t) / 3_600_000;
}

/**
 * @param {object} input
 * @param {Date} input.now
 * @param {object} input.freshness parsed /api/ingestion/freshness body
 * @param {string|null} input.upstreamSpeechDate
 * @param {string|null} input.upstreamVoteDate
 * @param {string|null} input.upstreamSpeechError
 * @param {string|null} input.upstreamVoteError
 */
export function evaluateFreshness({
  now,
  freshness,
  upstreamSpeechDate,
  upstreamVoteDate,
  upstreamSpeechError = null,
  upstreamVoteError = null,
}) {
  const reasons = [];
  if (!freshness || typeof freshness !== "object") {
    reasons.push("freshness report missing");
    return { ok: false, reasons };
  }

  const jobs = Array.isArray(freshness.jobs) ? freshness.jobs : [];
  for (const job of jobs) {
    const limit = job.schedule === "weekly" ? WEEKLY_JOB_HOURS : DAILY_JOB_HOURS;
    const label = job.schedule === "weekly" ? "weekly" : "daily";
    if (!job.lastSuccessAt) {
      reasons.push(`${job.name}: has never succeeded`);
      continue;
    }
    const age = hoursSince(job.lastSuccessAt, now);
    if (age == null) {
      reasons.push(`${job.name}: last success ${job.lastSuccessAt} is not a time`);
      continue;
    }
    if (age > limit) {
      reasons.push(
        `${job.name}: last success ${job.lastSuccessAt} is ${Math.floor(age)}h old (${label} limit ${limit}h)`,
      );
    }
  }

  if (upstreamSpeechError) reasons.push(`riksdagen speeches: ${upstreamSpeechError}`);
  if (upstreamVoteError) reasons.push(`riksdagen votes: ${upstreamVoteError}`);

  const speechLag = lagReason(
    "speeches",
    dateOnly(freshness.speeches?.latestDate),
    upstreamSpeechDate,
  );
  if (speechLag) reasons.push(speechLag);
  const voteLag = lagReason("votes", dateOnly(freshness.votes?.latestDate), upstreamVoteDate);
  if (voteLag) reasons.push(voteLag);

  return { ok: reasons.length === 0, reasons };
}

function lagReason(kind, ours, upstream) {
  if (!upstream) return null;
  if (!ours) return `${kind}: we hold none; Riksdagen has ${upstream}`;
  const lag = daysBetween(ours, upstream);
  if (lag == null) return `${kind}: could not compare ${ours} with ${upstream}`;
  if (lag > LAG_DAYS) {
    return `${kind}: ours ${ours} is ${lag} days behind Riksdagen ${upstream} (limit ${LAG_DAYS})`;
  }
  return null;
}

function maxDate(values) {
  const days = values.map(dateOnly).filter(Boolean).sort();
  return days.length ? days[days.length - 1] : null;
}

async function fetchJSON(url) {
  const res = await fetch(url, { headers: { accept: "application/json" } });
  if (!res.ok) {
    throw new Error(`${url} returned ${res.status}`);
  }
  return res.json();
}

export async function latestSpeechDate(rm) {
  const url = `https://data.riksdagen.se/anforandelista/?rm=${encodeURIComponent(rm)}&sz=50&utformat=json`;
  const body = await fetchJSON(url);
  let rows = body?.anforandelista?.anforande ?? [];
  if (!Array.isArray(rows)) rows = [rows];
  return maxDate(rows.map((row) => row?.dok_datum));
}

export async function latestVoteDate(rm) {
  const url =
    `https://data.riksdagen.se/dokumentlista/?doktyp=votering&rm=${encodeURIComponent(rm)}` +
    `&sz=20&utformat=json&sort=datum&sortorder=desc`;
  const body = await fetchJSON(url);
  let rows = body?.dokumentlista?.dokument ?? [];
  if (!Array.isArray(rows)) rows = [rows];
  return maxDate(rows.map((row) => row?.systemdatum || row?.datum));
}

async function main() {
  const now = new Date();
  const rm = currentRiksmote(now);
  const lines = [`site ${SITE}`, `riksmote ${rm}`, `checked ${now.toISOString()}`];

  let freshness = null;
  let freshnessError = null;
  try {
    freshness = await fetchJSON(`${SITE}/api/ingestion/freshness`);
    lines.push(`speeches ${freshness?.speeches?.latestDate ?? "none"}`);
    lines.push(`votes ${freshness?.votes?.latestDate ?? "none"}`);
  } catch (err) {
    freshnessError = err instanceof Error ? err.message : String(err);
    lines.push(`freshness error: ${freshnessError}`);
  }

  let upstreamSpeechDate = null;
  let upstreamSpeechError = null;
  let upstreamVoteDate = null;
  let upstreamVoteError = null;
  try {
    upstreamSpeechDate = await latestSpeechDate(rm);
    lines.push(`riksdagen speeches ${upstreamSpeechDate ?? "none"}`);
  } catch (err) {
    upstreamSpeechError = err instanceof Error ? err.message : String(err);
    lines.push(`riksdagen speeches error: ${upstreamSpeechError}`);
  }
  try {
    upstreamVoteDate = await latestVoteDate(rm);
    lines.push(`riksdagen votes ${upstreamVoteDate ?? "none"}`);
  } catch (err) {
    upstreamVoteError = err instanceof Error ? err.message : String(err);
    lines.push(`riksdagen votes error: ${upstreamVoteError}`);
  }

  const result = freshnessError
    ? { ok: false, reasons: [`freshness endpoint: ${freshnessError}`] }
    : evaluateFreshness({
        now,
        freshness,
        upstreamSpeechDate,
        upstreamVoteDate,
        upstreamSpeechError,
        upstreamVoteError,
      });

  if (!result.ok) {
    lines.push("", "stale:");
    for (const reason of result.reasons) lines.push(`- ${reason}`);
  } else {
    lines.push("", "fresh");
  }

  const report = lines.join("\n") + "\n";
  process.stdout.write(report);
  if (!result.ok) process.exitCode = 1;
}

const isMain = process.argv[1] && import.meta.url.endsWith(process.argv[1].replaceAll("\\", "/"));
if (isMain) {
  main().catch((err) => {
    console.error(err instanceof Error ? err.message : err);
    process.exitCode = 1;
  });
}
