import {test, expect, type Page} from '@playwright/test';
import {readFileSync} from 'node:fs';
import {randomUUID} from 'node:crypto';
import {Pool} from 'pg';

const env = Object.fromEntries(readFileSync(process.env.CAFE_ENV_FILE ?? '.local/dev.env', 'utf8')
  .trim().split('\n').map(line => line.split('=')));
const base = `http://127.0.0.1:${env.WEB_PORT}`;
const live = (page: Page) => expect(page.getByTestId('connection')).toHaveText('Live updates connected');
async function admin(method: string, body: object) {
  const response = await fetch(`http://127.0.0.1:${env.REALTIME_ADMIN_PORT}/api/${method}`, {
    method: 'POST', headers: {'Content-Type': 'application/json', 'X-API-Key': env.CENTRIFUGO_API_KEY!},
    body: JSON.stringify(body),
  });
  expect(response.status).toBe(200);
  const result = await response.json() as {error?: unknown; result?: {offset?: number}};
  expect(result.error).toBeUndefined();
  return result.result;
}
async function createDrink(name: string) {
  const response = await fetch(`http://127.0.0.1:${env.API_PORT}/api/v1/menu/drinks/${randomUUID()}`, {
    method: 'POST', headers: {'Content-Type': 'application/json', Authorization: 'Bearer '+env.API_KEY,
      'Idempotency-Key': randomUUID(), 'If-Match': '0', 'X-Correlation-ID': randomUUID()},
    body: JSON.stringify({name}),
  });
  expect(response.status).toBe(200);
}

async function projected(owner: 'menu' | 'ordering', name: string, key: string) {
  // Browser publication and RabbitMQ consumption are independent deliveries.
  // Observe committed prerequisites from the test host, using runtime read access.
  const pool = new Pool({host: '127.0.0.1', port: Number(env.PG_PORT), database: 'cafe_'+owner,
    user: 'cafe_'+owner, password: env[owner.toUpperCase()+'_DB_PASSWORD'],
    max: 1, connectionTimeoutMillis: 5_000, query_timeout: 5_000});
  try {
    await expect.poll(async () => {
      const result = await pool.query<{arrived: boolean}>(
        'SELECT EXISTS(SELECT 1 FROM cafe.projections WHERE name=$1 AND key=$2) AS arrived', [name, key]);
      return result.rows[0]?.arrived;
    }, {message: `${owner} must receive ${name} before its next command`}).toBe(true);
  } finally { await pool.end(); }
}

test('six contexts, independent windows, binary recovery, durable logout and uncertain commands', async ({page, context}) => {
  const errors: string[] = [];
  page.on('pageerror', error => errors.push(error.message));
  const businessGets: string[] = [];
  context.on('request', request => {
    if (request.method() === 'GET' && request.url().includes('/api/v1/')) businessGets.push(request.url());
  });
  await page.goto(base);
  await page.getByLabel('Local operator access code').fill(env.OPERATOR_PASSWORD!);
  await page.getByRole('button', {name: 'Open the café'}).click();
  await live(page);
  await page.getByRole('button', {name: 'New customer'}).click();
  const customer = await page.getByLabel('Customer identity').inputValue();
  const second = await context.newPage();
  second.on('pageerror', error => errors.push(error.message));
  await second.goto(base);
  await live(second);
  await expect(second.getByLabel('Customer identity')).toHaveValue(customer);

  const drink = 'Browser coffee '+randomUUID().slice(0, 6);
  await page.getByLabel('Drink name', {exact: true}).fill(drink);
  // Lose the response after the server commits, then retry the same persisted identity.
  let commandKey = '', replayKey = '';
  await page.route('**/api/v1/menu/drinks/*', async route => {
    if (route.request().method() !== 'POST') return route.continue();
    if (!commandKey) {
      commandKey = route.request().headers()['idempotency-key']!;
      await route.fetch();
      await route.abort('failed');
    } else {
      replayKey = route.request().headers()['idempotency-key']!;
      await route.continue();
    }
  });
  await page.getByRole('button', {name: 'Create drink', exact: true}).click();
  await page.getByRole('button', {name: 'Retry the same command'}).click();
  await expect.poll(() => replayKey).toBe(commandKey);
  await expect(second.locator('.list-row').filter({hasText: drink})).toHaveCount(1);
  await page.locator('.list-row').filter({hasText: drink}).getByRole('button', {name: 'Publish drink'}).click();
  await expect(page.locator('.list-row').filter({hasText: drink})).toContainText('Published');
  await page.getByRole('button', {name: 'New edition', exact: true}).click();
  await expect(page.getByLabel('Published drink')).toBeVisible();
  const [drinkId] = await page.getByLabel('Published drink').selectOption({label: drink});
  await projected('menu', 'published-drinks', drinkId+'/1');
  await page.getByRole('button', {name: 'Add offer', exact: true}).click();
  await expect(page.getByRole('button', {name: 'Publish menu', exact: true})).toBeEnabled();
  await page.getByRole('button', {name: 'Publish menu', exact: true}).click();
  await expect(page.getByText('This edition is published.', {exact: false})).toBeVisible();
  await projected('ordering', 'published-menus', await page.getByLabel('Menu edition').inputValue());
  const readsAfterInitial = businessGets.length;
  expect(readsAfterInitial).toBe(16); // Eight owner queries in each independent window.
  for (let index = 0; index < 3; index++) {
    await page.getByRole('button', {name: 'Start an order', exact: true}).click();
    const order = page.getByTestId('order').filter({hasText: 'draft'});
    await expect(order).toHaveCount(1);
    await order.getByRole('button', {name: /^\+ /}).first().click();
    await expect(order.getByRole('button', {name: 'Place order'})).toBeEnabled();
    const title = await order.locator('strong').first().innerText();
    await order.getByRole('button', {name: 'Place order'}).click();
    const ticket = second.getByTestId('ticket').filter({hasText: title});
    await ticket.getByRole('button', {name: 'Start preparation'}).click();
    await ticket.getByRole('button', {name: 'Mark drinks ready'}).click();
    const pickup = page.getByTestId('pickup').filter({hasText: title});
    await expect(pickup).toBeVisible();
    await pickup.getByRole('textbox').fill('WRONG1');
    await pickup.getByRole('button', {name: 'Collect order'}).click();
    await expect(page.getByRole('status')).toContainText('does not match');
    await pickup.getByRole('textbox').fill(await pickup.getByTestId('collection-code').innerText());
    await pickup.getByRole('button', {name: 'Collect order'}).click();
    await expect(pickup).toHaveCount(0);
    await expect(second.getByTestId('loyalty-account')).toContainText(`${index+1} orders collected`);
  }
  await expect(page.getByTestId('reward')).toHaveCount(1);
  await expect(second.getByTestId('notification').filter({hasText: 'sent'})).toHaveCount(4);
  expect(businessGets.length).toBe(readsAfterInitial);
  await page.screenshot({path: '.local/cafe-desktop.png', fullPage: true});

  await context.setOffline(true);
  await admin('disconnect', {user: '00000000-0000-4000-8000-000000000001',
    disconnect: {code: 3001, reason: 'Browser recovery exercise'}});
  await expect(page.getByTestId('connection')).toHaveText('Reconnecting…');
  const beforeRecovery = (await admin('history', {channel: 'cafe:menu', limit: 0}))!.offset!;
  const recoveredName = 'Recovered '+randomUUID().slice(0, 8);
  await createDrink(recoveredName);
  await expect.poll(async () => (await admin('history', {channel: 'cafe:menu', limit: 0}))!.offset!).toBeGreaterThan(beforeRecovery);
  await context.setOffline(false);
  await expect(page.locator('.list-row').filter({hasText: recoveredName})).toBeVisible();
  await expect(second.locator('.list-row').filter({hasText: recoveredName})).toBeVisible();
  expect(businessGets.length).toBe(readsAfterInitial);

  await context.setOffline(true);
  await admin('disconnect', {user: '00000000-0000-4000-8000-000000000001',
    disconnect: {code: 3001, reason: 'Browser history gap exercise'}});
  await expect(page.getByTestId('connection')).toHaveText('Reconnecting…');
  const beforeGap = (await admin('history', {channel: 'cafe:menu', limit: 0}))!.offset!;
  const missedName = 'Missed '+randomUUID().slice(0, 8);
  await createDrink(missedName);
  await expect.poll(async () => (await admin('history', {channel: 'cafe:menu', limit: 0}))!.offset!).toBeGreaterThan(beforeGap);
  await admin('history_remove', {channel: 'cafe:menu'});
  const reconciledName = 'Reconciled '+randomUUID().slice(0, 8);
  await createDrink(reconciledName);
  await context.setOffline(false);
  await expect(page.locator('.list-row').filter({hasText: reconciledName})).toBeVisible();
  await expect(second.locator('.list-row').filter({hasText: reconciledName})).toBeVisible();
  await expect(page.locator('.list-row').filter({hasText: missedName})).toBeVisible();
  await expect.poll(() => businessGets.length).toBe(readsAfterInitial+16);
  await page.setViewportSize({width: 390, height: 844});
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await page.screenshot({path: '.local/cafe-mobile.png', fullPage: true});
  await page.getByRole('button', {name: 'Sign out', exact: true}).click();
  await expect(second.getByLabel('Local operator access code')).toBeVisible();
  expect(errors).toEqual([]);
});
