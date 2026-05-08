import { Link } from "react-router-dom";
import { useRiksdag, useRecentSpeeches } from "@/hooks/useDemocracy";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { Pill } from "@/components/charts";
import { AgendaList } from "@/components/AgendaList";
import { PanelCard } from "./components/PanelCard";
import { PartySpeechCard } from "./components/PartySpeechCard";
import type { LiveVote } from "@/types/democracy";
import type { Speech } from "@/features/speeches/api";

const PARTY_ORDER = ["S", "M", "SD", "C", "V", "KD", "L", "MP"] as const;

function pillTone(status: string): "pass" | "fail" | "pending" | "neutral" {
  if (status === "Bifall") return "pass";
  if (status === "Avslag") return "fail";
  if (status === "Återremiss") return "pending";
  return "neutral";
}

function beslutHref(v: LiveVote): string | null {
  if (!v.beteckning) return null;
  const p = new URLSearchParams({ title: v.title, status: v.status, tag: v.tag, time: v.time });
  return `/beslut/${encodeURIComponent(v.beteckning)}?${p}`;
}

function isToday(time: string): boolean {
  return time.toLowerCase().startsWith("idag");
}

function isThisWeek(time: string): boolean {
  const t = time.toLowerCase();
  if (t.startsWith("idag") || t.startsWith("igår")) return true;
  const m = t.match(/^(\d{1,2})\s+(\S+)/);
  if (!m) return false;
  const day = Number(m[1]);
  const now = new Date();
  const diffDays = Math.abs(now.getDate() - day);
  return diffDays <= 7;
}

function pickLatestPerParty(speeches: Speech[]): Record<string, Speech | null> {
  const out: Record<string, Speech | null> = {};
  for (const p of PARTY_ORDER) out[p] = null;
  for (const s of speeches) {
    if (!(s.party in out)) continue;
    const existing = out[s.party];
    if (!existing || s.date > existing.date) out[s.party] = s;
  }
  return out;
}

export function HomePage() {
  const isMobile = useMediaQuery("(max-width: 640px)");
  const { data: riksdag, isLoading: riksdagLoading } = useRiksdag();
  const { data: speeches, isLoading: speechesLoading } = useRecentSpeeches(100);

  if (riksdagLoading || !riksdag) {
    return (
      <div className="sdt-page" style={{ padding: isMobile ? "20px 14px" : "40px 32px" }}>
        <div
          style={{
            height: 60,
            background: "var(--color-track)",
            borderRadius: 4,
            marginBottom: 24,
            maxWidth: 480,
          }}
        />
        <div style={{ height: 320, background: "var(--color-track)", borderRadius: 4 }} />
      </div>
    );
  }

  const liveVotes = riksdag.liveVotes ?? [];
  const agenda = riksdag.agenda ?? [];

  const todayVotes = liveVotes.filter((v) => isToday(v.time)).slice(0, 5);
  const weekVotes = liveVotes.filter((v) => isThisWeek(v.time)).slice(0, 5);

  const recentSpeeches = (speeches ?? []).slice(0, 5);
  const latestByParty = pickLatestPerParty(speeches ?? []);

  const panelGrid: React.CSSProperties = {
    display: "grid",
    gap: 1,
    gridTemplateColumns: isMobile ? "minmax(0, 1fr)" : "minmax(0, 1fr) minmax(0, 1fr)",
    background: "var(--color-border)",
    border: "1px solid var(--color-border)",
    margin: isMobile ? "14px 14px 0" : "22px 32px 0",
  };

  const partyGridStyle: React.CSSProperties = {
    display: "grid",
    gridTemplateColumns: isMobile
      ? "minmax(0, 1fr) minmax(0, 1fr)"
      : "repeat(4, minmax(0, 1fr))",
    gap: 12,
    margin: isMobile ? "14px 14px 28px" : "22px 32px 28px",
  };

  return (
    <div className="sdt-page">
      {/* ── Hero ─────────────────────────────────────────────────────── */}
      <header
        style={{
          padding: isMobile ? "20px 14px 16px" : "40px 32px 24px",
          display: "flex",
          flexDirection: "column",
          gap: 8,
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "2px",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
          }}
        >
          KAMMARE ETT · RIKSDAGEN
        </div>
        <h1
          style={{
            fontFamily: "var(--font-serif)",
            fontSize: isMobile ? 28 : 44,
            fontWeight: 400,
            letterSpacing: "-1.2px",
            color: "var(--color-fg)",
            margin: 0,
            lineHeight: 1.05,
          }}
        >
          Vad fokuserar riksdagen på?
        </h1>
        <p
          style={{
            fontFamily: "var(--font-body)",
            fontSize: 13,
            color: "var(--color-fg-muted)",
            margin: 0,
            maxWidth: 540,
          }}
        >
          Beslut, debatter och vad partierna driver — i realtid.
        </p>
      </header>

      {/* ── Panels 1 + 2 (row) ───────────────────────────────────────── */}
      <div style={panelGrid}>
        {/* Panel 1 — Beslut idag */}
        <PanelCard
          title="BESLUT IDAG"
          showAllHref="/votes"
          isEmpty={todayVotes.length === 0}
          emptyText="Inga beslut idag — kammaren kan ha sommaruppehåll."
        >
          {todayVotes.map((v, i) => {
            const href = beslutHref(v);
            const row = (
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "12px 1fr auto",
                  gap: 10,
                  alignItems: "center",
                  minWidth: 0,
                }}
              >
                <span
                  style={{
                    width: 7,
                    height: 7,
                    borderRadius: "50%",
                    background: i < 2 ? "var(--color-pulse)" : "var(--color-accent)",
                    display: "inline-block",
                  }}
                />
                <span
                  style={{
                    fontSize: 13,
                    color: "var(--color-fg)",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    minWidth: 0,
                  }}
                  title={v.title}
                >
                  {v.title}
                </span>
                <Pill tone={pillTone(v.status)}>{v.status}</Pill>
              </div>
            );
            return href ? (
              <Link key={i} to={href} style={{ textDecoration: "none", color: "inherit" }}>
                {row}
              </Link>
            ) : (
              <div key={i}>{row}</div>
            );
          })}
        </PanelCard>

        {/* Panel 2 — Aktuella debatter */}
        <PanelCard
          title="AKTUELLA DEBATTER"
          showAllHref="/politicians"
          isEmpty={!speechesLoading && recentSpeeches.length === 0}
          emptyText="Inga registrerade debatter ännu."
        >
          {recentSpeeches.map((s) => {
            const date = new Date(s.date);
            const stamp = `${String(date.getDate()).padStart(2, "0")}/${String(date.getMonth() + 1).padStart(2, "0")}`;
            return (
              <Link
                key={s.id}
                to={`/anforanden/${s.id}`}
                style={{
                  display: "grid",
                  gridTemplateColumns: "44px 1fr",
                  gap: 10,
                  alignItems: "baseline",
                  textDecoration: "none",
                  color: "inherit",
                  minWidth: 0,
                }}
              >
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 10,
                    color: "var(--color-fg-muted)",
                  }}
                >
                  {stamp}
                </span>
                <span
                  style={{
                    fontSize: 13,
                    color: "var(--color-fg)",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    minWidth: 0,
                  }}
                  title={s.politicianName}
                >
                  <strong style={{ fontWeight: 600 }}>{s.politicianName || "Anonym"}</strong>
                  <span style={{ color: "var(--color-fg-muted)", marginLeft: 6 }}>
                    [{s.party}]
                  </span>
                  {s.topicHeading && (
                    <span style={{ marginLeft: 8, color: "var(--color-fg-muted)" }}>
                      · {s.topicHeading}
                    </span>
                  )}
                </span>
              </Link>
            );
          })}
        </PanelCard>
      </div>

      {/* ── Panels 3 + 5 (row) ───────────────────────────────────────── */}
      <div style={{ ...panelGrid, marginTop: 1, borderTop: "none" }}>
        {/* Panel 3 — Veckans omröstningar */}
        <PanelCard
          title="VECKANS OMRÖSTNINGAR"
          showAllHref="/votes"
          isEmpty={weekVotes.length === 0}
          emptyText="Inga omröstningar denna vecka."
        >
          {weekVotes.map((v, i) => {
            const href = beslutHref(v);
            const row = (
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "60px 1fr auto",
                  gap: 10,
                  alignItems: "center",
                  minWidth: 0,
                }}
              >
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: 10,
                    color: "var(--color-fg-muted)",
                  }}
                >
                  {v.time}
                </span>
                <span
                  style={{
                    fontSize: 13,
                    color: "var(--color-fg)",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    minWidth: 0,
                  }}
                  title={v.title}
                >
                  {v.title}
                </span>
                <Pill tone={pillTone(v.status)}>{v.status}</Pill>
              </div>
            );
            return href ? (
              <Link key={i} to={href} style={{ textDecoration: "none", color: "inherit" }}>
                {row}
              </Link>
            ) : (
              <div key={i}>{row}</div>
            );
          })}
        </PanelCard>

        {/* Panel 5 — Kommande beslut */}
        <PanelCard
          title="KOMMANDE BESLUT"
          showAllHref="/votes"
          isEmpty={agenda.length === 0}
          emptyText="Ingen agenda publicerad."
        >
          <AgendaList items={agenda.slice(0, 5)} />
        </PanelCard>
      </div>

      {/* ── Panel 4 — Vad partierna säger (full width) ───────────────── */}
      <section
        style={{
          margin: isMobile ? "14px 14px 6px" : "22px 32px 6px",
        }}
      >
        <header
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 10,
            letterSpacing: "0.15em",
            color: "var(--color-fg-muted)",
            textTransform: "uppercase",
          }}
        >
          VAD PARTIERNA SÄGER
        </header>
      </section>
      <div style={partyGridStyle}>
        {PARTY_ORDER.map((p) => (
          <PartySpeechCard key={p} party={p} speech={latestByParty[p]} />
        ))}
      </div>
    </div>
  );
}
