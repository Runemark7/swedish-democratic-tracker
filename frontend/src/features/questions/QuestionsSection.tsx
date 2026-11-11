import { useQuery } from "@tanstack/react-query";
import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";
import { PartyBadge } from "@/shared/components";
import { SourceMarker } from "@/components/sources/SourceMarker";
import { swedishDate } from "@/shared/dates";

type QuestionItem = components["schemas"]["QuestionItem"];

const KIND: Record<string, string> = { fraga: "Skriftlig fråga", interpellation: "Interpellation" };
const RULE: Record<string, string> = {
  "riksdagen-reference": "Riksdagen pekar ut motionen",
  "cited-beteckning": "Frågan anger motionens beteckning",
  "redaktor-confirmed": "Bekräftad av redaktören: frågan nämner motionen i text",
};

interface Props {
  /** Path under /api that returns { questions }. */
  path: string;
  queryKey: unknown[];
  heading: string;
  empty: string;
  /** Show the member's role (asked or answered) on each row. */
  showRole?: boolean;
  /** Show the rule behind a motion link on each row. */
  showRule?: boolean;
}

export function QuestionsSection({ path, queryKey, heading, empty, showRole, showRule }: Props) {
  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: () => api.get<{ questions: QuestionItem[] }>(path),
  });
  const items = data?.questions ?? [];
  return (
    <section className="mb-8" aria-label={heading} data-testid="questions-section">
      <h2 id="questions-heading" className="text-lg font-semibold mb-2">
        {heading}
        <SourceMarker sourceId="questions" />
      </h2>
      {isLoading ? null : items.length === 0 ? (
        <p className="text-sm text-on-surface-variant">{empty}</p>
      ) : (
        <ul className="flex flex-col gap-3">
          {items.map((q) => (
            <li key={q.dokId} className="text-sm" data-testid="question-row">
              <a href={q.sourceUrl} className="text-link hover:underline font-medium">
                {KIND[q.kind]} {q.riksmote}:{q.beteckning} – {q.title}
              </a>
              <div className="text-on-surface-variant">
                {q.asker.name} {q.asker.party ? <PartyBadge party={q.asker.party} /> : null} till {q.addressed.name}
                {q.answerDate ? `, besvarad ${swedishDate(q.answerDate)}` : ", ej besvarad"}
                {q.answerer.name && q.answerer.name !== q.addressed.name ? ` av ${q.answerer.name}` : ""}
              </div>
              {showRole && q.role ? (
                <div className="text-xs text-on-surface-variant">{q.role === "asker" ? "Ställd av ledamoten" : "Besvarad av statsrådet"}</div>
              ) : null}
              {showRule && q.rule ? <div className="text-xs text-on-surface-variant">Koppling: {RULE[q.rule] ?? q.rule}</div> : null}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
