import { useState } from "react";
import { PartyBadge } from "@/shared/components";
import { PARTY_COLORS } from "@/shared/design";
import { SourceMarker } from "@/components/sources/SourceMarker";
import type { PartyCode } from "@/shared/types";

interface Manifesto {
  party: PartyCode;
  year: number;
  title: string;
  type: "valmanifest" | "valplattform" | "partiprogram" | "politiska-riktlinjer";
  url: string;
  pdfUrl?: string;
  status: "published" | "expected";
  summary?: string;
}

const PARTY_NAMES: Record<PartyCode, string> = {
  S:       "Socialdemokraterna",
  M:       "Moderaterna",
  SD:      "Sverigedemokraterna",
  C:       "Centerpartiet",
  V:       "Vänsterpartiet",
  KD:      "Kristdemokraterna",
  L:       "Liberalerna",
  MP:      "Miljöpartiet",
  unknown: "Okänt",
};

const TYPE_LABELS: Record<Manifesto["type"], string> = {
  valmanifest:          "Valmanifest",
  valplattform:         "Valplattform",
  partiprogram:         "Partiprogram",
  "politiska-riktlinjer": "Politiska riktlinjer",
};

const MANIFESTOS: Manifesto[] = [
  // === 2026 ===
  { party: "S",  year: 2026, title: "Plan för Sverige – Socialdemokraternas valplattform 2026", type: "valplattform", url: "https://www.socialdemokraterna.se/nyheter/nyheter/2026-02-05-plan-for-sverige---socialdemokraterna-presenterar-valplattform-2026", status: "published", summary: "Fokus på ekonomi, välfärd, trygghet och jobb." },
  { party: "M",  year: 2026, title: "Moderaternas valmanifest 2026",           type: "valmanifest", url: "https://moderaterna.se/valmanifest", status: "expected", summary: "Förväntas presenteras inför valet september 2026." },
  { party: "SD", year: 2026, title: "Sverigedemokraternas valmanifest 2026",   type: "valmanifest", url: "https://sd.se",                     status: "expected", summary: "Förväntas presenteras inför valet september 2026." },
  { party: "C",  year: 2026, title: "Centerpartiets valmanifest 2026",         type: "valmanifest", url: "https://centerpartiet.se",          status: "expected", summary: "Förväntas presenteras inför valet september 2026." },
  { party: "V",  year: 2026, title: "Vänsterpartiets valplattform 2026",       type: "valplattform", url: "https://vansterpartiet.se",        status: "expected", summary: "Förväntas presenteras inför valet september 2026." },
  { party: "KD", year: 2026, title: "Kristdemokraternas valmanifest 2026",     type: "valmanifest", url: "https://kristdemokraterna.se",      status: "expected", summary: "Förväntas presenteras inför valet september 2026." },
  { party: "L",  year: 2026, title: "Liberalernas valmanifest 2026",           type: "valmanifest", url: "https://liberalerna.se",            status: "expected", summary: "Förväntas presenteras inför valet september 2026." },
  { party: "MP", year: 2026, title: "Miljöpartiets valmanifest 2026",          type: "valmanifest", url: "https://mp.se",                     status: "expected", summary: "Förväntas presenteras inför valet september 2026." },
  // === 2022 ===
  { party: "S",  year: 2022, title: "Vårt Sverige kan bättre – Socialdemokraternas valmanifest 2022", type: "valmanifest", url: "https://www.socialdemokraterna.se/5.36cb33a81817fcdd3ec4a8d.html", pdfUrl: "https://snd.gu.se/sv/vivill/party/s/v/2022", status: "published", summary: "Gängkriminalitet, sjukvård, klimatomställning och ekonomi." },
  { party: "M",  year: 2022, title: "Så får vi ordning på Sverige – Moderaternas valmanifest 2022",   type: "valmanifest", url: "https://moderaterna.se/nyhet/moderaternasvalmanifest/", pdfUrl: "https://moderaterna.se/app/uploads/2022/08/moderaternas_valmanifest_2022_webversion.pdf", status: "published", summary: "250 reformer inom trygghet, ekonomi, energi, invandring, vård och skola." },
  { party: "SD", year: 2022, title: "Sverigedemokraternas valmanifest 2022",   type: "valmanifest", url: "https://sd.se/wp-content/uploads/2022/08/valmanifest.pdf", pdfUrl: "https://sd.se/wp-content/uploads/2022/08/valmanifest.pdf", status: "published", summary: "Invandring, trygghet, energi och välfärd." },
  { party: "C",  year: 2022, title: "Centerpartiets valmanifest 2022",         type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/c/v/2022", pdfUrl: "https://snd.gu.se/sv/vivill/party/c/v/2022", status: "published", summary: "Landsbygd, företagande, grön omställning och frihet." },
  { party: "V",  year: 2022, title: "Vänsterpartiets valplattform 2022",       type: "valplattform", url: "https://snd.gu.se/sv/vivill/party/v/v/2022", pdfUrl: "https://snd.gu.se/sv/vivill/party/v/v/2022", status: "published", summary: "Välfärd, jämlikhet, klimat och arbetsvillkor." },
  { party: "KD", year: 2022, title: "Kristdemokraternas valmanifest 2022",     type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/kd/v/2022", pdfUrl: "https://snd.gu.se/sv/vivill/party/kd/v/2022", status: "published", summary: "Familj, trygghet, vård och skattesänkningar." },
  { party: "L",  year: 2022, title: "Liberalernas valmanifest 2022",           type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/l/v/2022", pdfUrl: "https://snd.gu.se/sv/vivill/party/l/v/2022", status: "published", summary: "Skola, integration, frihet och rättsstat." },
  { party: "MP", year: 2022, title: "Alla ska med när Sverige ställer om – Miljöpartiets valmanifest 2022", type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/mp/v/2022", pdfUrl: "https://snd.gu.se/sv/vivill/party/mp/v/2022", status: "published", summary: "Klimat, biologisk mångfald, rättvisa och jämlikhet." },
  // === 2018 ===
  { party: "S",  year: 2018, title: "Det största trygghetsprogrammet i modern tid – Valmanifest 2018", type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/s/v/2018",  pdfUrl: "https://snd.gu.se/sv/vivill/party/s/v/2018",  status: "published", summary: "Trygghet, välfärd och klimat." },
  { party: "M",  year: 2018, title: "Moderaternas valmanifest 2018",           type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/m/v/2018",  pdfUrl: "https://snd.gu.se/sv/vivill/party/m/v/2018",  status: "published", summary: "Jobb, integration, trygghet och ekonomi." },
  { party: "SD", year: 2018, title: "Sverigedemokraternas valmanifest 2018",   type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/sd/v/2018", pdfUrl: "https://snd.gu.se/sv/vivill/party/sd/v/2018", status: "published" },
  { party: "C",  year: 2018, title: "Centerpartiets valmanifest 2018",         type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/c/v/2018",  pdfUrl: "https://snd.gu.se/sv/vivill/party/c/v/2018",  status: "published" },
  { party: "V",  year: 2018, title: "Vänsterpartiets valplattform 2018",       type: "valplattform", url: "https://snd.gu.se/sv/vivill/party/v/v/2018", pdfUrl: "https://snd.gu.se/sv/vivill/party/v/v/2018", status: "published" },
  { party: "KD", year: 2018, title: "Kristdemokraternas valmanifest 2018",     type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/kd/v/2018", pdfUrl: "https://snd.gu.se/sv/vivill/party/kd/v/2018", status: "published" },
  { party: "L",  year: 2018, title: "Liberalernas valmanifest 2018",           type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/l/v/2018",  pdfUrl: "https://snd.gu.se/sv/vivill/party/l/v/2018",  status: "published" },
  { party: "MP", year: 2018, title: "Miljöpartiets valmanifest 2018",          type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/mp/v/2018", pdfUrl: "https://snd.gu.se/sv/vivill/party/mp/v/2018", status: "published" },
  // === 2014 ===
  { party: "S",  year: 2014, title: "Socialdemokraternas valmanifest 2014",    type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/s/v/2014",  status: "published" },
  { party: "M",  year: 2014, title: "Moderaternas valmanifest 2014",           type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/m/v/2014",  status: "published" },
  { party: "SD", year: 2014, title: "Sverigedemokraternas valmanifest 2014",   type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/sd/v/2014", status: "published" },
  { party: "C",  year: 2014, title: "Centerpartiets valmanifest 2014",         type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/c/v/2014",  status: "published" },
  { party: "V",  year: 2014, title: "Vänsterpartiets valplattform 2014",       type: "valplattform", url: "https://snd.gu.se/sv/vivill/party/v/v/2014", status: "published" },
  { party: "KD", year: 2014, title: "Kristdemokraternas valmanifest 2014",     type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/kd/v/2014", status: "published" },
  { party: "L",  year: 2014, title: "Folkpartiets valmanifest 2014",           type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/l/v/2014",  status: "published" },
  { party: "MP", year: 2014, title: "Miljöpartiets valmanifest 2014",          type: "valmanifest", url: "https://snd.gu.se/sv/vivill/party/mp/v/2014", status: "published" },
];

const YEARS = [2026, 2022, 2018, 2014];

function ManifestoCard({ m }: { m: Manifesto }) {
  const pc = PARTY_COLORS[m.party];
  const isExpected = m.status === "expected";

  return (
    <div
      className="rounded-xl p-5 flex flex-col gap-3"
      style={{
        background: isExpected ? "var(--color-surface-low)" : "var(--color-surface-lowest)",
        border: isExpected ? "1px dashed var(--color-surface-high)" : "none",
        opacity: isExpected ? 0.75 : 1,
      }}
    >
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <PartyBadge party={m.party} size="lg" />
          <span className="text-sm font-semibold text-on-surface">{PARTY_NAMES[m.party]}</span>
        </div>
        <span
          className="text-[10px] font-extrabold px-2 py-0.5 rounded shrink-0"
          style={isExpected
            ? { background: "#f9fafb", color: "#9ca3af" }
            : { background: pc?.light ?? "#f0fdf4", color: pc?.bg ?? "#16a34a" }
          }
        >
          {isExpected ? "Förväntas" : "Publicerat"}
        </span>
      </div>

      <div>
        <p className="text-[11px] text-on-surface-variant mb-0.5">{TYPE_LABELS[m.type]} <SourceMarker sourceId="seed-party-goals" /></p>
        <p className="text-xs font-medium text-on-surface leading-snug">{m.title}</p>
        {m.summary && (
          <p className="text-xs text-on-surface-variant mt-1 leading-relaxed">{m.summary}</p>
        )}
      </div>

      {!isExpected && (
        <div className="flex gap-2 mt-auto">
          <a
            href={m.url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-[11px] font-semibold text-primary hover:underline"
          >
            Läs manifest →
          </a>
          {m.pdfUrl && m.pdfUrl !== m.url && (
            <a
              href={m.pdfUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-[11px] font-semibold text-on-surface-variant hover:underline"
            >
              PDF
            </a>
          )}
        </div>
      )}
    </div>
  );
}

export function ManifestosPage() {
  const [selectedYear, setSelectedYear] = useState(2026);

  const yearManifestos = MANIFESTOS.filter((m) => m.year === selectedYear);

  return (
    <div>
      <div className="mb-6">
        <h2 className="font-display text-2xl font-extrabold tracking-tight mb-1">
          Valmanifest
        </h2>
        <p className="text-sm text-on-surface-variant leading-relaxed">
          Partiernas officiella valmanifest och valplattformar från 2014 till 2026.
        </p>
      </div>

      {/* ── Year tabs ─────────────────────────────────────────────── */}
      <div className="flex gap-1 mb-6" style={{ borderBottom: "1px solid var(--color-surface-high)" }}>
        {YEARS.map((year) => {
          const active = selectedYear === year;
          return (
            <button
              key={year}
              onClick={() => setSelectedYear(year)}
              className="px-4 py-2.5 text-sm font-medium transition-colors cursor-pointer"
              style={{
                borderBottom: active ? "2px solid var(--color-on-surface)" : "2px solid transparent",
                color: active ? "var(--color-on-surface)" : "var(--color-on-surface-variant)",
                fontWeight: active ? 700 : 500,
                marginBottom: "-1px",
              }}
            >
              {year}
            </button>
          );
        })}
      </div>

      {/* ── Manifesto grid ────────────────────────────────────────── */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {yearManifestos.map((m) => (
          <ManifestoCard key={`${m.party}-${m.year}`} m={m} />
        ))}
      </div>
    </div>
  );
}
