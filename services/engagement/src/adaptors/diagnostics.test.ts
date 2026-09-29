import {test} from 'node:test';
import assert from 'node:assert/strict';
import {once} from 'node:events';
import {failure, processWorkers} from './diagnostics.js';
import {server} from './http.js';

test('failure before claiming work survives without exposing credentials', () => {
  failure('test', 'connect', new Error('amqp://username:secret@host'));
  failure('test', 'consumer', new Error('secret payload'), 'event-id', 'correlation-id', true);
  const state = processWorkers();
  assert.equal(state['test/connect']?.failures, 1);
  assert.equal(state['test/consumer']?.deadLetterTransfers, 1);
  assert.ok(!JSON.stringify(state).includes('secret'));
});
test('diagnostics requires service authentication and failure keeps liveness healthy', async () => {
  const key = 'x'.repeat(32);
  const http = server(key, [], async () => { throw new Error('postgres://secret@host'); });
  http.listen(0, '127.0.0.1');
  await once(http, 'listening');
  const address = http.address();
  assert.ok(address && typeof address === 'object');
  const url = `http://127.0.0.1:${address.port}`;
  try {
    assert.equal((await fetch(url+'/diagnostics')).status, 401);
    const response = await fetch(url+'/diagnostics', {headers: {authorization: 'Bearer '+key}});
    assert.equal(response.status, 503);
    assert.ok(!(await response.text()).includes('secret'));
    assert.equal((await fetch(url+'/healthz')).status, 200);
  } finally { http.closeAllConnections(); await new Promise<void>(resolve => http.close(() => resolve())); }
});
