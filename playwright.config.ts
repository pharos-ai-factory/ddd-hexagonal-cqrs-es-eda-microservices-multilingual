import {defineConfig} from '@playwright/test';
export default defineConfig({
  testDir: './scripts/browser',
  fullyParallel: false,
  workers: 1,
  timeout: 120_000,
  expect: {timeout: 20_000},
  reporter: 'list',
  use: {headless: true, actionTimeout: 20_000, navigationTimeout: 20_000,
    viewport: {width: 1440, height: 1080}, trace: 'retain-on-failure'},
});
