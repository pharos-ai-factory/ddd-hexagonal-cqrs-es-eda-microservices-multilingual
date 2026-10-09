import {test} from 'node:test';
import assert from 'node:assert/strict';
import {subscriptionBarrier} from './queries';
import {scan} from './httpProjection';
import {merge, type Snapshot} from './model';

test('paginated scans load every row beyond 100 and preserve newer live revisions', async t => {
  const urls: string[] = [];
  let state: Snapshot = {};
  const rows = Array.from({length: 205}, (_, index) => ({kind: 'drink' as const, version: 1,
    state: {id: String(index), name: 'old', revision: 1, published: false}}));
  t.mock.method(globalThis, 'fetch', async (url: string) => {
    urls.push(url);
    const cursor = new URL(url, 'http://test').searchParams.get('cursor');
    const offset = Number(cursor ?? 0);
    if (offset === 100) {
      state = merge(state, {...rows[0]!, version: 2, state: {...rows[0]!.state, name: 'live'}});
      // A root inserted behind the current cursor arrives through the subscription.
      state = merge(state, {...rows[0]!, state: {...rows[0]!.state, id: '-1', name: 'new'}});
    }
    return new Response(JSON.stringify({items: rows.slice(offset, offset+100), nextCursor: offset < 200 ? String(offset+100) : null}));
  });
  const loaded = await scan('listDrinks', new AbortController().signal);
  loaded.forEach(row => { state = merge(state, {...row, kind: 'drink'}); });
  assert.equal(loaded.length, 205);
  assert.equal(urls.length, 3);
  assert.match(urls[1]!, /cursor=100/);
  assert.equal(state['drink/0']!.state.id, '0');
  assert.equal(state['drink/0']!.version, 2);
  assert.equal(Object.keys(state).length, 206);
  assert.equal(state['drink/-1']!.state.id, '-1');
});

test('a failed continuation fails the scan and repeated cursors cannot loop', async t => {
  let calls = 0;
  t.mock.method(globalThis, 'fetch', async () => {
    if (++calls === 2) throw new Error('unavailable');
    return new Response(JSON.stringify({items: [], nextCursor: 'next'}));
  });
  await assert.rejects(scan('listDrinks', new AbortController().signal), /unavailable/);
  t.mock.method(globalThis, 'fetch', async () => new Response(JSON.stringify({items: [], nextCursor: 'same'})));
  await assert.rejects(scan('listDrinks', new AbortController().signal), /did not advance/);
  t.mock.method(globalThis, 'fetch', async () => new Response(JSON.stringify({items: [], nextCursor: null})));
  assert.deepEqual(await scan('listDrinks', new AbortController().signal), []);
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
