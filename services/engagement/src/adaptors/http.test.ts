import {test} from 'node:test';
import assert from 'node:assert/strict';
import {once} from 'node:events';
import {readFileSync} from 'node:fs';
import type {AddressInfo} from 'node:net';
import {server} from './http.js';

const key = 'test-operational-key-'.repeat(2);

test('owner HTTP exposes health and authenticated diagnostics, with no business routes', async t => {
  let diagnosticCalls = 0;
  const http = server(key, async () => { diagnosticCalls++; return {databases: {}}; });
  http.listen(0, '127.0.0.1');
  await once(http, 'listening');
  t.after(() => new Promise<void>((resolve, reject) => http.close(error => error ? reject(error) : resolve())));
  const base = `http://127.0.0.1:${(http.address() as AddressInfo).port}`;
  const headers = {Authorization: 'Bearer '+key};
  assert.equal((await fetch(base+'/healthz')).status, 200);
  assert.equal((await fetch(base+'/diagnostics')).status, 401);
  const diagnostics = await fetch(base+'/diagnostics', {headers});
  assert.equal(diagnostics.status, 200);
  assert.equal(diagnostics.headers.get('cache-control'), 'no-store');
  assert.deepEqual(await diagnostics.json(), {databases: {}});
  const document = JSON.parse(readFileSync(new URL('../../../../contracts/services/api/http_api/api.openapi.json', import.meta.url), 'utf8')) as {paths: Record<string, unknown>};
  const paths = Object.keys(document.paths).filter(path => path.startsWith('/api/v1/'));
  assert.ok(paths.length > 0);
  for (const path of paths) {
    const url = base+path.replace(/^\/api/, '').replace('{id}', '11111111-1111-4111-8111-111111111111');
    for (const method of ['GET', 'POST']) {
      assert.equal((await fetch(url, {method, headers})).status, 404, method+' '+path);
    }
  }
  assert.equal((await fetch(base+'/healthz', {method: 'POST', headers})).status, 404);
  assert.equal((await fetch(base+'/diagnostics', {method: 'POST', headers})).status, 404);
  assert.equal(diagnosticCalls, 1);
});

test('diagnostic infrastructure errors produce a sanitised retryable response', async t => {
  const http = server(key, async () => { throw new Error('database-password'); });
  http.listen(0, '127.0.0.1');
  await once(http, 'listening');
  t.after(() => new Promise<void>((resolve, reject) => http.close(error => error ? reject(error) : resolve())));
  const response = await fetch(`http://127.0.0.1:${(http.address() as AddressInfo).port}/diagnostics`, {
    headers: {Authorization: 'Bearer '+key},
  });
  assert.equal(response.status, 503);
  assert.deepEqual(await response.json(), {code: 'temporarily_unavailable'});
});
