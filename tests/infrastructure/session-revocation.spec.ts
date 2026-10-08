import {test, expect, type Page} from '@playwright/test';
import {readFileSync} from 'node:fs';

const env = Object.fromEntries(readFileSync(process.env.CAFE_ENV_FILE!, 'utf8')
  .trim().split('\n').map(line => line.split('=')));
const host = env.CAFE_TEST_HOST ?? '127.0.0.1';
const api = `http://${host}:${env.API_PORT}`;
type Reply = {connect?: {client: string}; push?: {pub?: unknown}; error?: {code: number}};
declare global {
  interface Window {
    revocationProbe: {socket: WebSocket; replies: Reply[]; closed: boolean};
  }
}

async function connect(page: Page) {
  await page.evaluate(({host, port}) => {
    const socket = new WebSocket(`ws://${host}:${port}/connection/websocket`);
    const state = {socket, replies: [] as Reply[], closed: false};
    window.revocationProbe = state;
    socket.onopen = () => socket.send(JSON.stringify({id: 1, connect: {}}));
    socket.onmessage = event => state.replies.push(JSON.parse(String(event.data)) as Reply);
    socket.onclose = () => { state.closed = true; };
  }, {host, port: env.REALTIME_PORT!});
}

test('logout revokes a connection whose successful authentication response is still in flight',
  async ({page, context}) => {
    const login = await context.request.post(api+'/auth/login', {
      headers: {Origin: api}, data: {password: env.OPERATOR_PASSWORD},
    });
    expect(login.status()).toBe(200);
    await page.goto(api+'/auth/session');
    const publish = async () => {
      const response = await context.request.post(`http://${host}:${env.REALTIME_ADMIN_PORT}/api/publish`, {
        headers: {'X-API-Key': env.CENTRIFUGO_API_KEY!, 'X-Centrifugo-Error-Mode': 'transport'},
        data: {channel: 'cafe:menu', b64data: Buffer.from(JSON.stringify({marker: 'revocation regression'})).toString('base64')},
      });
      expect(response.status()).toBe(200);
      expect((await response.json() as {error?: unknown}).error).toBeUndefined();
    };
    // A healthy session must really subscribe and receive data from this broker.
    await connect(page);
    await expect.poll(() => page.evaluate(() =>
      window.revocationProbe.replies.some(reply => Boolean(reply.connect)))).toBe(true);
    await expect.poll(async () => {
      const response = await context.request.get(api+'/test/state');
      return (await response.json() as {refreshes: number}).refreshes;
    }).toBeGreaterThan(0);
    await publish();
    await expect.poll(() => page.evaluate(() =>
      window.revocationProbe.replies.some(reply => Boolean(reply.push?.pub)))).toBe(true);
    await page.evaluate(() => window.revocationProbe.socket.close());
    await expect.poll(() => page.evaluate(() => window.revocationProbe.closed)).toBe(true);
    expect((await context.request.post(api+'/test/hold')).status()).toBe(204);
    await connect(page);
    const evidence = async () => {
      const response = await context.request.get(api+'/test/state');
      expect(response.status()).toBe(200);
      return await response.json() as {
        captured: boolean; released: boolean; cancelled: boolean; disconnects: number; pending: number;
      };
    };
    try {
      await expect.poll(async () => (await evidence()).captured).toBe(true);
      expect((await context.request.post(api+'/auth/logout', {headers: {Origin: api}})).status()).toBe(200);
      expect((await context.request.get(api+'/auth/session')).status()).toBe(401);
      // The actual Go worker must complete its confirmed disconnect before release.
      await expect.poll(async () => {
        const state = await evidence();
        return state.disconnects > 0 && state.pending === 0;
      }).toBe(true);
      expect((await context.request.post(api+'/test/release')).status()).toBe(204);
      await expect.poll(async () => {
        const state = await evidence();
        return state.released || state.cancelled;
      }).toBe(true);
      expect((await evidence()).cancelled, 'A proxy timeout must not masquerade as successful revocation').toBe(false);
      await expect.poll(() => page.evaluate(() => {
        const probe = window.revocationProbe;
        return probe.closed || probe.replies.some(reply => Boolean(reply.connect || reply.error));
      })).toBe(true);
      const errors = await page.evaluate(() => window.revocationProbe.replies.flatMap(reply => reply.error ? [reply.error.code] : []));
      if (errors.length) {
        // An expired connect grant is a protocol rejection; the bare WebSocket
        // can remain open without an authenticated client or subscriptions.
        expect(errors).toEqual([110]); // Centrifugo's expired connection grant.
        expect(await page.evaluate(() => window.revocationProbe.replies.some(reply => Boolean(reply.connect)))).toBe(false);
      }
      await publish();
      await expect.poll(() => page.evaluate(() => {
        const probe = window.revocationProbe;
        return probe.closed || probe.replies.some(reply => Boolean(reply.error || reply.push?.pub));
      })).toBe(true);
      expect(await page.evaluate(() => window.revocationProbe.replies
        .filter(reply => Boolean(reply.push?.pub))),
      'A session revoked before connection registration must not receive publications').toEqual([]);
      if (!errors.length) await expect.poll(() => page.evaluate(() => window.revocationProbe.closed)).toBe(true);
    } finally {
      await context.request.post(api+'/test/release');
      await page.evaluate(() => window.revocationProbe.socket.close());
    }
  });
