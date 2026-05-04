import { useState } from "react";
import { Link } from "react-router-dom";
import { Pill } from "@/components/charts";
import { TC_PARTY_COLORS } from "@/types/democracy";
import type { LiveVote } from "@/types/democracy";

function beslutHref(v: LiveVote): string | null {
  if (!v.beteckning) return null;
  const p = new URLSearchParams({ title: v.title, status: v.status, tag: v.tag, time: v.time });
  return `/beslut/${encodeURIComponent(v.beteckning)}?${p}`;
}

const TOPICS = [
  { t: "Skola & utbildning",  n: 24, lvl: ["I", "III"], tag: "Skola"    },
  { t: "Vård & omsorg",       n: 31, lvl: ["I", "II"],  tag: "Vård"     },
  { t: "Kollektivtrafik",     n: 12, lvl: ["II", "III"], tag: "Trafik"  },
  { t: "Bostad & planering",  n: 18, lvl: ["I", "III"], tag: "Plan"     },
  { t: "Skatt & ekonomi",     n: 22, lvl: ["I"],        tag: "Skatt"    },
  { t: "Miljö & klimat",      n: 15, lvl: ["I", "II", "III"], tag: "Miljö" },
  { t: "Arbetsmarknad",       n: 9,  lvl: ["I"],        tag: "Arbete"   },
  { t: "Trygghet & brott",    n: 14, lvl: ["I", "III"], tag: "Trygghet" },
];

const allParties = Object.entries(TC_PARTY_COLORS).map(([short, color]) => ({ short, color }));

const allVotes: LiveVote[] = [];

function statusTone(status: string): "pass" | "fail" | "pending" | "neutral" {
  if (status === "Bifall") return "pass";
  if (status === "Avslag") return "fail";
  if (status === "Återremiss") return "pending";
  return "neutral";
}

export function SearchPage() {
  const [query, setQuery] = useState("");
  const [activeTopic, setActiveTopic] = useState<string | null>(null);
  const [activeParty, setActiveParty] = useState<string | null>(null);

  function selectTopic(topic: typeof TOPICS[number]) {
    if (activeTopic === topic.t) {
      setActiveTopic(null);
      setQuery("");
    } else {
      setActiveTopic(topic.t);
      setActiveParty(null);
      setQuery(topic.t);
    }
  }

  function selectParty(short: string) {
    if (activeParty === short) {
      setActiveParty(null);
      setQuery("");
    } else {
      setActiveParty(short);
      setActiveTopic(null);
      setQuery(short);
    }
  }

  const activeTag = activeTopic
    ? TOPICS.find((t) => t.t === activeTopic)?.tag
    : null;

  const visibleVotes = allVotes.filter((v) => {
    if (!query) return true;
    const q = query.toLowerCase();
    const matchesTag = activeTag ? v.tag?.toLowerCase() === activeTag.toLowerCase() : false;
    const matchesParty = activeParty ? v.tag?.toLowerCase().includes(q) || v.title.toLowerCase().includes(q) : false;
    const matchesText = v.title.toLowerCase().includes(q) || (v.tag ?? "").toLowerCase().includes(q);
    return matchesTag || matchesParty || matchesText;
  });

  return (
    <div
      style={{
        background: "var(--color-bg)",
        color: "var(--color-fg)",
        minHeight: "100vh",
        fontFamily: "var(--font-body)",
        paddingBottom: "64px",
      }}
    >
      {/* ── Header ───────────────────────────────────────────────── */}
      <div style={{ maxWidth: 1280, margin: "0 auto", padding: "48px 32px 0" }}>
        <p
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            letterSpacing: "0.1em",
            textTransform: "uppercase",
            color: "var(--color-fg-muted)",
            margin: "0 0 12px",
          }}
        >
          SÖK · UTFORSKA
        </p>

        <h1
          style={{
            fontFamily: "var(--font-serif)",
            fontSize: 48,
            fontWeight: 700,
            lineHeight: 1.1,
            margin: "0 0 32px",
            color: "var(--color-fg)",
          }}
        >
          Vad{" "}
          <em style={{ fontStyle: "italic", color: "var(--color-accent-2)" }}>
            händer
          </em>{" "}
          om…
        </h1>

        {/* ── Search input ─────────────────────────────────────── */}
        <div style={{ position: "relative", marginBottom: 48 }}>
          <input
            type="text"
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              if (!e.target.value) {
                setActiveTopic(null);
                setActiveParty(null);
              }
            }}
            placeholder="Sök parti, politiker, ämne eller fråga…"
            style={{
              width: "100%",
              boxSizing: "border-box",
              background: "var(--color-sdt-surface)",
              border: "1px solid var(--color-border)",
              borderRadius: 4,
              color: "var(--color-fg)",
              fontFamily: "var(--font-body)",
              fontSize: 16,
              padding: "14px 48px 14px 16px",
              outline: "none",
            }}
          />
          <span
            style={{
              position: "absolute",
              right: 16,
              top: "50%",
              transform: "translateY(-50%)",
              color: "var(--color-fg-muted)",
              fontSize: 20,
              pointerEvents: "none",
              fontFamily: "var(--font-mono)",
            }}
          >
            ⌕
          </span>
        </div>

        {/* ── Two-column grid ──────────────────────────────────── */}
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr",
            gap: 1,
            border: "1px solid var(--color-border)",
            borderRadius: 4,
            overflow: "hidden",
            background: "var(--color-border)",
          }}
        >
          {/* LEFT: ÄMNEN */}
          <div style={{ background: "var(--color-sdt-surface)", padding: 24 }}>
            <p
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                letterSpacing: "0.1em",
                textTransform: "uppercase",
                color: "var(--color-fg-muted)",
                margin: "0 0 16px",
              }}
            >
              Ämnen
            </p>

            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: 8,
              }}
            >
              {TOPICS.map((topic) => {
                const isActive = activeTopic === topic.t;
                return (
                  <button
                    key={topic.t}
                    onClick={() => selectTopic(topic)}
                    style={{
                      background: isActive ? "var(--color-accent)" : "var(--color-plane)",
                      border: `1px solid ${isActive ? "var(--color-accent)" : "var(--color-border)"}`,
                      borderRadius: 4,
                      padding: "10px 12px",
                      textAlign: "left",
                      cursor: "pointer",
                      color: isActive ? "#fff" : "var(--color-fg)",
                      transition: "background 0.15s, border-color 0.15s",
                    }}
                  >
                    <div
                      style={{
                        fontSize: 12,
                        fontWeight: 600,
                        marginBottom: 4,
                        fontFamily: "var(--font-body)",
                      }}
                    >
                      {topic.t}
                    </div>
                    <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
                      <span
                        style={{
                          fontFamily: "var(--font-mono)",
                          fontSize: 10,
                          letterSpacing: "0.1em",
                          textTransform: "uppercase",
                          color: isActive ? "rgba(255,255,255,0.7)" : "var(--color-fg-muted)",
                        }}
                      >
                        {topic.n} frågor
                      </span>
                      <span style={{ color: isActive ? "rgba(255,255,255,0.4)" : "var(--color-border)" }}>·</span>
                      {topic.lvl.map((l) => (
                        <em
                          key={l}
                          style={{
                            fontStyle: "italic",
                            color: isActive ? "rgba(255,255,255,0.85)" : "var(--color-accent-2)",
                            fontSize: 11,
                            fontFamily: "var(--font-serif)",
                          }}
                        >
                          {l}
                        </em>
                      ))}
                    </div>
                  </button>
                );
              })}
            </div>
          </div>

          {/* RIGHT: PARTIER + RESULTAT */}
          <div
            style={{
              background: "var(--color-sdt-surface)",
              padding: 24,
              display: "flex",
              flexDirection: "column",
              gap: 32,
            }}
          >
            {/* Parties */}
            <div>
              <p
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  letterSpacing: "0.1em",
                  textTransform: "uppercase",
                  color: "var(--color-fg-muted)",
                  margin: "0 0 16px",
                }}
              >
                Partier
              </p>
              <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
                {allParties.map((p) => {
                  const isActive = activeParty === p.short;
                  return (
                    <button
                      key={p.short}
                      onClick={() => selectParty(p.short)}
                      style={{
                        display: "flex",
                        alignItems: "center",
                        gap: 6,
                        background: isActive ? p.color : "var(--color-plane)",
                        border: `1px solid ${isActive ? p.color : "var(--color-border)"}`,
                        borderRadius: 999,
                        padding: "5px 12px",
                        cursor: "pointer",
                        color: isActive ? "#fff" : "var(--color-fg)",
                        transition: "background 0.15s, border-color 0.15s",
                      }}
                    >
                      <span
                        style={{
                          width: 8,
                          height: 8,
                          borderRadius: "50%",
                          background: isActive ? "rgba(255,255,255,0.8)" : p.color,
                          flexShrink: 0,
                        }}
                      />
                      <span
                        style={{
                          fontFamily: "var(--font-mono)",
                          fontSize: 11,
                          letterSpacing: "0.08em",
                          textTransform: "uppercase",
                        }}
                      >
                        {p.short}
                      </span>
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Results / recent votes */}
            <div>
              <p
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                  letterSpacing: "0.1em",
                  textTransform: "uppercase",
                  color: "var(--color-fg-muted)",
                  margin: "0 0 16px",
                }}
              >
                {query ? `Resultat för "${query}"` : "Senaste beslut"}
              </p>
              <div style={{ display: "flex", flexDirection: "column", gap: 1 }}>
                {(query ? visibleVotes : allVotes.slice(0, 4)).map((v, i) => {
                  const list = query ? visibleVotes : allVotes.slice(0, 4);
                  const href = beslutHref(v);
                  const row = (
                    <div
                      style={{
                        display: "flex",
                        alignItems: "flex-start",
                        justifyContent: "space-between",
                        gap: 12,
                        padding: "10px 0",
                        borderBottom: i < list.length - 1 ? "1px solid var(--color-border)" : "none",
                        cursor: href ? "pointer" : "default",
                      }}
                    >
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div
                          style={{
                            fontSize: 13,
                            fontWeight: 500,
                            color: "var(--color-fg)",
                            marginBottom: 2,
                            whiteSpace: "nowrap",
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                          }}
                        >
                          {v.title}
                        </div>
                        <div
                          style={{
                            fontFamily: "var(--font-mono)",
                            fontSize: 10,
                            letterSpacing: "0.08em",
                            textTransform: "uppercase",
                            color: "var(--color-fg-muted)",
                          }}
                        >
                          {v.time}
                          {v.tag ? ` · ${v.tag}` : ""}
                        </div>
                      </div>
                      <Pill tone={statusTone(v.status ?? "")}>{v.status}</Pill>
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
                {query && visibleVotes.length === 0 && (
                  <p
                    style={{
                      fontFamily: "var(--font-mono)",
                      fontSize: 12,
                      color: "var(--color-fg-muted)",
                      margin: 0,
                      padding: "12px 0",
                    }}
                  >
                    Inga resultat för "{query}"
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
