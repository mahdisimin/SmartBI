import { defineConfig } from '@playwright/test'

// Backend (:8091) must already be running against the local SMARTBI DB.
export default defineConfig({
  testDir: './tests',
  timeout: 30_000,
  workers: 1,
  reporter: [['list']],
  use: { baseURL: 'http://localhost:5173', channel: 'chrome', headless: true, trace: 'retain-on-failure' },
  webServer: {
    command: 'npx vite --port 5173 --strictPort',
    cwd: '../../dashboard-monitor',
    url: 'http://localhost:5173',
    reuseExistingServer: true,
    timeout: 60_000,
  },
})
