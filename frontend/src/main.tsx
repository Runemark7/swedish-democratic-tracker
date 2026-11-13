import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import "./index.css";
import App from "./App";
import { retryDelayMs, shouldRetry } from "./shared/api-errors";
import { ThemeProvider } from "./contexts/ThemeContext";
import { MyPartyProvider } from "./contexts/MyPartyContext";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      // Temporary failures (429, 5xx, network) get two more attempts; a 404 none.
      retry: shouldRetry,
      retryDelay: retryDelayMs,
    },
  },
});

const root = document.getElementById("root");
if (!root) throw new Error("Root element not found");

createRoot(root).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <ThemeProvider>
          <MyPartyProvider>
            <App />
          </MyPartyProvider>
        </ThemeProvider>
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>,
);
