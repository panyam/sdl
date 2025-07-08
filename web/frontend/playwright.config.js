import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 30000,
  expect: {
    timeout: 5000
  },
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:8080',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure'
  },
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
      },
    }
  ],
  // webServer: {
  //   command: '../sdl console --port 8080',
  //   port: 8080,
  //   reuseExistingServer: !process.env.CI,
  //   timeout: 10000
  // }
});