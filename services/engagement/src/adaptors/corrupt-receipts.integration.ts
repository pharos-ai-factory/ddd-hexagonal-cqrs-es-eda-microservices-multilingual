import {test} from 'node:test';
import assert from 'node:assert/strict';
import {randomUUID} from 'node:crypto';
import {Pool} from 'pg';
import {PostgresContextDatabase, PostgresAggregateCommandStore, PostgresAggregateQueries} from './postgres.js';
import {restoreAccount} from '../contexts/loyalty/adaptors/persistence/accounts.js';
import type {AccountState} from '../contexts/loyalty/domain/loyalty-account.js';
import type {Metadata, Outcome} from '../foundation/application.js';
import {CorruptState, Rejection} from '../foundation/domain.js';

for (const receipt of ['command', 'consumer'] as const) {
  for (const shape of ['empty', 'null', 'incomplete_rejection'] as const) {
    test(`corrupt ${receipt} receipt (${shape}) requires repair before an identical retry`, async () => {
      assert.ok(process.env.CAFE_DISPOSABLE_PROJECT?.startsWith('cafe-reference-test-'),
        'Receipt corruption requires the disposable integration project');
      const url = new URL(process.env.DATABASE_ADMIN_URL!);
      url.pathname = '/cafe_loyalty';
      const admin = new Pool({connectionString: url.toString()});
      const db = new PostgresContextDatabase('loyalty', process.env.LOYALTY_DATABASE_URL!);
      try {
        await db.verify();
        const commands = new PostgresAggregateCommandStore<AccountState>(db, 'account', restoreAccount);
        const queries = new PostgresAggregateQueries<AccountState>(db, 'account', restoreAccount);
        const id = randomUUID();
        const state: AccountState = {id, collections: 0, grantsEarned: 0, stampBalance: 0};
        const m: Metadata = {id: randomUUID(), target: id, name: 'test.open', correlation: randomUUID(),
          expected: 0, input: {}, consumer: 'loyalty.test-receipt-'+id, sourceId: randomUUID(), sourceHash: 'original-wire'};
        const initial = {...m};
        if (receipt === 'command') delete initial.consumer;
        const first = await commands.execute(initial,
          () => ({state, status: 'active', changed: true}));
        assert.equal(first.rejection, undefined);
        const original = await queries.get(id);
        const broken: unknown = shape === 'empty' ? {} : shape === 'null' ? null :
          {aggregateId: id, version: 1, status: '', rejection: {code: 'not_found'}};
        const table = receipt === 'command' ? 'command_receipts' : 'consumer_receipts';
        const key = receipt === 'command' ? 'command_id' : 'event_id';
        const identity = receipt === 'command' ? m.id : m.sourceId;
        const update = `UPDATE cafe.${table} SET outcome=$1::jsonb WHERE ${key}=$2`;
        await admin.query(update, [JSON.stringify(broken), identity]);
        let decisions = 0;
        const retry = (): Promise<Outcome> => commands.execute(m, () => {
          decisions++;
          throw new Error('A saved receipt must not run the decision again');
        });
        await assert.rejects(retry(), error => error instanceof Error && !(error instanceof Rejection),
          'Corrupt receipt authority must fail transiently instead of returning an outcome');
        assert.equal(decisions, 0);
        assert.deepEqual(await queries.get(id), original);
        const {rows: [counts]} = await admin.query(`SELECT
          (SELECT count(*) FROM cafe.command_receipts WHERE aggregate_id=$1) AS commands,
          (SELECT count(*) FROM cafe.consumer_receipts WHERE consumer=$2) AS consumers,
          (SELECT count(*) FROM cafe.outbox_events WHERE aggregate_id=$1) AS events,
          (SELECT count(*) FROM cafe.realtime_publications WHERE aggregate_id=$1) AS realtime`,
        [id, m.consumer]);
        assert.deepEqual(counts, {commands: '1', consumers: receipt === 'command' ? '0' : '1', events: '0', realtime: '1'});
        await admin.query(update, [JSON.stringify(first), identity]);
        assert.deepEqual(await retry(), first);
        assert.equal(decisions, 0);
        assert.deepEqual(await queries.get(id), original);
        assert.equal((await admin.query('SELECT count(*) AS n FROM cafe.consumer_receipts WHERE consumer=$1',
          [m.consumer])).rows[0].n, '1');
      } finally {
        await Promise.all([db.pool.end(), admin.end()]);
      }
    });
  }
}

test('stored and proposed account identities must match the locked root', async () => {
  assert.ok(process.env.CAFE_DISPOSABLE_PROJECT?.startsWith('cafe-reference-test-'));
  const url = new URL(process.env.DATABASE_ADMIN_URL!);
  url.pathname = '/cafe_loyalty';
  const admin = new Pool({connectionString: url.toString()});
  const db = new PostgresContextDatabase('loyalty', process.env.LOYALTY_DATABASE_URL!);
  try {
    await db.verify();
    const commands = new PostgresAggregateCommandStore<AccountState>(db, 'account', restoreAccount);
    const queries = new PostgresAggregateQueries<AccountState>(db, 'account', restoreAccount);
    const id = randomUUID(), other = randomUUID();
    const state: AccountState = {id, collections: 0, grantsEarned: 0, stampBalance: 0};
    const m: Metadata = {id: randomUUID(), target: id, name: 'test.identity', correlation: randomUUID(),
      expected: 0, input: {}, consumer: 'loyalty.test-identity-'+id, sourceId: randomUUID(), sourceHash: 'wire'};
    const evidence = async () => (await admin.query(`SELECT
      (SELECT count(*) FROM cafe.command_receipts WHERE aggregate_id=$1) AS commands,
      (SELECT count(*) FROM cafe.consumer_receipts WHERE consumer=$2) AS consumers,
      (SELECT count(*) FROM cafe.realtime_publications WHERE aggregate_id=$1) AS realtime,
      (SELECT count(*) FROM cafe.outbox_events WHERE aggregate_id=$1) AS events`, [id, m.consumer])).rows[0];
    await assert.rejects(commands.execute(m, () => ({state: {...state, id: other}, status: 'active', changed: true})), CorruptState);
    assert.equal(await queries.get(id), undefined);
    assert.deepEqual(await evidence(), {commands: '0', consumers: '0', realtime: '0', events: '0'});
    await commands.execute(m, () => ({state, status: 'active', changed: true}));
    const repair = async (snapshot: AccountState) => {
      const client = await admin.connect();
      try {
        await client.query('BEGIN');
        await client.query("SET LOCAL session_replication_role='replica'");
        await client.query("UPDATE cafe.aggregates SET state=$1 WHERE kind='account' AND id=$2", [snapshot, id]);
        await client.query('COMMIT');
      } catch (error) {
        await client.query('ROLLBACK');
        throw error;
      } finally { client.release(); }
    };
    await repair({...state, id: other});
    const retry = {...m, id: randomUUID(), expected: 1, sourceId: randomUUID()};
    let decisions = 0;
    const execute = () => commands.execute(retry, loaded => {
      decisions++;
      return {state: loaded!.state, status: 'active', changed: false};
    });
    await assert.rejects(execute(), CorruptState);
    await assert.rejects(queries.get(id), CorruptState);
    assert.equal(decisions, 0);
    assert.deepEqual(await evidence(), {commands: '1', consumers: '1', realtime: '1', events: '0'});
    await repair(state);
    assert.equal((await execute()).version, 1);
    assert.deepEqual(await evidence(), {commands: '2', consumers: '2', realtime: '1', events: '0'});
  } finally {
    await Promise.all([db.pool.end(), admin.end()]);
  }
});
