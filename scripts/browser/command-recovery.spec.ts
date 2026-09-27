import {test, expect} from '@playwright/test';
import {readFileSync} from 'node:fs';
import {randomUUID} from 'node:crypto';

const env = Object.fromEntries(readFileSync(process.env.CAFE_ENV_FILE ?? '.local/dev.env', 'utf8')
  .trim().split('\n').map(line => line.split('=')));

test('an uncertain command survives lost authentication and recovers after signing in again', async ({page, context}) => {
  await page.goto(`http://127.0.0.1:${env.WEB_PORT}`);
  const login = async () => {
    await page.getByLabel('Local operator access code').fill(env.OPERATOR_PASSWORD!);
    await page.getByRole('button', {name: 'Open the café'}).click();
    await expect(page.getByTestId('connection')).toHaveText('Live updates connected');
  };
  await login();
  const name = 'Session recovery '+randomUUID().slice(0, 8);
  await page.getByLabel('Drink name', {exact: true}).fill(name);
  const attempts: string[] = [];
  await page.route('**/api/v1/menu/drinks/*', async route => {
    if (route.request().method() !== 'POST') return route.continue();
    attempts.push(route.request().headers()['idempotency-key']!);
    if (attempts.length !== 1) return route.continue();
    expect((await route.fetch()).status()).toBe(200);
    await route.abort('failed');
  });
  const retry = page.getByRole('button', {name: 'Retry the same command'});
  await page.getByRole('button', {name: 'Create drink', exact: true}).click();
  await expect(retry).toBeVisible();
  const original = await page.evaluate(() => sessionStorage.getItem('cafe:pending-command'));
  expect(original).not.toBeNull();
  await context.clearCookies();
  const feedbackBeforeRetry = await page.getByRole('status').allTextContents();
  const rejected = page.waitForResponse(response => response.request().method() === 'POST' && response.status() === 401);
  await retry.click();
  await rejected;
  await expect.poll(() => page.getByRole('status').allTextContents()).not.toEqual(feedbackBeforeRetry);
  expect(await page.evaluate(() => sessionStorage.getItem('cafe:pending-command'))).toBe(original);
  await expect(page.getByLabel('Local operator access code')).toBeVisible();
  await login();
  const recovered = page.waitForResponse(response => response.request().method() === 'POST' &&
    response.url().includes('/api/v1/menu/drinks/') && response.status() === 200 &&
    response.request().headers()['idempotency-key'] === attempts[0]);
  await retry.click();
  await recovered;
  await expect(retry).toHaveCount(0);
  expect(attempts).toHaveLength(3);
  expect(new Set(attempts).size).toBe(1);
  await expect.poll(() => page.evaluate(() => sessionStorage.getItem('cafe:pending-command'))).toBeNull();
  await expect(page.locator('.list-row').filter({hasText: name})).toHaveCount(1);
});
