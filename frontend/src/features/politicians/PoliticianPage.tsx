import { useQuery } from "@tanstack/react-query";
import { useParams, Link, useSearchParams } from "react-router-dom";
import { politiciansApi } from "./api";
import { PartyBadge, TopicTag, SpecificityBadge } from "@/shared/components";
import { PARTY_COLORS } from "@/shared/design";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { useSpeechesByPolitician, useDocument } from "@/hooks/useDemocracy";
import { SpeechRow } from "@/features/speeches/SpeechRow";
import type { Vote, PromiseWithMatches } from "@/shared/types";
import type { Speech } from "@/features/speeches/api";

const ALIGNMENT_STYLES: Record<string, { color: string; bg: string; label: string }> = {
  supports:    { color: "#16a34a", bg: "#f0fdf4", label: "Stödjer" },
  contradicts: { color: "#dc2626", bg: "#fef2f2", label: "Motsäger" },
  unclear:     { color: "#9ca3af", bg: "#f9fafb", label: "Oklart" },
};

function PromiseCard({ promise }: { promise: PromiseWithMatches }) {
  return (
    <div className="bg-surface-lowest rounded-xl p-5">
      <div className="flex flex-wrap items-center gap-2 mb-2">
        <TopicTag topic={promise.topic} />
        <SpecificityBadge specificity={promise.specificity} />
        {promise.speechDate && (
          <span className="text-[11px] text-on-surface-variant">{promise.speechDate}</span>
        )}
      </div>
      <p className="text-sm text-on-surface leading-relaxed mb-3">&ldquo;{promise.promiseText}&rdquo;</p>
      {promise.speechTopic && (
        <p className="text-xs text-on-surface-variant italic mb-3">Ämne: {promise.speechTopic}</p>
      )}
      {(promise.voteMatches ?? []).length > 0 && (
        <div className="space-y-2 pt-3" style={{ borderTop: "1px solid var(--color-surface-high)" }}>
          {promise.voteMatches!.map((m) => {
            const s = ALIGNMENT_STYLES[m.alignment] ?? ALIGNMENT_STYLES.unclear;
            return (
              <div key={m.voteId} className="flex items-start justify-between gap-3">
                <Link
                  to={`/votes/${m.beteckning}/${m.forslagspunkt}`}
                  className="text-xs text-on-surface-variant hover:text-primary transition-colors truncate flex-1"
                >
                  {m.documentTitle || `${m.beteckning} punkt ${m.forslagspunkt}`}
                </Link>
                <span
                  className="text-[10px] font-extrabold px-2 py-0.5 rounded shrink-0"
                  style={{ background: s.bg, color: s.color }}
                >
                  {s.label}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

type ActivityEvent =
  | { kind: "vote"; date: string; vote: Vote }
  | { kind: "speech"; date: string; speech: Speech };

function buildActivity(votes: Vote[], speeches: Speech[]): ActivityEvent[] {
  const events: ActivityEvent[] = [];
  for (const v of votes) {
    events.push({ kind: "vote", date: v.date ?? "", vote: v });
  }
  for (const s of speeches) {
    events.push({ kind: "speech", date: s.date, speech: s });
  }
  events.sort((a, b) => (b.date || "").localeCompare(a.date || ""));
  return events.slice(0, 50);
}

function ActivityFeed({
  votes,
  speeches,
  loading,
}: {
  votes: Vote[];
  speeches: Speech[];
  loading: boolean;
}) {
  if (loading) {
    return <div className="text-on-surface-variant text-sm py-6 text-center">Laddar...</div>;
  }
  const events = buildActivity(votes, speeches);
  if (events.length === 0) {
    return (
      <p className="text-sm italic text-on-surface-variant py-4">
        Ingen registrerad aktivitet ännu.
      </p>
    );
  }
  return (
    <div className="rounded-xl p-4" style={{ background: "var(--color-surface-lowest)" }}>
      <div className="text-[10px] font-mono uppercase tracking-widest text-on-surface-variant mb-3">
        Senaste 50 händelser
      </div>
      <div>
        {events.map((e, i) =>
          e.kind === "speech" ? (
            <SpeechRow key={`s-${e.speech.id}`} speech={e.speech} hidePolitician />
          ) : (
            <ActivityVoteRow key={`v-${i}`} vote={e.vote} />
          ),
        )}
      </div>
    </div>
  );
}

function ActivityVoteRow({ vote }: { vote: Vote }) {
  return (
    <Link
      to={`/votes/${vote.beteckning}/${vote.forslagspunkt}`}
      style={{
        display: "flex",
        gap: 12,
        padding: "12px 0",
        borderBottom: "1px solid var(--color-border)",
        alignItems: "flex-start",
        textDecoration: "none",
        color: "inherit",
        minWidth: 0,
      }}
    >
      <span
        style={{
          fontFamily: "var(--font-mono)",
          fontSize: 9,
          letterSpacing: "0.05em",
          background: "var(--color-track)",
          color: "var(--color-fg-muted)",
          padding: "2px 6px",
          flexShrink: 0,
          height: 18,
        }}
      >
        RÖST
      </span>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            fontSize: 13,
            color: "var(--color-fg)",
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
            marginBottom: 4,
          }}
          title={vote.documentTitle ?? `${vote.beteckning} punkt ${vote.forslagspunkt}`}
        >
          {vote.documentTitle || `${vote.beteckning} punkt ${vote.forslagspunkt}`}
        </div>
        <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)" }}>
          {vote.beteckning} · {vote.voteResult}
          {vote.date ? ` · ${vote.date}` : ""}
        </div>
      </div>
    </Link>
  );
}

function DebatterPanel({ speeches, loading }: { speeches: Speech[]; loading: boolean }) {
  // Group by relatedDokId, sort by count desc, take top 20.
  const groups = new Map<string, { count: number; minDate: string; maxDate: string }>();
  for (const s of speeches) {
    if (!s.relatedDokId) continue;
    const g = groups.get(s.relatedDokId);
    if (!g) {
      groups.set(s.relatedDokId, { count: 1, minDate: s.date, maxDate: s.date });
    } else {
      g.count++;
      if (s.date < g.minDate) g.minDate = s.date;
      if (s.date > g.maxDate) g.maxDate = s.date;
    }
  }
  const top = Array.from(groups.entries())
    .sort((a, b) => b[1].count - a[1].count)
    .slice(0, 20);

  if (loading) {
    return <div className="text-on-surface-variant text-sm py-6 text-center">Laddar...</div>;
  }
  if (top.length === 0) {
    return (
      <p className="text-sm italic text-on-surface-variant py-4">
        Inga debatter har kopplats till anförandena ännu.
      </p>
    );
  }
  return (
    <div className="rounded-xl p-4" style={{ background: "var(--color-surface-lowest)" }}>
      <div className="text-[10px] font-mono uppercase tracking-widest text-on-surface-variant mb-3">
        {top.length} aktiva debatter
      </div>
      <div>
        {top.map(([dokId, info]) => (
          <DebatterRow key={dokId} dokId={dokId} count={info.count} minDate={info.minDate} maxDate={info.maxDate} />
        ))}
      </div>
    </div>
  );
}

function DebatterRow({
  dokId,
  count,
  minDate,
  maxDate,
}: {
  dokId: string;
  count: number;
  minDate: string;
  maxDate: string;
}) {
  const { data: doc } = useDocument(dokId);
  const title = doc?.title || dokId;
  const typeLabel = doc?.type ? doc.type.toUpperCase() : "";
  const link = doc?.type === "bet" && doc.beteckning ? `/beslut/${encodeURIComponent(doc.beteckning)}` : null;
  const range = minDate === maxDate ? minDate.slice(0, 10) : `${minDate.slice(0, 10)} → ${maxDate.slice(0, 10)}`;

  const body = (
    <div
      style={{
        display: "flex",
        gap: 12,
        padding: "12px 0",
        borderBottom: "1px solid var(--color-border)",
        alignItems: "flex-start",
        minWidth: 0,
      }}
    >
      {typeLabel && (
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 9,
            letterSpacing: "0.05em",
            background: "var(--color-track)",
            color: "var(--color-fg-muted)",
            padding: "2px 6px",
            flexShrink: 0,
            height: 18,
          }}
        >
          {typeLabel}
        </span>
      )}
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            fontSize: 14,
            color: "var(--color-fg)",
            fontFamily: "var(--font-serif)",
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
            marginBottom: 4,
          }}
          title={title}
        >
          {title}
        </div>
        <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)" }}>
          {count} anförande{count === 1 ? "" : "n"} · {range}
        </div>
      </div>
    </div>
  );

  return link ? (
    <Link to={link} style={{ textDecoration: "none", color: "inherit", display: "block" }}>
      {body}
    </Link>
  ) : (
    body
  );
}

function AmnenPanel({ speeches, loading }: { speeches: Speech[]; loading: boolean }) {
  const counts = new Map<string, { count: number; lastDate: string }>();
  for (const s of speeches) {
    const t = (s.topicHeading ?? "").trim();
    if (!t) continue;
    const c = counts.get(t);
    if (!c) {
      counts.set(t, { count: 1, lastDate: s.date });
    } else {
      c.count++;
      if (s.date > c.lastDate) c.lastDate = s.date;
    }
  }
  const top = Array.from(counts.entries())
    .sort((a, b) => b[1].count - a[1].count)
    .slice(0, 10);

  if (loading) {
    return <div className="text-on-surface-variant text-sm py-6 text-center">Laddar...</div>;
  }
  if (top.length === 0) {
    return (
      <p className="text-sm italic text-on-surface-variant py-4">
        Inga teman går att utläsa ännu.
      </p>
    );
  }
  return (
    <div className="rounded-xl p-4" style={{ background: "var(--color-surface-lowest)" }}>
      <div className="text-[10px] font-mono uppercase tracking-widest text-on-surface-variant mb-3">
        Återkommande teman
      </div>
      <div>
        {top.map(([topic, info]) => (
          <div
            key={topic}
            style={{
              padding: "10px 0",
              borderBottom: "1px solid var(--color-border)",
              minWidth: 0,
            }}
          >
            <div
              style={{
                fontSize: 13,
                color: "var(--color-fg)",
                marginBottom: 4,
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
              }}
              title={topic}
            >
              {topic}
            </div>
            <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)" }}>
              {info.count} anförande{info.count === 1 ? "" : "n"} · senast {info.lastDate.slice(0, 10)}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

type Tab = "aktivitet" | "debatter" | "amnen" | "loften";
const ALL_TABS: Tab[] = ["aktivitet", "debatter", "amnen", "loften"];

export function PoliticianPage() {
  const { id = "" } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const tabParam = searchParams.get("tab") as Tab | null;
  const activeTab: Tab = tabParam && ALL_TABS.includes(tabParam) ? tabParam : "aktivitet";
  const setActiveTab = (next: Tab) => {
    const params = new URLSearchParams(searchParams);
    if (next === "aktivitet") params.delete("tab");
    else params.set("tab", next);
    setSearchParams(params, { replace: true });
  };

  const { data: politician, isLoading: loadingPolitician } = useQuery({
    queryKey: ["politician", id],
    queryFn: () => politiciansApi.getById(id),
    enabled: !!id,
  });

  const { data: votesData, isLoading: loadingVotes } = useQuery({
    queryKey: ["politician-votes", id],
    queryFn: () => politiciansApi.listVotes(id, { pageSize: 100 }),
    enabled: !!id,
  });

  const { data: promises, isLoading: loadingPromises } = useQuery({
    queryKey: ["politician-promises", id],
    queryFn: () => politiciansApi.listPromises(id),
    enabled: !!id && activeTab === "loften",
  });

  const { data: speeches, isLoading: loadingSpeeches } = useSpeechesByPolitician(
    id,
    200,
  );

  if (loadingPolitician) {
    return <div className="text-on-surface-variant py-16 text-center text-sm">Laddar...</div>;
  }
  if (!politician) {
    return <div className="text-contradiction py-16 text-center text-sm">Politiker hittades inte.</div>;
  }

  const pc = PARTY_COLORS[politician.party];
  const votes = votesData?.data ?? [];

  const ja = votes.filter((v) => v.voteResult === "Ja").length;
  const nej = votes.filter((v) => v.voteResult === "Nej").length;
  const avstar = votes.filter((v) => v.voteResult === "Avstår").length;
  const franvarande = votes.filter((v) => v.voteResult === "Frånvarande").length;

  return (
    <div>
      {/* ── Profile header ─────────────────────────────────────────── */}
      <div
        className="rounded-xl p-6 mt-3 mb-5"
        style={{ background: pc?.light ?? "#f5f5f5" }}
      >
        <div className="flex items-center gap-5">
          {politician.imageUrl ? (
            <img
              src={politician.imageUrl}
              alt={politician.firstName}
              className="w-20 h-20 rounded-full object-cover shrink-0"
              loading="lazy"
            />
          ) : (
            <div
              className="w-20 h-20 rounded-full flex items-center justify-center shrink-0 text-xl font-extrabold"
              style={{ background: pc?.bg ?? "#666", color: pc?.text ?? "#fff" }}
            >
              {politician.firstName[0]}{politician.lastName[0]}
            </div>
          )}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-1">
              <h2 className="font-display text-[22px] font-extrabold tracking-tight m-0">
                {politician.firstName} {politician.lastName}
              </h2>
              <PartyBadge party={politician.party} size="lg" />
            </div>
            {politician.constituency && (
              <p className="text-sm text-on-surface-variant">{politician.constituency}</p>
            )}
          </div>
        </div>

        {/* Stats row */}
        {votes.length > 0 && (
          <div className="flex gap-6 mt-5 pt-4" style={{ borderTop: "1px solid rgba(0,0,0,0.06)" }}>
            {[
              { label: "Ja", value: ja, color: "#16a34a" },
              { label: "Nej", value: nej, color: "#dc2626" },
              { label: "Avstår", value: avstar, color: "#d97706" },
              { label: "Frånv.", value: franvarande, color: "#9ca3af" },
            ].map((s) => (
              <div key={s.label} className="text-center">
                <div className="font-mono font-extrabold text-lg" style={{ color: s.color }}>
                  {s.value}
                  <SourceMarker sourceId="riksdagen" />
                </div>
                <div className="text-[10px] text-on-surface-variant uppercase tracking-widest">
                  {s.label}
                </div>
              </div>
            ))}
            <div className="text-center">
              <div className="font-mono font-extrabold text-lg text-on-surface">
                {votes.length}
                <SourceMarker sourceId="riksdagen" />
              </div>
              <div className="text-[10px] text-on-surface-variant uppercase tracking-widest">
                Totalt
              </div>
            </div>
          </div>
        )}
      </div>

      {/* ── Tabs ───────────────────────────────────────────────────── */}
      <div className="flex gap-1 mb-5" style={{ borderBottom: "1px solid var(--color-surface-high)" }}>
        {ALL_TABS.map((tab) => {
          const label =
            tab === "aktivitet" ? "Aktivitet"
            : tab === "debatter" ? "Debatter"
            : tab === "amnen" ? "Ämnen"
            : "Löften";
          const active = activeTab === tab;
          return (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className="px-4 py-2.5 text-sm font-medium transition-colors cursor-pointer"
              style={{
                borderBottom: active ? "2px solid var(--color-on-surface)" : "2px solid transparent",
                color: active ? "var(--color-on-surface)" : "var(--color-on-surface-variant)",
                fontWeight: active ? 700 : 500,
                marginBottom: "-1px",
              }}
            >
              {label}
            </button>
          );
        })}
      </div>

      {/* ── Aktivitet ───────────────────────────────────────────────── */}
      {activeTab === "aktivitet" && (
        <ActivityFeed votes={votes} speeches={speeches ?? []} loading={loadingVotes || loadingSpeeches} />
      )}

      {/* ── Debatter ────────────────────────────────────────────────── */}
      {activeTab === "debatter" && (
        <DebatterPanel speeches={speeches ?? []} loading={loadingSpeeches} />
      )}

      {/* ── Ämnen ───────────────────────────────────────────────────── */}
      {activeTab === "amnen" && (
        <AmnenPanel speeches={speeches ?? []} loading={loadingSpeeches} />
      )}

      {/* ── Löften ──────────────────────────────────────────────────── */}
      {activeTab === "loften" && (
        loadingPromises ? (
          <div className="text-on-surface-variant text-center py-12 text-sm">Laddar löften...</div>
        ) : (
          <>
            <div className="space-y-3">
              {(promises ?? []).map((p) => (
                <PromiseCard key={p.id} promise={p} />
              ))}
            </div>
            {(promises ?? []).length === 0 && (
              <p className="text-on-surface-variant text-center py-16 text-sm">
                Inga löften extraherade ännu.
              </p>
            )}
          </>
        )
      )}
    </div>
  );
}
