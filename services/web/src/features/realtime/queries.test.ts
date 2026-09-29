import {test} from 'node:test';
import assert from 'node:assert/strict';
import {query, queryAll, subscriptionBarrier} from './queries';
import {merge, type Snapshot} from './model';

test('paginated scans load every row beyond 100 and preserve newer live revisions', async () => {
  const urls: string[] = [];
  let state: Snapshot = {};
  const rows = Array.from({length: 205}, (_, index) => ({kind: 'drink' as const, version: 1,
    state: {id: String(index), name: 'old', revision: 1, published: false}}));
  const loaded = await queryAll('/drinks', async url => {
    urls.push(url);
    const cursor = new URL(url, 'http://test').searchParams.get('cursor');
    const offset = Number(cursor ?? 0);
    if (offset === 100) {
      state = merge(state, {...rows[0]!, version: 2, state: {...rows[0]!.state, name: 'live'}});
      // A root inserted behind the current cursor arrives through the subscription.
      state = merge(state, {...rows[0]!, state: {...rows[0]!.state, id: '-1', name: 'new'}});
    }
    return {items: rows.slice(offset, offset+100), nextCursor: offset < 200 ? String(offset+100) : null};
  });
  loaded.forEach(row => { state = merge(state, row); });
  assert.equal(loaded.length, 205);
  assert.equal(urls.length, 3);
  assert.match(urls[1]!, /cursor=100/);
  assert.equal(state['drink/0']!.state.id, '0');
  assert.equal(state['drink/0']!.version, 2);
  assert.equal(Object.keys(state).length, 206);
  assert.equal(state['drink/-1']!.state.id, '-1');
});

test('a failed continuation fails the scan and repeated cursors cannot loop', async () => {
  let calls = 0;
  await assert.rejects(queryAll('/rows', async () => {
    if (++calls === 2) throw new Error('unavailable');
    return {items: [1], nextCursor: 'next'};
  }), /unavailable/);
  await assert.rejects(queryAll('/rows', async () => ({items: [], nextCursor: 'same'})), /did not advance/);
  const empty: number[] = await queryAll('/rows', async () => ({items: [], nextCursor: null}));
  assert.deepEqual(empty, []);
});

test('ordinary query responses need no pagination fields', async t => {
  const urls: string[] = [];
  t.mock.method(globalThis, 'fetch', async (url: string) => {
    urls.push(url);
    return new Response(JSON.stringify({total: 205}));
  });
  assert.deepEqual(await query<{total: number}>('/summary'), {total: 205});
  assert.deepEqual(urls, ['/summary']);
});

test('all subscriptions must attach before initial or gap scans; recovered history needs none', () => {
  const barrier = subscriptionBarrier(['a', 'b']);
  barrier.reset(false);
  assert.equal(barrier.subscribed('a', false), false);
  assert.equal(barrier.subscribed('b', false), true);
  assert.equal(barrier.subscribed('b', false), false);
  barrier.reset(false);
  assert.equal(barrier.subscribed('a', true), false);
  assert.equal(barrier.subscribed('b', true), false);
  barrier.reset(false);
  assert.equal(barrier.subscribed('a', false), false);
  assert.equal(barrier.subscribed('b', true), true);
  barrier.reset(true);
  assert.equal(barrier.subscribed('a', true), false);
  assert.equal(barrier.subscribed('b', true), true);
});
