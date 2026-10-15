import { test } from 'node:test';
import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { PostgresContextDatabase } from './postgres-database.js';
import { PostgresAggregateTransaction } from './aggregate-transaction.js';
import { PostgresRealtimeEventBus } from './realtime-event-bus.js';
import { restoreRewardSnapshot } from '../../contexts/loyalty/adaptors/persistence/reward-snapshot.js';
import type { RewardState } from '../../contexts/loyalty/domain/reward.js';
import type { Metadata, RealtimeEvent } from '../../foundation/application.js';

void test('explicit publication, subscription redelivery and rootless rejections commit durably', async () => {
  const db = new PostgresContextDatabase('loyalty', process.env.LOYALTY_DATABASE_URL!);
  const id = randomUUID();
  const state: RewardState = {
    id,
    grantId: randomUUID(),
    customerId: randomUUID(),
    benefit: 'Tea',
    status: 'issued',
    expiresAt: '2027-01-01T12:00:00.000Z',
  };
  const unit = new PostgresAggregateTransaction<RewardState, RewardState>(
    db,
    'reward',
    restoreRewardSnapshot,
    (r) => r,
  );
  const m: Metadata = {
    id: randomUUID(),
    target: id,
    name: 'test.explicit',
    correlation: randomUUID(),
    expected: 0,
    input: {},
  };
  const count = async () =>
    (
      await db.pool.query<{ n: number }>(
        'SELECT count(*)::int AS n FROM cafe.realtime_publications WHERE aggregate_id=$1',
        [id],
      )
    ).rows[0]!.n;
  try {
    await unit.execute(m, async (r) => {
      await r.save(state);
      return { status: 'issued' };
    });
    assert.equal(await count(), 0);
    const next: RewardState = { ...state, status: 'redeemed', redeemedFor: randomUUID() };
    const changed = { ...m, id: randomUUID(), expected: 1 };
    await assert.rejects(
      unit.execute(changed, async (r) => {
        await r.save(next);
        return { status: 'redeemed', realtime: [{ snapshot: state }] };
      }),
      /differs from committed/,
    );
    const result = await unit.execute(changed, async (r) => {
      await r.save(next);
      return { status: 'redeemed', realtime: [{ snapshot: next }] };
    });
    assert.equal(result.version, 2);
    await unit.execute(changed, async () => assert.fail('Duplicate command executed'));
    assert.equal(await count(), 1);
    const source = {
      ...m,
      id: randomUUID(),
      consumer: 'loyalty.realtime-test',
      sourceId: randomUUID(),
      sourceHash: 'a'.repeat(64),
      sourceVersion: 1,
    };
    const bus = new PostgresRealtimeEventBus(db, source.consumer, 'reward');
    const event: RealtimeEvent = { source, revision: 2, snapshot: next };
    await bus.publish(event);
    await bus.publish(event);
    assert.equal(await count(), 2);
    await assert.rejects(
      bus.publish({ ...event, source: { ...source, sourceHash: 'b'.repeat(64) } }),
      /Conflicting/,
    );
    const missing = { ...m, id: randomUUID(), target: randomUUID() };
    const rejection = { code: 'not_found', message: 'The reward does not exist' };
    const outcome = await unit.execute(missing, async () => ({
      status: '',
      rejection,
      realtime: [{ rejection }],
    }));
    assert.equal(outcome.version, 0);
    assert.deepEqual(outcome.rejection, rejection);
    const rows = await db.pool.query(
      'SELECT id FROM cafe.realtime_publications WHERE aggregate_id=$1 AND message_kind=$2',
      [missing.target, 'operation_rejected'],
    );
    assert.equal(rows.rowCount, 1);
  } finally {
    await db.pool.end();
  }
});
