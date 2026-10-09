import {test} from 'node:test';
import assert from 'node:assert/strict';
import {randomUUID} from 'node:crypto';
import {PostgresContextDatabase} from './postgres.js';
import {PostgresSnapshotReadRepository} from './snapshot-read-repository.js';
import {PostgresAggregateTransaction} from './aggregate-transaction.js';
import {restoreAccountSnapshot} from '../contexts/loyalty/adaptors/persistence/account-snapshot.js';
import type {AccountState} from '../contexts/loyalty/domain/loyalty-account.js';
import type {Metadata} from '../foundation/application.js';
import type {WriteRepository} from '../foundation/write-repository.js';

import {CorruptState, Rejection} from '../foundation/domain.js';

test('local command transaction bounds repository identity and lifetime and discards rejected saves', async () => {
  const db = new PostgresContextDatabase('loyalty', process.env.LOYALTY_DATABASE_URL!);
  const unit = new PostgresAggregateTransaction<AccountState, AccountState>(db, 'account', restoreAccountSnapshot, repo => repo);
  const queries = new PostgresSnapshotReadRepository(db, 'account', restoreAccountSnapshot);
  const id = randomUUID();
  const state = {id, collections: 0, stampBalance: 0, grantsEarned: 0};
  const metadata: Metadata = {id: randomUUID(), target: id, name: 'test.scope', correlation: id, expected: 0, input: {}};
  const escaped: WriteRepository<AccountState>[] = [];
  try {
    const outcome = await unit.execute(metadata, async repository => {
      escaped.push(repository);
      assert.equal(await repository.get(id), undefined);
      await assert.rejects(repository.get(randomUUID()), /another aggregate/);
      await assert.rejects(repository.save({...state, id: randomUUID()}), CorruptState);
      await repository.save(state);
      state.collections = 1; // Unsaved mutation cannot affect the staged snapshot.
      return {status: 'active'};
    });
    assert.equal(outcome.version, 1);
    await assert.rejects(escaped.at(-1)!.get(id), /outside its command transaction/);
    await assert.rejects(escaped.at(-1)!.save(state), /outside its command transaction/);
    const loaded = await queries.get(id);
    assert.equal(loaded?.state.collections, 0);
    const rejectedMetadata = {...metadata, id: randomUUID(), expected: 1};
    const rejected = await unit.execute(rejectedMetadata, async repository => {
      escaped.push(repository);
      await repository.save(state);
      throw new Rejection('fixture_rejection', 'Reject after staging');
    });
    assert.equal(rejected.rejection?.code, 'fixture_rejection');
    assert.equal(rejected.version, 1);
    assert.deepEqual(await queries.get(id), loaded);
    assert.deepEqual(await unit.execute(rejectedMetadata, async () => assert.fail('Rejected replay ran again')), rejected);
    await assert.rejects(escaped.at(-1)!.save(state), /outside its command transaction/);
    await assert.rejects(unit.execute({...metadata, id: randomUUID(), expected: 1}, async repository => {
      escaped.push(repository);
      await repository.save(state);
      throw new Error('Connection fixture failure');
    }), /Connection fixture failure/);
    await assert.rejects(escaped.at(-1)!.get(id), /outside its command transaction/);
    assert.deepEqual(await queries.get(id), loaded);
    const {rows: [counts]} = await db.pool.query(`SELECT
      (SELECT count(*) FROM cafe.command_receipts WHERE aggregate_id=$1) AS receipts,
      (SELECT count(*) FROM cafe.realtime_publications WHERE aggregate_id=$1) AS publications`, [id]);
    assert.deepEqual(counts, {receipts: '2', publications: '1'});
  } finally { await db.pool.end(); }
});
