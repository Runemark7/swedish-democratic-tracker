import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useParams, Link, useSearchParams } from "react-router-dom";
import { politiciansApi } from "./api";
import { PartyBadge } from "@/shared/components";
import { PARTY_COLORS } from "@/shared/design";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { useSpeechesByPolitician, useDocument } from "@/hooks/useDemocracy";
import { MemberMotions } from "./MemberMotions";
import { QuestionsSection } from "@/features/questions/QuestionsSection";
import { SpeechRow } from "@/features/speeches/SpeechRow";
import type { components } from "@/shared/api-contract";
import type { Speech } from "@/features/speeches/api";

type Vote = components["schemas"]["Vote"];

const VOTES_PAGE_SIZE = 50;
const SPEECHES_SHOWN = 50;

// Votes and speeches are two lists, not one merged feed. Votes carry no
// decision date — the record holds only system_datum (when Riksdagen last
// touched the row) and createdAt (our insert time), and neither says when the
// chamber decided — so they cannot be interleaved with the dated speeches
// without inventing a chronology. Merged, the dated speeches filled the whole
// feed and a member who had spoken 50 times showed no vote at all (#350).
function ActivityFeed({
  votes,
  totalVotes,
  loadingVotes,
  votesFailed,
  hasMoreVotes,
  loadingMoreVotes,
  onMoreVotes,
  speeches,
  loadingSpeeches,
}: {
  votes: Vote[];
  totalVotes: number;
  loadingVotes: boolean;
  votesFailed: boolean;
  hasMoreVotes: boolean;
  loadingMoreVotes: boolean;
  onMoreVotes: () => void;
  speeches: Speech[];
  loadingSpeeches: boolean;
}) {
  const shownSpeeches = [...speeches]
    .sort((a, b) => (b.date || "").localeCompare(a.date || ""))
    .slice(0, SPEECHES_SHOWN);
  const heading = "text-[10px] font-mono uppercase tracking-widest text-on-surface-variant";
  return (
    <div className="flex flex-col gap-4">
      <section
        data-testid="member-votes"
        className="rounded-xl p-4"
        style={{ background: "var(--color-surface-lowest)" }}
      >
        <div className="flex flex-wrap items-baseline justify-between gap-2 mb-1">
          <h3 className={`${heading} m-0`}>Röster</h3>
          {votes.length > 0 && (
            <span className={heading}>
              Visar {votes.length} av {totalVotes}
            </span>
          )}
        </div>
        <p className="text-[11px] text-on-surface-variant mt-0 mb-2">
          Så röstade ledamoten i varje förslagspunkt. Senast inlästa först – registret anger
          riksmöte men inget beslutsdatum per röst. Varje rad länkar till beslutet.
        </p>
        {loadingVotes ? (
          <div className="text-on-surface-variant text-sm py-6 text-center">Laddar...</div>
        ) : votesFailed && votes.length === 0 ? (
          <p className="text-sm italic text-on-surface-variant py-4">
            Rösterna kunde inte hämtas just nu. Försök igen om en stund.
          </p>
        ) : votes.length === 0 ? (
          <p className="text-sm italic text-on-surface-variant py-4">Inga registrerade röster.</p>
        ) : (
          <>
            <div>
              {votes.map((v) => (
                <ActivityVoteRow key={`${v.voteringId}-${v.politicianId}`} vote={v} />
              ))}
            </div>
            {hasMoreVotes && (
              <button
                type="button"
                onClick={onMoreVotes}
                disabled={loadingMoreVotes}
                className="mt-3 px-4 py-2 text-sm font-medium rounded-lg cursor-pointer"
                style={{
                  border: "1px solid var(--color-border)",
                  background: "transparent",
                  color: "var(--color-fg)",
                }}
              >
                {loadingMoreVotes ? "Laddar..." : "Visa fler röster"}
              </button>
            )}
          </>
        )}
      </section>

      <section
        data-testid="member-speeches"
        className="rounded-xl p-4"
        style={{ background: "var(--color-surface-lowest)" }}
      >
        <h3 className={`${heading} m-0 mb-3`}>
          {shownSpeeches.length >= SPEECHES_SHOWN
            ? `Senaste ${SPEECHES_SHOWN} anförandena`
            : "Anföranden"}
        </h3>
        {loadingSpeeches ? (
          <div className="text-on-surface-variant text-sm py-6 text-center">Laddar...</div>
        ) : shownSpeeches.length === 0 ? (
          <p className="text-sm italic text-on-surface-variant py-4">Inga registrerade anföranden.</p>
        ) : (
          <div>
            {shownSpeeches.map((s) => (
              <SpeechRow key={`s-${s.id}`} speech={s} hidePolitician />
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

function ActivityVoteRow({ vote }: { vote: Vote }) {
  const fallbackTitle = `${vote.beteckning} punkt ${vote.forslagspunkt}`;
  return (
    <Link
      to={`/votes/${vote.beteckning}/${vote.forslagspunkt}?session=${encodeURIComponent(vote.session)}`}
      data-testid="member-vote-row"
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
          title={vote.proposalOrigin.documentTitle ?? fallbackTitle}
        >
          {vote.proposalOrigin.documentTitle || fallbackTitle}
        </div>
        {/* No date: a vote row carries only system_datum, which is when
            Riksdagen last touched the record, not when the chamber decided. */}
        <div style={{ fontFamily: "var(--font-mono)", fontSize: 10, color: "var(--color-fg-muted)" }}>
          {vote.session}:{vote.beteckning} punkt {vote.forslagspunkt} ·{" "}
          <span data-testid="member-vote-result" style={{ fontWeight: 700, color: "var(--color-fg)" }}>
            {vote.voteResult}
          </span>
          {vote.avvikare && (
            <span
              data-testid="avvikare-mark"
              title={`Röstade ${vote.voteResult}; partilinjen var ${vote.partilinje}`}
              style={{ marginLeft: 8, fontWeight: 700, color: "var(--color-fg)" }}
            >
              ◆ Avvikare från partilinjen ({vote.partilinje})
            </span>
          )}
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

type Tab = "aktivitet" | "motioner" | "debatter" | "amnen";
const ALL_TABS: Tab[] = ["aktivitet", "motioner", "debatter", "amnen"];

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

  const votesQuery = useInfiniteQuery({
    queryKey: ["politician-votes", id],
    queryFn: ({ pageParam }) =>
      politiciansApi.listVotes(id, { page: pageParam, pageSize: VOTES_PAGE_SIZE }),
    initialPageParam: 1,
    getNextPageParam: (last) =>
      last.page * last.pageSize < last.total ? last.page + 1 : undefined,
    enabled: !!id,
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
  const votePages = votesQuery.data?.pages ?? [];
  const votes = votePages.flatMap((p) => p.data ?? []);

  // Counts come from the server's aggregate over the member's whole recorded
  // history — the list below is paged and must not be counted.
  const summary = votePages[0]?.summary;
  const ja = summary?.ja ?? 0;
  const nej = summary?.nej ?? 0;
  const avstar = summary?.avstar ?? 0;
  const franvarande = summary?.franvarande ?? 0;
  const totalVotes = summary?.total ?? 0;
  const pct = (n: number) =>
    totalVotes > 0 ? `${((n / totalVotes) * 100).toFixed(1).replace(".", ",")} %` : "";

  return (
    <div>
      {/* ── Profile header ─────────────────────────────────────────── */}
      <div
        className="rounded-xl p-6 mt-3 mb-5"
        style={{ background: pc?.light ?? "var(--color-surface-low)" }}
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
        {totalVotes > 0 && (
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
                <div className="text-[11px] text-on-surface-variant">{pct(s.value)}</div>
                <div className="text-[10px] text-on-surface-variant uppercase tracking-widest">
                  {s.label}
                </div>
              </div>
            ))}
            <div className="text-center">
              <div className="font-mono font-extrabold text-lg text-on-surface">
                {totalVotes}
                <SourceMarker sourceId="riksdagen" />
              </div>
              <div className="text-[10px] text-on-surface-variant uppercase tracking-widest">
                Totalt
              </div>
            </div>
            <div className="text-[11px] text-on-surface-variant self-center">
              Gäller alla {totalVotes} röstningar i registret (hela ledamotens registrerade
              historik), inte bara de som visas i listan nedan.
            </div>
          </div>
        )}
      </div>

      {/* ── Tabs ───────────────────────────────────────────────────── */}
      <div className="flex gap-1 mb-5" style={{ borderBottom: "1px solid var(--color-surface-high)" }}>
        {ALL_TABS.map((tab) => {
          const label =
            tab === "aktivitet" ? "Aktivitet"
            : tab === "motioner" ? "Motioner"
            : tab === "debatter" ? "Frågor och debatter"
            : "Ämnen";
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
        <ActivityFeed
          votes={votes}
          totalVotes={votePages[0]?.total ?? votes.length}
          loadingVotes={votesQuery.isLoading}
          votesFailed={votesQuery.isError}
          hasMoreVotes={!!votesQuery.hasNextPage}
          loadingMoreVotes={votesQuery.isFetchingNextPage}
          onMoreVotes={() => void votesQuery.fetchNextPage()}
          speeches={speeches ?? []}
          loadingSpeeches={loadingSpeeches}
        />
      )}

      {activeTab === "motioner" && <MemberMotions politicianId={id} />}

      {/* ── Debatter ────────────────────────────────────────────────── */}
      {activeTab === "debatter" && (
        <>
          <QuestionsSection
            path={`/politicians/${encodeURIComponent(id)}/questions`}
            queryKey={["politician-questions", id]}
            heading="Frågor och interpellationer"
            empty="Inga frågor eller interpellationer registrerade."
            showRole
          />
          <DebatterPanel speeches={speeches ?? []} loading={loadingSpeeches} />
        </>
      )}

      {/* ── Ämnen ───────────────────────────────────────────────────── */}
      {activeTab === "amnen" && (
        <AmnenPanel speeches={speeches ?? []} loading={loadingSpeeches} />
      )}
    </div>
  );
}
