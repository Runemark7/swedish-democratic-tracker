import type { BudgetDocumentRef } from "@/shared/types";

interface Props {
  documents: BudgetDocumentRef[];
  compact?: boolean;
}

export function DocumentLinks({ documents, compact }: Props) {
  if (!documents || documents.length === 0) return null;

  return (
    <div className={`flex ${compact ? "gap-3" : "gap-4"} flex-wrap`}>
      {documents.map((doc) => (
        <a
          key={doc.url}
          href={doc.url}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1 text-primary hover:underline no-underline transition-colors"
          style={{ fontSize: compact ? "10px" : "11px" }}
          title={doc.description}
        >
          <span className="font-semibold">{doc.title}</span>
          <svg
            width={compact ? 10 : 12}
            height={compact ? 10 : 12}
            viewBox="0 0 12 12"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M5 1H2a1 1 0 0 0-1 1v8a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1V7" />
            <path d="M7 1h4v4" />
            <path d="M11 1 5.5 6.5" />
          </svg>
        </a>
      ))}
    </div>
  );
}
