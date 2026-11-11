import { defineConfig, devices } from "@playwright/test";

// BASE_URL picks the dev server's port too, so parallel checkouts can each run their own.
const baseURL = process.env.BASE_URL ?? "http://127.0.0.1:5173";
const port = new URL(baseURL).port || "5173";

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  workers: 1,
  reporter: process.env.CI
    ? "github"
    : [["list"], ["html", { open: "never" }]],
  use: {
    baseURL,
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  webServer: {
    // cwd is the tests/ directory (where this config lives).
    command: `npm run dev --prefix ../frontend -- --host 127.0.0.1 --port ${port}`,
    url: `http://127.0.0.1:${port}`,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
