import { useState } from "react";
import { Link } from "react-router-dom";
import { SpecificityBadge, TopicTag, ConsequencePanel } from "@/shared/components";
import { SourceMarker } from "@/components/sources/SourceMarker";
import type { components } from "@/shared/api-contract";

type Goal = components["schemas"]["Goal"];

/**
 * The one renderer for a party goal, used by the party record view and the
 * committee page alike. Three surfaces used to render a goal and two of them
 * stated the same derived number in two different wordings; keeping a single
 * component makes that drift impossible by construction rather than policed.
 *
 * The prop is `Goal`, deliberately not `GoalWithAlignment`. `relevantVotes` is
 * not merely unrendered here — it is not in scope, so no future edit can put a
 * count back beside a goal without first widening this type and being asked why.
 *
 * No denominator beside a goal is honest. The keyword count under-claims,
 * because the keyword list is ours rather than the party's; a committee's full
 * count over-claims, because a committee deciding forty things does not mean one
 * promise was tested forty times. So the card states what the party said, cites
 * where it said it, and leaves the judgement to the reader.
 */
export function GoalCard({ goal, party }: { goal: Goal; party: string }) {
  const [open, setOpen] = useState(false);
  const keywords = goal.keywords ?? [];

  return (
    <div
      className="rounded-xl overflow-hidden transition-all"
      style={{
        background: open ? "var(--color-surface-low)" : "var(--color-surface-lowest)",
        borderLeft: "3px solid var(--color-surface-high)",
      }}
    >
      <button
        onClick={() => setOpen(!open)}
        className="w-full text-left px-5 pt-4 pb-2 cursor-pointer"
        aria-expanded={open}
      >
        <div className="flex justify-between items-start gap-4">
          <div className="flex-1 min-w-0">
            <div className="flex gap-2 items-center flex-wrap mb-2">
              <SpecificityBadge specificity={goal.specificity} />
              <TopicTag topic={goal.topic} />
            </div>
            <p className="text-sm font-semibold text-on-surface leading-snug">
              &ldquo;{goal.goalText}&rdquo;
            </p>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            <span
              className="text-lg text-on-surface-variant transition-transform duration-300"
              style={{ transform: open ? "rotate(180deg)" : "rotate(0deg)" }}
            >
              ▾
            </span>
          </div>
        </div>
      </button>

      {/* The source line sits outside the toggle button, not inside it.
          SourceMarker renders its own <button>, and nesting one button in
          another is invalid HTML — React reports a hydration error and the
          marker's clicks bubble into the card toggle, so the source popover
          fights the expander. The card this replaced had exactly that defect;
          since every surface now renders this one component, fixing it here
          fixes it everywhere. */}
      <div className="px-5 pb-4 flex items-center gap-3 text-xs text-on-surface-variant flex-wrap">
        <span>
          {goal.sourceDocument} <SourceMarker sourceId="seed-party-goals" />
        </span>
      </div>

      {open && (
        <div className="px-5 pb-5">
          <div className="h-px bg-surface-high mb-4" />
          {keywords.length > 0 && (
            <div className="mb-4">
              <Link
                to={`/parties/${party}/goals/${goal.id}/votes`}
                className="inline-block text-xs text-primary font-semibold hover:underline"
              >
                Omröstningar vars titel nämner {keywords.join(", ")} →
              </Link>
              {/* The keywords are ours. Saying so is the point: a title search is
                  a way in to the record, not a verdict on the promise. The card
                  this replaced stated an untested-goal verdict beside a Riksdagen
                  source marker, attributing our keyword gap to their record. */}
              <p className="text-[11px] text-on-surface-variant mt-1.5 leading-relaxed">
                Nyckelorden är våra, inte partiets. En träff betyder att titeln
                nämner ordet — inte att omröstningen prövade målet, och en
                utebliven träff betyder inte att målet är oprövat.{" "}
                <SourceMarker sourceId="seed-party-goals" />
              </p>
            </div>
          )}
          <ConsequencePanel topic={goal.topic} />
        </div>
      )}
    </div>
  );
}
