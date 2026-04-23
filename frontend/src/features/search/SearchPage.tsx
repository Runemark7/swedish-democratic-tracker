import { mockRiksdag, mockKommun } from "@/mock/democracy";
import { Pill } from "@/components/charts";

const TOPICS = [
  { t: "Skola & utbildning",  n: 24, lvl: ["I", "III"] },
  { t: "Vård & omsorg",       n: 31, lvl: ["I", "II"]  },
  { t: "Kollektivtrafik",     n: 12, lvl: ["II", "III"] },
  { t: "Bostad & planering",  n: 18, lvl: ["I", "III"] },
  { t: "Skatt & ekonomi",     n: 22, lvl: ["I"]        },
  { t: "Miljö & klimat",      n: 15, lvl: ["I", "II", "III"] },
  { t: "Arbetsmarknad",       n: 9,  lvl: ["I"]        },
  { t: "Trygghet & brott",    n: 14, lvl: ["I", "III"] },
];

const allParties = [
  ...mockRiksdag.ruling.parties,
  ...(mockRiksdag.ruling.support ?? []),
  ...mockRiksdag.ruling.opposition,
];

const recentVotes = mockKommun.liveVotes.slice(0, 3);

function statusTone(status: string): "pass" | "fail" | "pending" | "neutral" {
  if (status === "Bifall") return "pass";
  if (status === "Avslag") return "fail";
  if (status === "Återremiss") return "pending";
  return "neutral";
}

export function SearchPage() {
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
          <em
            style={{
              fontStyle: "italic",
              color: "var(--color-accent-2)",
            }}
          >
            händer
          </em>{" "}
          om…
        </h1>

        {/* ── Search input ─────────────────────────────────────── */}
        <div
          style={{
            position: "relative",
            marginBottom: 48,
          }}
        >
          <input
            type="text"
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
          <div
            style={{
              background: "var(--color-sdt-surface)",
              padding: 24,
            }}
          >
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
              {TOPICS.map((topic) => (
                <button
                  key={topic.t}
                  style={{
                    background: "var(--color-plane)",
                    border: "1px solid var(--color-border)",
                    borderRadius: 4,
                    padding: "10px 12px",
                    textAlign: "left",
                    cursor: "pointer",
                    color: "var(--color-fg)",
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
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 6,
                    }}
                  >
                    <span
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 10,
                        letterSpacing: "0.1em",
                        textTransform: "uppercase",
                        color: "var(--color-fg-muted)",
                      }}
                    >
                      {topic.n} frågor
                    </span>
                    <span style={{ color: "var(--color-border)" }}>·</span>
                    {topic.lvl.map((l) => (
                      <em
                        key={l}
                        style={{
                          fontStyle: "italic",
                          color: "var(--color-accent-2)",
                          fontSize: 11,
                          fontFamily: "var(--font-serif)",
                        }}
                      >
                        {l}
                      </em>
                    ))}
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* RIGHT: PARTIER + SENASTE */}
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
                {allParties.map((p) => (
                  <button
                    key={p.short}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 6,
                      background: "var(--color-plane)",
                      border: "1px solid var(--color-border)",
                      borderRadius: 999,
                      padding: "5px 12px",
                      cursor: "pointer",
                      color: "var(--color-fg)",
                    }}
                  >
                    <span
                      style={{
                        width: 8,
                        height: 8,
                        borderRadius: "50%",
                        background: p.color,
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
                    <span
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 10,
                        color: "var(--color-fg-muted)",
                      }}
                    >
                      {p.seats}
                    </span>
                  </button>
                ))}
              </div>
            </div>

            {/* Recent votes from municipality */}
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
                Senaste från din kommun
              </p>
              <div style={{ display: "flex", flexDirection: "column", gap: 1 }}>
                {recentVotes.map((v, i) => (
                  <div
                    key={i}
                    style={{
                      display: "flex",
                      alignItems: "flex-start",
                      justifyContent: "space-between",
                      gap: 12,
                      padding: "10px 0",
                      borderBottom:
                        i < recentVotes.length - 1
                          ? "1px solid var(--color-border)"
                          : "none",
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
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
