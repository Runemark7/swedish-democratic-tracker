import { useQuery } from "@tanstack/react-query";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { partyRecordApi } from "./api";
import { PartyDecisions, PartyMotions } from "./PartyRecordSections";
import { politiciansApi } from "@/features/politicians/api";
import { regeringApi } from "@/features/regering/api";
import { PARTY_COLORS, partyShortToName } from "@/shared/design";
import { useSpeechesByParty } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { loftenApi } from "@/features/loften/api";
import { ArchiveNote, LoefteCard, ValmanifestList } from "@/features/loften/components";
import { SourceMarker } from "@/components/sources/SourceMarker";

type Tab = "beslut" | "motioner" | "loften" | "anforanden" | "politiker";

const TABS: { key: Tab; label: string }[] = [
  { key: "beslut", label: "Beslut" },
  { key: "motioner", label: "Kommittémotioner" },
  { key: "loften", label: "Löften" },
  { key: "anforanden", label: "Anföranden" },
  { key: "politiker", label: "Politiker" },
];

export function PartyDetailPage() {
  const { party = "" } = useParams<{ party: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const isMobile = useMediaQuery("(max-width: 640px)");

  // ?tab=mal was the Mål tab; the 2022 goals now live in the Löften archive.
  const rawTab = searchParams.get("tab") === "mal" ? "loften" : searchParams.get("tab");
  const tab: Tab = TABS.some((t) => t.key === rawTab) ? (rawTab as Tab) : "beslut";

  const setTab = (next: Tab) => {
    const params = new URLSearchParams(searchParams);
    if (next === "beslut") params.delete("tab");
    else params.set("tab", next);
    setSearchParams(params, { replace: true });
  };

  const { data: record, isLoading: loadingRecord, isError: recordError } = useQuery({
    queryKey: ["party-record", party],
    queryFn: () => partyRecordApi.get(party),
    enabled: !!party && (tab === "beslut" || tab === "motioner"),
    staleTime: 60_000,
  });

  const { data: loften, isLoading: loadingLoften, isError: loftenError } = useQuery({
    queryKey: ["loften", party],
    queryFn: () => loftenApi.list(party),
    enabled: !!party && tab === "loften",
    staleTime: 60_000,
  });

  // Always loaded: every party page lists its Valmanifest as sources.
  const { data: valmanifest } = useQuery({
    queryKey: ["valmanifest", party],
    queryFn: () => loftenApi.valmanifest(party),
    enabled: !!party,
    staleTime: 5 * 60_000,
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

  const { data: ministerGroups } = useQuery({
    queryKey: ["ministers"],
    queryFn: regeringApi.listMinisters,
    staleTime: 5 * 60_000,
  });
  const partyMinisters = (ministerGroups ?? [])
    .flatMap((g) => g.ministers ?? [])
    .filter((m) => m.party === party && m.active);

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
        </div>
      </header>

      {/* Every party page lists its Valmanifest as sources. */}
      <section
        aria-label="Valmanifest"
        data-testid="party-valmanifest"
        style={{ padding: isMobile ? "0 14px 16px" : "0 32px 20px", maxWidth: 920 }}
      >
        <h2 className="text-xs font-semibold text-on-surface-variant mb-2">
          Valmanifest <SourceMarker sourceId="party-manifestos" />
        </h2>
        <ValmanifestList items={valmanifest ?? []} />
      </section>

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
        {(tab === "beslut" || tab === "motioner") && (
          <section>
            {loadingRecord && (
              <div style={{ fontSize: 13, color: "var(--color-fg-muted)" }}>Laddar...</div>
            )}
            {recordError && (
              <p className="text-sm text-on-surface-variant">
                Kunde inte ladda {partyShortToName(party)}s {tab === "beslut" ? "beslut" : "kommittémotioner"}. Försök igen.
              </p>
            )}
            {record && tab === "beslut" && <PartyDecisions party={party} record={record} />}
            {record && tab === "motioner" && <PartyMotions party={party} record={record} />}
          </section>
        )}

        {tab === "loften" && (
          <section>
            {loadingLoften && (
              <div style={{ fontSize: 13, color: "var(--color-fg-muted)" }}>Laddar löften...</div>
            )}
            {loftenError && (
              <p className="text-sm text-on-surface-variant">Kunde inte ladda löften. Försök igen.</p>
            )}
            {loften && (
              <>
                <ArchiveNote />
                {loften.length === 0 ? (
                  <p style={{ fontStyle: "italic", color: "var(--color-fg-muted)", fontSize: 13 }}>
                    Vi har inga inlästa löften från {partyShortToName(party)}.{" "}
                    <SourceMarker sourceId="seed-party-goals" />
                  </p>
                ) : (
                  <div className="space-y-3 mt-4">
                    {loften.map((l) => <LoefteCard key={l.id} l={l} />)}
                  </div>
                )}
                <p className="mt-6 text-xs">
                  <Link to="/loften" className="text-link underline">Löften för alla partier</Link>
                </p>
              </>
            )}
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

      {partyMinisters.length > 0 && (
        <div
          className="mt-8"
          style={{ padding: isMobile ? "0 14px" : "0 32px", maxWidth: 920 }}
        >
          <h3 className="text-xs font-mono uppercase tracking-widest text-on-surface-variant mb-3">
            Statsråd från partiet
          </h3>
          <div className="bg-surface-lowest rounded-xl overflow-hidden">
            {partyMinisters.map((m, i) => (
              <Link
                key={m.id}
                to={`/regering/${m.id}`}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 12,
                  padding: "12px 20px",
                  borderTop: i > 0 ? "1px solid var(--color-surface-high)" : undefined,
                  textDecoration: "none",
                  color: "inherit",
                  transition: "background 0.15s",
                }}
                onMouseEnter={(e) => { e.currentTarget.style.background = "var(--color-surface-high)"; }}
                onMouseLeave={(e) => { e.currentTarget.style.background = "transparent"; }}
              >
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 14, fontWeight: 600 }}>{m.name}</div>
                  <div style={{ fontSize: 12, color: "var(--color-fg-muted)" }}>{m.title}</div>
                </div>
                <span style={{ fontSize: 12, color: "var(--color-fg-muted)" }}>→</span>
              </Link>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
