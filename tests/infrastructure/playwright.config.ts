import {defineConfig} from '@playwright/test';

export default defineConfig({
  testDir: '.',
  testMatch: 'session-revocation.spec.ts',
  workers: 1,
  timeout: 30_000,
  expect: {timeout: 5_000},
  reporter: 'list',
  outputDir: '../../test-results/session-revocation',
  use: {headless: true, trace: 'retain-on-failure'},
});
