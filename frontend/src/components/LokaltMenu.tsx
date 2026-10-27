import { useEffect, useId, useRef, useState } from "react";
import { NavLink } from "react-router-dom";
import { LOKALT_NAV, isActive, isLokaltSection } from "@/shared/nav";

/**
 * Side-product entry. The word goes to the landing; the menu reaches
 * kommuner and regioner without leaving the header.
 */
export function LokaltMenu({ pathname }: { pathname: string }) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const menuId = useId();
  const active = isLokaltSection(pathname);

  useEffect(() => {
    if (!open) return;
    function onPointer(event: MouseEvent) {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false);
    }
    function onKey(event: KeyboardEvent) {
      if (event.key === "Escape") setOpen(false);
    }
    document.addEventListener("mousedown", onPointer);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onPointer);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  const pill: React.CSSProperties = {
    display: "inline-flex",
    alignItems: "center",
    padding: "6px 14px",
    borderRadius: 999,
    border: active ? "none" : "1px solid var(--color-accent)",
    background: active ? "var(--color-fg)" : "transparent",
    color: active ? "var(--color-bg)" : "var(--color-accent)",
    fontFamily: "var(--font-mono)",
    fontSize: 11,
    letterSpacing: "1px",
    textTransform: "uppercase",
    textDecoration: "none",
    whiteSpace: "nowrap",
  };

  return (
    <div ref={rootRef} style={{ position: "relative" }}>
      <div style={{ display: "inline-flex", alignItems: "stretch" }}>
        <NavLink to="/lokalt" style={pill} aria-current={active ? "page" : undefined}>
          Lokalt
        </NavLink>
        <button
          type="button"
          aria-expanded={open}
          aria-haspopup="menu"
          aria-controls={menuId}
          aria-label="Visa kommuner och regioner"
          onClick={() => setOpen((value) => !value)}
          style={{
            ...pill,
            borderRadius: 999,
            marginLeft: 4,
            padding: "6px 10px",
            cursor: "pointer",
            background: open ? "var(--color-track)" : pill.background,
          }}
        >
          ▾
        </button>
      </div>
      {open && (
        <div
          id={menuId}
          role="menu"
          style={{
            position: "absolute",
            top: "calc(100% + 6px)",
            left: 0,
            zIndex: 30,
            minWidth: 180,
            background: "var(--color-sdt-surface)",
            color: "var(--color-fg)",
            border: "1px solid var(--color-border)",
            borderRadius: 8,
            padding: 6,
            boxShadow: "0 12px 32px color-mix(in srgb, var(--color-fg) 12%, transparent)",
          }}
        >
          {LOKALT_NAV.map((item) => {
            const itemActive = isActive(item, pathname);
            return (
              <NavLink
                key={item.to}
                to={item.to}
                role="menuitem"
                onClick={() => setOpen(false)}
                style={{
                  display: "block",
                  padding: "8px 10px",
                  borderRadius: 6,
                  textDecoration: "none",
                  fontFamily: "var(--font-body)",
                  fontSize: 14,
                  fontWeight: itemActive ? 700 : 500,
                  color: itemActive ? "var(--color-fg)" : "var(--color-fg-muted)",
                  background: itemActive ? "var(--color-track)" : "transparent",
                }}
              >
                {item.label}
              </NavLink>
            );
          })}
        </div>
      )}
    </div>
  );
}
