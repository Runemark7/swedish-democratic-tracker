import { useQuery } from "@tanstack/react-query";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { partiesApi } from "./api";
import { politiciansApi } from "@/features/politicians/api";
import { PARTY_COLORS, partyShortToName } from "@/shared/design";
import { useSpeechesByParty } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
import { useMediaQuery } from "@/hooks/useMediaQuery";

type Tab = "mal" | "anforanden" | "politiker";

const TABS: { key: Tab; label: string }[] = [
  { key: "mal", label: "Mål" },
  { key: "anforanden", label: "Anföranden" },
  { key: "politiker", label: "Politiker" },
];

export function PartyDetailPage() {
  const { party = "" } = useParams<{ party: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const isMobile = useMediaQuery("(max-width: 640px)");

  const tab: Tab =
    (searchParams.get("tab") as Tab | null) && TABS.some((t) => t.key === searchParams.get("tab"))
      ? (searchParams.get("tab") as Tab)
      : "mal";

  const setTab = (next: Tab) => {
    const params = new URLSearchParams(searchParams);
    if (next === "mal") params.delete("tab");
    else params.set("tab", next);
    setSearchParams(params, { replace: true });
  };

  const { data: parties } = useQuery({
    queryKey: ["parties"],
    queryFn: partiesApi.listParties,
    staleTime: 5 * 60 * 1000,
  });
  const summary = parties?.find((p) => p.party === party);

  const { data: goals, isLoading: loadingGoals } = useQuery({
    queryKey: ["party-goals", party],
    queryFn: () => partiesApi.listGoals(party),
    enabled: !!party && tab === "mal",
    staleTime: 60_000,
  });

  const { data: speeches, isLoading: loadingSpeeches } = useSpeechesByParty(
    tab === "anforanden" ? party : undefined,
    50,
  );

  const { data: politiciansResp, isLoading: loadingPoliticians } = useQuery({
    queryKey: ["politicians", { party, pageSize: 200 }],
    queryFn: () => politiciansApi.list({ party, pageSize: 200 }),
    enabled: !!party && tab === "politiker",
    staleTime: 60_000,
  });

  const pc = PARTY_COLORS[party];

  return (
    <div className="sdt-page" style={{ paddingBottom: 64 }}>
      {/* Back link */}
      <div style={{ padding: isMobile ? "16px 14px 0" : "32px 32px 0" }}>
        <Link
          to="/parties"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "var(--color-accent)",
            textDecoration: "none",
            letterSpacing: "0.1em",
          }}
        >
          ← Alla partier
        </Link>
      </div>

      {/* Hero */}
      <header
        style={{
          padding: isMobile ? "12px 14px 16px" : "16px 32px 24px",
          display: "flex",
          alignItems: "center",
          gap: 16,
        }}
      >
        <span
          style={{
            display: "inline-block",
            width: 6,
            height: 56,
            background: pc?.bg ?? "#888",
            flexShrink: 0,
          }}
        />
        <div style={{ minWidth: 0 }}>
          <div
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--color-fg-muted)",
              letterSpacing: "0.1em",
            }}
          >
            {party}
          </div>
          <h1
            style={{
              fontFamily: "var(--font-serif)",
              fontSize: isMobile ? 28 : 40,
              fontWeight: 400,
              margin: 0,
              lineHeight: 1.05,
              color: "var(--color-fg)",
            }}
          >
            {partyShortToName(party)}
          </h1>
          {summary && (
            <div
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                color: "var(--color-fg-muted)",
                marginTop: 6,
              }}
            >
              {summary.totalGoals ?? 0} mål · {summary.avgAlignmentPct ?? 0}% i linje
            </div>
          )}
        </div>
      </header>

      {/* Tab strip */}
      <div
        style={{
          display: "flex",
          gap: 0,
          padding: isMobile ? "0 14px" : "0 32px",
          borderBottom: "1px solid var(--color-border)",
          overflowX: "auto",
        }}
      >
        {TABS.map((t) => {
          const active = tab === t.key;
          return (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              style={{
                background: "transparent",
                border: "none",
                padding: "10px 16px",
                fontFamily: "var(--font-mono)",
                fontSize: 12,
                fontWeight: active ? 700 : 500,
                color: active ? "var(--color-fg)" : "var(--color-fg-muted)",
                borderBottom: active ? "2px solid var(--color-fg)" : "2px solid transparent",
                cursor: "pointer",
                whiteSpace: "nowrap",
                marginBottom: -1,
                letterSpacing: "0.05em",
              }}
            >
              {t.label}
            </button>
          );
        })}
      </div>

      {/* Tab body */}
      <main
        style={{
          padding: isMobile ? "16px 14px 0" : "24px 32px 0",
          maxWidth: 920,
        }}
      >
        {tab === "mal" && (
          <section>
            {loadingGoals && (
              <div style={{ fontSize: 13, color: "var(--color-fg-muted)" }}>Laddar mål...</div>
            )}
            {!loadingGoals && (!goals || goals.length === 0) && (
              <p style={{ fontStyle: "italic", color: "var(--color-fg-muted)", fontSize: 13 }}>
                Inga registrerade mål för {partyShortToName(party)}.
              </p>
            )}
            {goals && goals.length > 0 && (
              <ul style={{ listStyle: "none", margin: 0, padding: 0 }}>
                {goals.map((g) => (
                  <li
                    key={g.id}
                    style={{
                      borderBottom: "1px solid var(--color-border)",
                      padding: "12px 0",
                    }}
                  >
                    <div style={{ fontFamily: "var(--font-serif)", fontSize: 15, color: "var(--color-fg)" }}>
                      {g.goalText}
                    </div>
                    <div
                      style={{
                        fontFamily: "var(--font-mono)",
                        fontSize: 11,
                        color: "var(--color-fg-muted)",
                        marginTop: 4,
                      }}
                    >
                      {g.topic} · {g.alignmentPct ?? 0}% i linje · {g.relevantVotes ?? 0} röster
                    </div>
                  </li>
                ))}
              </ul>
            )}
            <Link
              to={`/parties/${party}/goals`}
              style={{
                display: "inline-block",
                marginTop: 16,
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                color: "var(--color-accent)",
                textDecoration: "none",
              }}
            >
              Visa fullständig målvy →
            </Link>
          </section>
        )}

        {tab === "anforanden" && (
          <section>
            {loadingSpeeches && (
              <div style={{ fontSize: 13, color: "var(--color-fg-muted)" }}>Laddar anföranden...</div>
            )}
            {!loadingSpeeches && (!speeches || speeches.length === 0) && (
              <p style={{ fontStyle: "italic", color: "var(--color-fg-muted)", fontSize: 13 }}>
                Inga registrerade anföranden från {partyShortToName(party)} ännu.
              </p>
            )}
            {speeches && speeches.length > 0 && (
              <div>
                {speeches.map((s) => (
                  <SpeechRow key={s.id} speech={s} />
                ))}
              </div>
            )}
          </section>
        )}

        {tab === "politiker" && (
          <section>
            {loadingPoliticians && (
              <div style={{ fontSize: 13, color: "var(--color-fg-muted)" }}>Laddar politiker...</div>
            )}
            {!loadingPoliticians &&
              (!politiciansResp || politiciansResp.data.length === 0) && (
                <p style={{ fontStyle: "italic", color: "var(--color-fg-muted)", fontSize: 13 }}>
                  Inga politiker hittades för {partyShortToName(party)}.
                </p>
              )}
            {politiciansResp && politiciansResp.data.length > 0 && (
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: isMobile
                    ? "minmax(0, 1fr)"
                    : "repeat(auto-fill, minmax(220px, 1fr))",
                  gap: 8,
                }}
              >
                {politiciansResp.data.map((p) => (
                  <Link
                    key={p.intressentId}
                    to={`/politicians/${p.intressentId}`}
                    style={{
                      background: "var(--color-sdt-surface)",
                      border: "1px solid var(--color-border)",
                      padding: "10px 12px",
                      borderRadius: 4,
                      textDecoration: "none",
                      color: "inherit",
                      display: "flex",
                      alignItems: "center",
                      gap: 10,
                      minWidth: 0,
                    }}
                  >
                    {p.imageUrl ? (
                      <img
                        src={p.imageUrl}
                        alt=""
                        style={{
                          width: 32,
                          height: 32,
                          borderRadius: "50%",
                          objectFit: "cover",
                          flexShrink: 0,
                        }}
                      />
                    ) : (
                      <span
                        style={{
                          width: 32,
                          height: 32,
                          borderRadius: "50%",
                          background: pc?.bg ?? "#888",
                          color: pc?.text ?? "#fff",
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",
                          fontSize: 11,
                          fontFamily: "var(--font-mono)",
                          fontWeight: 700,
                          flexShrink: 0,
                        }}
                      >
                        {p.firstName[0]}{p.lastName[0]}
                      </span>
                    )}
                    <div style={{ minWidth: 0 }}>
                      <div
                        style={{
                          fontFamily: "var(--font-serif)",
                          fontSize: 14,
                          color: "var(--color-fg)",
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                          whiteSpace: "nowrap",
                        }}
                      >
                        {p.firstName} {p.lastName}
                      </div>
                      {p.constituency && (
                        <div
                          style={{
                            fontFamily: "var(--font-mono)",
                            fontSize: 10,
                            color: "var(--color-fg-muted)",
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                            whiteSpace: "nowrap",
                          }}
                        >
                          {p.constituency}
                        </div>
                      )}
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </section>
        )}
      </main>
    </div>
  );
}
