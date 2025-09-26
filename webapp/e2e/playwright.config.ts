import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  fullyParallel: false,
  retries: 1,
  workers: 1,
  // ← 証跡を成功時だけ出すレポーターを追加（下でコード例）
  reporter: [["line"], ["./tests/_stampReporter.ts"]],
  use: {
    baseURL: process.env.BASE_URL ?? "http://tuning-nginx",
    ignoreHTTPSErrors: true,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  expect: {
    timeout: 120_000,
  },
  timeout: 600_000,
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
});
