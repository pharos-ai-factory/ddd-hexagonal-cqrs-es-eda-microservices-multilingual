import {test} from 'node:test';
import assert from 'node:assert/strict';
import {pageRequest, pageResponse} from './pagination.js';
const id = '00000000-0000-4000-8000-000000000001';
test('pagination is opt-in and cursors are bound to the queried resource', () => {
  assert.equal(pageRequest(new URLSearchParams(), '/accounts'), undefined);
  const page = pageResponse({items: [], nextId: id}, '/accounts');
  assert.deepEqual(pageRequest(new URLSearchParams({limit: '100', cursor: page.nextCursor!}), '/accounts'), {limit: 100, after: id});
  assert.equal(pageResponse({items: []}, '/accounts').nextCursor, null);
  assert.throws(() => pageRequest(new URLSearchParams({limit: '1', cursor: page.nextCursor!}), '/rewards'));
});
test('malformed, repeated or unbounded page parameters are rejected', () => {
  for (const query of ['limit=0', 'limit=101', 'limit=-1', 'limit=01', 'limit=1.2', 'limit=1&limit=2',
    'cursor=abc', 'limit=1&cursor=', 'limit=1&cursor=***', 'limit=1&cursor=a&cursor=b']) {
    assert.throws(() => pageRequest(new URLSearchParams(query), '/accounts'), query);
  }
});
