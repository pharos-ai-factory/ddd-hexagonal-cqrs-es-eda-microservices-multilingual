import {test} from 'node:test';
import assert from 'node:assert/strict';
import {randomUUID} from 'node:crypto';
import {Database, Commands, Queries} from './postgres.js';
import {claim, finish} from './dispatch.js';
import type {AccountState} from '../contexts/loyalty/domain/account.js';
import type {Metadata} from '../foundation/application.js';
import {RedeemReward} from '../contexts/loyalty/application/commands.js';
import type {RewardState} from '../contexts/loyalty/domain/reward.js';
import {Rejection} from '../foundation/domain.js';
import {spawn} from 'node:child_process';
import {once} from 'node:events';
import {fileURLToPath} from 'node:url';

test('real transactions atomically record one root, receipts and realtime; rejected writes roll back', async () => {
  const db = new Database('loyalty', process.env.LOYALTY_DATABASE_URL!);
  try {
    await db.verify();
    const commands = new Commands<AccountState>(db, 'account'), queries = new Queries<AccountState>(db, 'account');
    const id = randomUUID();
    const metadata: Metadata = {id: randomUUID(), target: id, name: 'test.open', correlation: randomUUID(), expected: 0, input: {}};
    const state: AccountState = {id, collections: 0, stampBalance: 0, grantsEarned: 0};
    await assert.rejects(commands.execute(metadata, () => ({state, status: 'active', changed: true,
      publications: [{name: 'invalid', payload: {}}]})));
    assert.equal(await queries.get(id), undefined);
    const first = await commands.execute(metadata, () => ({state, status: 'active', changed: true}));
    assert.deepEqual(await commands.execute(metadata, () => assert.fail('Replayed decision')), first);
    assert.equal((await commands.execute({...metadata, input: {different: true}}, () => assert.fail('Conflicting decision')))
      .rejection?.code, 'idempotency_conflict');
    const {rows: [counts]} = await db.pool.query(`SELECT
      (SELECT count(*) FROM cafe.command_receipts WHERE aggregate_id=$1) AS receipts,
      (SELECT count(*) FROM cafe.realtime_publications WHERE aggregate_id=$1) AS publications`, [id]);
    assert.deepEqual(counts, {receipts: '1', publications: '1'});
    const outcomes = await Promise.all([1, 2].map(() => commands.execute({...metadata, id: randomUUID(), expected: 1},
      () => ({state: {...state, collections: 1, stampBalance: 1}, status: 'active', changed: true}))));
    assert.equal(outcomes.filter(item => !item.rejection).length, 1);
    assert.equal((await queries.get(id))?.version, 2);
    const row = await claim(db, true);
    assert.ok(row);
    await db.pool.query("UPDATE cafe.realtime_dispatches SET lease_until=clock_timestamp()-interval '1 second' WHERE event_id=$1", [row.id]);
    const reclaimed = await claim(db, true);
    assert.ok(reclaimed); assert.equal(reclaimed.id, row.id);
    await finish(db, row, true);
    assert.equal((await db.pool.query('SELECT completed_at FROM cafe.realtime_dispatches WHERE event_id=$1', [row.id])).rows[0].completed_at, null);
    await finish(db, reclaimed, true);
    assert.ok((await db.pool.query('SELECT completed_at FROM cafe.realtime_dispatches WHERE event_id=$1', [row.id])).rows[0].completed_at);
  } finally { await db.pool.end(); }
});

test('corrupt rewards do not become durable business rejections', async () => {
  const db = new Database('loyalty', process.env.LOYALTY_DATABASE_URL!);
  try {
    await db.verify();
    const commands = new Commands<RewardState>(db, 'reward');
    const id = randomUUID(), commandId = randomUUID(), sourceId = randomUUID();
    const state: RewardState = {id, grantId: randomUUID(), customerId: 'broken', benefit: 'coffee',
      status: 'issued', expiresAt: '2026-10-04T12:00:00Z'};
    await commands.execute({id: randomUUID(), target: id, name: 'test.fixture', correlation: id, input: {}},
      () => ({state, status: 'issued', changed: true}));
    const command = {orderId: randomUUID()};
    const metadata: Metadata = {id: commandId, target: id, name: 'loyalty.RedeemReward', correlation: id,
      input: command, consumer: 'test.corrupt-reward', sourceId, sourceHash: 'fixture'};
    const handler = new RedeemReward(commands, () => new Date('2026-09-27T12:00:00Z'));
    await assert.rejects(handler.execute(metadata, command), error => error instanceof Error && !(error instanceof Rejection));
    assert.equal((await db.pool.query('SELECT count(*) FROM cafe.command_receipts WHERE command_id=$1', [commandId])).rows[0].count, '0');
    assert.equal((await db.pool.query('SELECT count(*) FROM cafe.consumer_receipts WHERE event_id=$1', [sourceId])).rows[0].count, '0');
    await commands.execute({id: randomUUID(), target: id, name: 'test.repair', correlation: id, input: {}, expected: 1},
      () => ({state: {...state, customerId: randomUUID()}, status: 'issued', changed: true}));
    assert.equal((await handler.execute(metadata, command)).status, 'redeemed');
  } finally { await db.pool.end(); }
});

test('losing an idle PostgreSQL connection does not terminate the service and the pool recovers', async () => {
  const db = new Database('loyalty', process.env.LOYALTY_DATABASE_URL!);
  const child = spawn(process.execPath, ['--import', 'tsx',
    fileURLToPath(new URL('../../tests/pool-failure.mts', import.meta.url))],
  {stdio: ['ignore', 'ignore', 'ignore', 'ipc']});
  const exited = once(child, 'exit');
  try {
    const ready = await Promise.race([
      once(child, 'message').then(([message]) => message as {pid: number}),
      exited.then(() => { throw new Error('Pool probe exited before establishing its connection'); }),
    ]);
    await db.pool.query('SELECT pg_terminate_backend($1)', [ready.pid]);
    child.send('recover');
    const [code] = await exited;
    assert.equal(code, 0, 'An idle connection failure must be handled and permit a subsequent query');
  } finally {
    if (child.exitCode === null) child.kill();
    await db.pool.end();
  }
});

test('queries traverse more than one page and retain unpaginated reads', async () => {
  const db = new Database('loyalty', process.env.LOYALTY_DATABASE_URL!);
  try {
    const commands = new Commands<AccountState>(db, 'account'), queries = new Queries<AccountState>(db, 'account');
    for (let index = 0; index < 105; index++) {
      const id = randomUUID(), state: AccountState = {id, collections: 0, stampBalance: 0, grantsEarned: 0};
      await commands.execute({id: randomUUID(), target: id, name: 'test.pagination', correlation: id, expected: 0, input: {}},
        () => ({state, status: 'active', changed: true}));
    }
    const all = await queries.list();
    assert.ok(all.length >= 105);
    const combined = [];
    let after: string | undefined;
    do {
      const page = await queries.page({limit: 37, after});
      assert.ok(page.items.length <= 37);
      combined.push(...page.items);
      if (page.nextId) assert.ok(page.nextId > (after ?? ''));
      after = page.nextId;
    } while (after);
    assert.deepEqual(combined, all);
  } finally { await db.pool.end(); }
});
