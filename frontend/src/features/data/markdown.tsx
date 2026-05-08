import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

interface MarkdownProps {
  source: string;
}

export function Markdown({ source }: MarkdownProps) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        h1: ({ children }) => (
          <h1 style={{ fontFamily: "var(--font-serif)", fontSize: 28, margin: "24px 0 12px", color: "var(--color-fg)" }}>{children}</h1>
        ),
        h2: ({ children }) => (
          <h2 style={{ fontFamily: "var(--font-serif)", fontSize: 22, margin: "24px 0 8px", color: "var(--color-fg)" }}>{children}</h2>
        ),
        h3: ({ children }) => (
          <h3 style={{ fontFamily: "var(--font-mono)", fontSize: 11, letterSpacing: "0.15em", textTransform: "uppercase", color: "var(--color-fg-muted)", margin: "20px 0 8px" }}>{children}</h3>
        ),
        p: ({ children }) => (
          <p style={{ fontSize: 14, lineHeight: 1.7, color: "var(--color-fg)", margin: "0 0 12px" }}>{children}</p>
        ),
        a: ({ href, children }) => {
          const isExternal = !!href && /^https?:/.test(href);
          return (
            <a
              href={href}
              target={isExternal ? "_blank" : undefined}
              rel={isExternal ? "noopener noreferrer" : undefined}
              style={{ color: "var(--color-accent)", textDecoration: "none", borderBottom: "1px solid var(--color-border)" }}
            >
              {children}
            </a>
          );
        },
        code: ({ children, className }) => {
          const isInline = !className;
          if (isInline) {
            return (
              <code style={{ fontFamily: "var(--font-mono)", fontSize: 12, background: "var(--color-track)", padding: "1px 5px", borderRadius: 2 }}>{children}</code>
            );
          }
          return <code className={className}>{children}</code>;
        },
        pre: ({ children }) => (
          <pre style={{ fontFamily: "var(--font-mono)", fontSize: 12, background: "var(--color-sdt-surface)", border: "1px solid var(--color-border)", padding: 12, overflow: "auto", lineHeight: 1.5 }}>{children}</pre>
        ),
        ul: ({ children }) => (
          <ul style={{ paddingLeft: 20, margin: "0 0 12px", fontSize: 14, lineHeight: 1.7, color: "var(--color-fg)" }}>{children}</ul>
        ),
        table: ({ children }) => (
          <div style={{ overflowX: "auto", margin: "12px 0" }}>
            <table style={{ borderCollapse: "collapse", fontSize: 13 }}>{children}</table>
          </div>
        ),
        th: ({ children }) => (
          <th style={{ borderBottom: "1px solid var(--color-border)", padding: "6px 12px", textAlign: "left", fontFamily: "var(--font-mono)", fontSize: 10, textTransform: "uppercase", letterSpacing: "0.1em" }}>{children}</th>
        ),
        td: ({ children }) => (
          <td style={{ borderBottom: "1px solid var(--color-border)", padding: "6px 12px" }}>{children}</td>
        ),
      }}
    >
      {source}
    </ReactMarkdown>
  );
}
