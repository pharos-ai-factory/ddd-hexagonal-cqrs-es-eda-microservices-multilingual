import {test} from 'node:test';
import assert from 'node:assert/strict';
import {request, commandRequest} from './client';

test('typed headers, query values and command bodies reach the HTTP boundary', async () => {
  const original = globalThis.fetch;
  const sent: Array<{url: string; init: RequestInit}> = [];
  globalThis.fetch = async (url, init) => {
    sent.push({url: String(url), init: init ?? {}});
    return new Response('{}');
  };
  try {
    await request('listDrinks', {query: {limit: 2, cursor: 'a/b+c'}});
    await commandRequest('createDrink', {name: 'Tea'}, {path: {id: 'drink-id'}, headers: {
      'Idempotency-Key': 'command-id', 'If-Match': '0', 'X-Correlation-ID': 'correlation-id',
    }});
    assert.equal(sent[0]?.url, '/api/v1/menu/drinks?limit=2&cursor=a%2Fb%2Bc');
    assert.equal(sent[1]?.url, '/api/v1/menu/drinks/drink-id');
    assert.equal(sent[1]?.init.method, 'POST');
    const headers = new Headers(sent[1]?.init.headers);
    assert.equal(headers.get('Idempotency-Key'), 'command-id');
    assert.equal(headers.get('If-Match'), '0');
    assert.equal(headers.get('X-Correlation-ID'), 'correlation-id');
    assert.equal(headers.get('Content-Type'), 'application/json');
    assert.deepEqual(JSON.parse(String(sent[1]?.init.body)), {name: 'Tea'});
  } finally { globalThis.fetch = original; }
});
