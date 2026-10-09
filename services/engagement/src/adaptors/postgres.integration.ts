import {test} from 'node:test';
import assert from 'node:assert/strict';
import {randomUUID} from 'node:crypto';
import {PostgresContextDatabase, PostgresAggregateCommandStore, PostgresAggregateQueries} from './postgres.js';
import {Pool} from 'pg';
import {restoreAccount, restoreReward} from './restore.js';
import {claim, finish} from './dispatch.js';
import type {AccountState} from '../contexts/loyalty/domain/account.js';
import type {Metadata} from '../foundation/application.js';
import {RedeemRewardCommandHandler} from '../contexts/loyalty/application/commands.js';
import type {RewardState} from '../contexts/loyalty/domain/reward.js';
import {CorruptState, Rejection} from '../foundation/domain.js';
import {spawn} from 'node:child_process';
import {once} from 'node:events';
import {fileURLToPath} from 'node:url';

test('real transactions atomically record one root, receipts and realtime; rejected writes roll back', async () => {
  const db = new PostgresContextDatabase('loyalty', process.env.LOYALTY_DATABASE_URL!);
  try {
    await db.verify();
    const commands = new PostgresAggregateCommandStore<AccountState>(db, 'account', restoreAccount);
    const queries = new PostgresAggregateQueries<AccountState>(db, 'account', restoreAccount);
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
  assert.ok(process.env.CAFE_DISPOSABLE_PROJECT?.startsWith('cafe-reference-test-'));
  const url = new URL(process.env.DATABASE_ADMIN_URL!);
  url.pathname = '/cafe_loyalty';
  const admin = new Pool({connectionString: url.toString()});
  const db = new PostgresContextDatabase('loyalty', process.env.LOYALTY_DATABASE_URL!);
  try {
    await db.verify();
    const commands = new PostgresAggregateCommandStore<RewardState>(db, 'reward', restoreReward);
    const queries = new PostgresAggregateQueries<RewardState>(db, 'reward', restoreReward);
    const id = randomUUID(), commandId = randomUUID(), sourceId = randomUUID();
    const state: RewardState = {id, grantId: randomUUID(), customerId: 'broken', benefit: 'coffee',
      status: 'issued', expiresAt: '2026-10-04T12:00:00Z'};
    await commands.execute({id: randomUUID(), target: id, name: 'test.fixture', correlation: id, input: {}},
      () => ({state, status: 'issued', changed: true}));
    await assert.rejects(queries.get(id), CorruptState);
    const command = {orderId: randomUUID()};
    const metadata: Metadata = {id: commandId, target: id, name: 'loyalty.RedeemReward', correlation: id,
      input: command, consumer: 'test.corrupt-reward', sourceId, sourceHash: 'fixture'};
    const handler = new RedeemRewardCommandHandler(commands, () => new Date('2026-09-27T12:00:00Z'));
    await assert.rejects(handler.execute(metadata, command), error => error instanceof Error && !(error instanceof Rejection));
    assert.equal((await db.pool.query('SELECT count(*) FROM cafe.command_receipts WHERE command_id=$1', [commandId])).rows[0].count, '0');
    assert.equal((await db.pool.query('SELECT count(*) FROM cafe.consumer_receipts WHERE event_id=$1', [sourceId])).rows[0].count, '0');
    const client = await admin.connect();
    try {
      await client.query('BEGIN');
      await client.query("SET LOCAL session_replication_role='replica'");
      await client.query("UPDATE cafe.aggregates SET state=$1 WHERE kind='reward' AND id=$2",
        [{...state, customerId: randomUUID()}, id]);
      await client.query('COMMIT');
    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    } finally { client.release(); }
    assert.equal((await handler.execute(metadata, command)).status, 'redeemed');
  } finally { await Promise.all([db.pool.end(), admin.end()]); }
});

test('losing an idle PostgreSQL connection does not terminate the service and the pool recovers', async () => {
  const db = new PostgresContextDatabase('loyalty', process.env.LOYALTY_DATABASE_URL!);
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
  const db = new PostgresContextDatabase('loyalty', process.env.LOYALTY_DATABASE_URL!);
  try {
    const commands = new PostgresAggregateCommandStore<AccountState>(db, 'account', restoreAccount);
    const queries = new PostgresAggregateQueries<AccountState>(db, 'account', restoreAccount);
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

test('reply bytes commit atomically and a fresh request recovers the stored command outcome', async () => {
  const {currentReply, claimReply} = await import('./replies.js');
  const db = new PostgresContextDatabase('loyalty', process.env.LOYALTY_DATABASE_URL!);
  const commands = new PostgresAggregateCommandStore<AccountState>(db, 'account', restoreAccount);
  const queries = new PostgresAggregateQueries<AccountState>(db, 'account', restoreAccount);
  const id = randomUUID(), metadata: Metadata = {id: randomUUID(), target: id, name: 'test.reply', correlation: id, expected: 0, input: {}};
  const state: AccountState = {id, collections: 0, stampBalance: 0, grantsEarned: 0};
  const change = () => ({state, status: 'active', changed: true});
  try {
    const broken = {id: randomUUID(), committed: false, encode: () => {throw new Error('Reply encoding failed');}};
    await assert.rejects(currentReply.run(broken, () => commands.execute(metadata, change)), /Reply encoding failed/);
    assert.equal(broken.committed, false);
    assert.equal(await queries.get(id), undefined);
    const first = {id: randomUUID(), committed: false, encode: (value: object) => Buffer.from(JSON.stringify(value))};
    const outcome = await currentReply.run(first, () => commands.execute(metadata, change));
    assert.ok(first.committed);
    const retry = {...first, id: randomUUID(), committed: false};
    assert.deepEqual(await currentReply.run(retry, () => commands.execute(metadata, () => assert.fail('Replayed decision'))), outcome);
    assert.ok(retry.committed);
    for (const intent of [first, retry]) {
      const {rows: [row]} = await db.pool.query<{body: Buffer}>('SELECT body FROM cafe.command_replies WHERE id=$1', [intent.id]);
      assert.deepEqual(row!.body, first.encode(outcome));
    }
    assert.equal((await db.pool.query('SELECT count(*) FROM cafe.command_receipts WHERE aggregate_id=$1', [id])).rows[0].count, '1');
    await assert.rejects(db.pool.query('UPDATE cafe.command_replies SET body=body WHERE id=$1', [first.id]));
    // Leave the committed lease untouched, as a crashed process would.
    const abandoned = await claimReply(db);
    assert.ok(abandoned);
    await db.pool.query("UPDATE cafe.command_reply_dispatches SET available_at=clock_timestamp()+interval '1 hour' WHERE event_id<>$1", [abandoned.id]);
    const deadline = Date.now()+11000;
    let recovered: Awaited<ReturnType<typeof claimReply>>;
    while (!(recovered = await claimReply(db))) {
      assert.ok(Date.now() < deadline, 'Reply lease did not recover within the caller deadline');
      await new Promise(resolve => setTimeout(resolve, 50));
    }
    assert.equal(recovered.id, abandoned.id);
    assert.equal(recovered.expired, false);
    assert.deepEqual(recovered.body, abandoned.body);

  } finally { await db.pool.end(); }
});
