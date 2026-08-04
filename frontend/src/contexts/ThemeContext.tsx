import { useEffect, useState } from "react";
import { ThemeContext, type Theme } from "./theme";

// The context, its type and useTheme live in ./theme so that this file exports
// only a component — Fast Refresh cannot preserve state across edits to a file
// that also exports non-components.

function getInitialTheme(): Theme {
  const stored = localStorage.getItem("sdt-theme");
  if (stored === "dark" || stored === "light") return stored;
  return window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<Theme>(getInitialTheme);

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", theme === "light" ? "light" : "");
    localStorage.setItem("sdt-theme", theme);
  }, [theme]);

  function toggleTheme() {
    setTheme((t) => (t === "dark" ? "light" : "dark"));
  }

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}
