import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {randomUUID, createHash} from 'node:crypto';
import {setTimeout as pause} from 'node:timers/promises';

export type State = Record<string, any>;
export type Loaded = {version: number; state: State};
export type Reply = {status: number; body: State};
export function configuration(): Record<string, string> {
  const path = process.env.CAFE_ENV_FILE;
  assert.ok(path, 'CAFE_ENV_FILE must select the disposable integration project');
  const values = Object.fromEntries(readFileSync(path, 'utf8').split('\n').filter(line => line && !line.startsWith('#'))
    .map(line => { const separator = line.indexOf('='); return [line.slice(0, separator), line.slice(separator + 1)]; }));
  assert.ok(values.COMPOSE_PROJECT_NAME?.startsWith('cafe-reference-test-'), 'Gherkin recovery scenarios require a disposable test project');
  return values;
}

// This is the published technical identity convention, not imported business code.
export function derived(purpose: string, key: string): string {
  const bytes = createHash('sha256').update('cafe-reference/v1\0'+purpose+'\0'+key).digest().subarray(0, 16);
  bytes[6] = (bytes[6]! & 15) | 128; bytes[8] = (bytes[8]! & 63) | 128;
  const hex = bytes.toString('hex');
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

export async function eventually<T>(description: string, observe: () => Promise<T | undefined>, timeout = 30_000): Promise<T> {
  const deadline = Date.now() + timeout;
  do {
    const result = await observe();
    if (result !== undefined) return result;
    await pause(100);
  } while (Date.now() < deadline);
  throw new Error('Timed out waiting for '+description);
}

export class Client {
  correlation = randomUUID();
  constructor(readonly values: Record<string, string>) {}
  async request(path: string, body?: object, version = 0, key = randomUUID()): Promise<Reply> {
    const response = await fetch(`http://127.0.0.1:${this.values.API_PORT}/api/v1/${path}`, {
      method: body === undefined ? 'GET' : 'POST',
      headers: {Authorization: 'Bearer '+this.values.API_KEY, 'Content-Type': 'application/json',
        'X-Correlation-ID': this.correlation, 'Idempotency-Key': key, 'If-Match': String(version)},
      ...(body === undefined ? {} : {body: JSON.stringify(body)}), signal: AbortSignal.timeout(10_000),
    });
    return {status: response.status, body: await response.json() as State};
  }
  async command(path: string, body: object, version: number) {
    const reply = await this.request(path, body, version);
    assert.equal(reply.status, 200, `${path}: ${JSON.stringify(reply.body)}`);
    return reply.body;
  }
  async availableCommand(path: string, body: object, version: number) {
    return eventually('projection for '+path, async () => {
      const reply = await this.request(path, body, version);
      if (reply.status === 200) return reply.body;
      assert.equal(reply.status, 422, JSON.stringify(reply));
      assert.ok(['menu_pending', 'drink_revision_pending'].includes(reply.body.rejection?.code), JSON.stringify(reply));
      // A rejection remains recorded. Each observation here makes a new attempt.
      return undefined;
    });
  }
  async get(path: string): Promise<Loaded> {
    const reply = await this.request(path);
    assert.equal(reply.status, 200, `${path}: ${JSON.stringify(reply.body)}`);
    return {version: Number(reply.body.version), state: reply.body.state};
  }
  async state(path: string, matches: (state: State) => boolean): Promise<Loaded> {
    return eventually(path, async () => {
      const reply = await this.request(path);
      assert.ok([200, 404].includes(reply.status), JSON.stringify(reply));
      return reply.status === 200 && matches(reply.body.state)
        ? {version: Number(reply.body.version), state: reply.body.state} : undefined;
    });
  }
  async provider(path: string, body?: object): Promise<any> {
    const response = await fetch(`http://127.0.0.1:${this.values.DELIVERY_PORT}${path}`, {
      method: body === undefined ? 'GET' : 'POST',
      headers: {Authorization: 'Bearer '+this.values.DELIVERY_KEY, 'Content-Type': 'application/json'},
      ...(body === undefined ? {} : {body: JSON.stringify(body)}), signal: AbortSignal.timeout(5_000),
    });
    assert.equal(response.status, 200);
    return response.json();
  }
}
