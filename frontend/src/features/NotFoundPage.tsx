import { Link, useLocation } from "react-router-dom";

const LINKS: { to: string; label: string }[] = [
  { to: "/", label: "Startsidan" },
  { to: "/beslut", label: "Beslut" },
  { to: "/motioner", label: "Motioner" },
  { to: "/politicians", label: "Ledamöter" },
  { to: "/parties", label: "Partier" },
];

/** Catch-all for URLs that match no route. */
export function NotFoundPage() {
  const { pathname } = useLocation();
  return (
    <div className="sdt-page" data-testid="not-found" style={{ padding: "24px 32px 64px", maxWidth: 760 }}>
      <h1 className="text-2xl font-bold">Sidan finns inte</h1>
      <p className="text-on-surface-variant" style={{ marginTop: 12 }}>
        Vi hittar ingen sida på adressen <code>{pathname}</code>. Den kan ha flyttats, eller så är adressen felstavad.
      </p>
      <p style={{ marginTop: 16 }}>Gå vidare till:</p>
      <ul style={{ marginTop: 8, paddingLeft: 20, listStyle: "disc" }}>
        {LINKS.map((l) => (
          <li key={l.to}>
            <Link to={l.to} style={{ textDecoration: "underline" }}>{l.label}</Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
