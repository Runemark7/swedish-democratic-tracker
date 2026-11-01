import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, useLocation } from "react-router-dom";
import { regeringApi } from "./api";
import { PartyBadge } from "@/shared/components";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { riksdagApi } from "@/features/riksdag/api";
import { GovernmentDocuments } from "./GovernmentDocuments";
import { BudgetPage } from "@/features/budget/BudgetPage";
import { MyndigheterSection, RiksdagPage } from "@/features/riksdag/RiksdagPage";
import { useRiksdag } from "@/hooks/useDemocracy";
import { swedishDate } from "@/shared/dates";
import { REGERINGEN_SECTIONS, type RegeringenSection } from "@/shared/regeringen";

const SECTION_LINKS: { id: RegeringenSection; label: string }[] = [
  { id: REGERINGEN_SECTIONS.statsrad, label: "Statsråd" },
  { id: REGERINGEN_SECTIONS.budget, label: "Budget" },
  { id: REGERINGEN_SECTIONS.riksdag, label: "Riksdag" },
  { id: REGERINGEN_SECTIONS.myndigheter, label: "Myndigheter" },
];

const sectionStyle: React.CSSProperties = {
  scrollMarginTop: 120,
  marginTop: 56,
  paddingTop: 28,
  borderTop: "1px solid var(--color-border)",
};

function useScrollToSection() {
  const { hash } = useLocation();
  useEffect(() => {
    const id = hash.startsWith("#") ? decodeURIComponent(hash.slice(1)) : "";
    if (!id) return;
    document.getElementById(id)?.scrollIntoView();
  }, [hash]);
}

/**
 * The type comes from the government row. Övergångsregering is the one case
 * the seat chart already refuses to draw as styre mot opposition, so the
 * sentence here uses the same limit: the label is what the row is called,
 * and the numbers live under Riksdag as chamber seats.
 */
function governmentSentence(type: string, since: string | null): string {
  if (type === "Övergångsregering") {
    const opened = since ? `Övergångsregering sedan ${since}.` : "Övergångsregering.";
    return `${opened} Mandaten under Riksdag är kammarens sammansättning, inte ett utsett styre.`;
  }
  return since ? `${type} sedan ${since}.` : `${type}.`;
}

export function RegeringenPage() {
  useScrollToSection();
  const { data, isLoading, error } = useQuery({
    queryKey: ["ministers"],
    queryFn: regeringApi.listMinisters,
  });
  // Government documents live here rather than on the front page: they are
  // government-level material, and the utskott partition the rest of the site
  // is organised around does not fit them.
  const { data: documents } = useQuery({
    queryKey: ["government-documents"],
    queryFn: riksdagApi.getAgenda,
  });
  const { data: riksdag } = useRiksdag();

  // No early return on the ministers query. The document index below is a
  // separate source, and bailing out of the whole page when the minister list is
  // unavailable would take the documents down with it — the reader would see
  // nothing rather than the half we still hold.
  const groups = data ?? [];
  const ruling = riksdag?.ruling;
  const governmentType = ruling && ruling.type && ruling.type !== "–" ? ruling.type : null;
  const since = ruling?.validFrom ? swedishDate(ruling.validFrom.slice(0, 10)) : null;

  return (
    <div>
      <div className="mb-8">
        <h2 className="font-display text-2xl font-extrabold tracking-tight mb-1">
          Regeringen
          <SourceMarker sourceId="riksdagen" />
        </h2>
        {governmentType && (
          <p className="text-sm text-on-surface mb-2 leading-relaxed max-w-lg">
            {governmentSentence(governmentType, since)}
            <SourceMarker sourceId="valmyndigheten-riksdagsval" />
          </p>
        )}
        <p className="text-sm text-on-surface-variant leading-relaxed max-w-lg">
          Statsråden som sitter nu, hur statsbudgeten är fördelad, riksdagens översikt och myndigheterna.
        </p>
        <nav aria-label="Avsnitt på sidan" style={{ display: "flex", gap: 18, flexWrap: "wrap", marginTop: 16 }}>
          {SECTION_LINKS.map((item) => (
            <Link
              key={item.id}
              to={{ hash: `#${item.id}` }}
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: 11,
                letterSpacing: "0.08em",
                textTransform: "uppercase",
                color: "var(--color-fg)",
                textDecoration: "none",
                borderBottom: "1px solid var(--color-border)",
                paddingBottom: 2,
              }}
            >
              {item.label}
            </Link>
          ))}
        </nav>
      </div>

      <section id={REGERINGEN_SECTIONS.statsrad} aria-labelledby="statsrad-rubrik" style={{ scrollMarginTop: 120 }}>
        <h3 id="statsrad-rubrik" className="text-xs font-mono uppercase tracking-widest text-on-surface-variant mb-4">
          Statsråd
        </h3>

        {isLoading && (
          <p className="text-on-surface-variant text-sm py-4">Laddar statsråd...</p>
        )}
        {error && (
          <p className="text-contradiction text-sm py-4">
            Kunde inte ladda statsråden. Regeringens dokument nedan påverkas inte.
          </p>
        )}

        <div className="space-y-8">
          {groups.map((group) => (
            <div key={group.departmentCode}>
              <h3 className="text-xs font-mono uppercase tracking-widest text-on-surface-variant mb-3">
                {group.department}
              </h3>
              <div className="bg-surface-lowest rounded-xl overflow-hidden">
                {group.ministers?.map((m, i) => (
                  <Link
                    key={m.id}
                    to={`/regering/${m.id}`}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 16,
                      padding: "14px 20px",
                      borderTop: i > 0 ? "1px solid var(--color-surface-high)" : undefined,
                      textDecoration: "none",
                      color: "inherit",
                      transition: "background 0.15s",
                    }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = "var(--color-surface-high)"; }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = "transparent"; }}
                  >
                    <PartyBadge party={m.party ?? ""} size="sm" />
                    <div style={{ flex: 1 }}>
                      <div style={{ fontSize: 14, fontWeight: 600 }}>{m.name}</div>
                      <div style={{ fontSize: 12, color: "var(--color-fg-muted)" }}>{m.title}</div>
                    </div>
                    <span style={{ fontSize: 12, color: "var(--color-fg-muted)" }}>→</span>
                  </Link>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div className="mt-10">
          <h3 className="text-xs font-mono uppercase tracking-widest text-on-surface-variant mb-3">
            Regeringens dokument
          </h3>
          {/* The selection rule is stated because the list is ours. A reader can
              argue with which documents we registered; they cannot argue with a
              paraphrase we never wrote. */}
          <p className="text-sm text-on-surface-variant leading-relaxed max-w-xl mb-4">
            Regeringens program- och budgetdokument för mandatperioden, som vi har
            registrerat dem. Vi länkar dokumenten i sin helhet och sammanfattar dem
            inte. Saknas ett dokument betyder det att vi inte har registrerat det.
          </p>
          <div className="bg-surface-lowest rounded-xl px-5 py-2">
            <GovernmentDocuments items={documents ?? []} />
          </div>
        </div>
      </section>

      <section id={REGERINGEN_SECTIONS.budget} aria-labelledby="budget-rubrik" style={sectionStyle}>
        <BudgetPage embedded />
      </section>

      <section id={REGERINGEN_SECTIONS.riksdag} aria-labelledby="riksdag-rubrik" style={sectionStyle}>
        <h2 id="riksdag-rubrik" className="text-xs font-mono uppercase tracking-widest text-on-surface-variant mb-3">
          Riksdag
        </h2>
        <RiksdagPage embedded />
      </section>

      <section id={REGERINGEN_SECTIONS.myndigheter} aria-labelledby="myndigheter-rubrik" style={sectionStyle}>
        <h2 id="myndigheter-rubrik" className="font-display text-2xl font-extrabold tracking-tight mb-1">
          Myndigheter
        </h2>
        <p className="text-sm text-on-surface-variant leading-relaxed max-w-lg mb-6">
          De tio myndigheterna med högst driftkostnad. Varje namn leder till myndighetens sida.
          Registret längst ned listar de övriga.
        </p>
        <MyndigheterSection />
      </section>
    </div>
  );
}
