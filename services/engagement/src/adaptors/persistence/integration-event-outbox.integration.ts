import {
  LoyaltyDomainEvent,
  LoyaltyIntegrationEvent,
  LoyaltyPrivateCommand,
  LoyaltySubscription,
} from '../generated/message-names.js';
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createHash, randomUUID } from 'node:crypto';
import { Pool } from 'pg';
import {
  createLoyaltyContainer,
  loyaltySubscriptions,
  loyaltyRequests,
} from '../../apps/composition/loyalty.js';
import type { RewardIssuedDomainEventPayload } from '../../contexts/loyalty/application/events.js';
import type { RewardIssuedIntegrationEvent } from '../../contracts/events/generated/loyalty.js';
import type { Metadata } from '../../foundation/application.js';
import { derivedId } from '../../foundation/identity.js';
import { published } from '../generated/event-types.js';
import { PostgresPublicationOutbox } from './integration-event-outbox.js';
import { decode } from '../messaging/event-codec.js';

void test('publication receipt and exact outbox bytes commit atomically per subscription', async () => {
  assert.match(process.env.CAFE_DISPOSABLE_PROJECT ?? '', /^cafe-reference-test-/);
  const container = createLoyaltyContainer(
    process.env.LOYALTY_DATABASE_URL!,
    () => new Date('2026-01-01T12:00:00.000Z'),
  );
  const c = container.cradle,
    db = c.database;
  const admin = new Pool({
    connectionString: process.env.DATABASE_ADMIN_URL!.replace(/\/[^/?]+(?=\?|$)/, '/cafe_loyalty'),
  });
  const grantId = randomUUID(),
    target = derivedId('reward', grantId),
    correlation = randomUUID();
  const consumer = LoyaltySubscription.PUBLISH_REWARD_ISSUED;
  const count = async (table: string, where: string, id: string) =>
    (await admin.query<{ n: number }>(`SELECT count(*)::int AS n FROM cafe.${table} WHERE ${where}=$1`, [id]))
      .rows[0]!.n;
  try {
    await db.verify();
    const issue = loyaltySubscriptions(c).find(
      (worker) => worker.consumer === String(LoyaltySubscription.ISSUE_REWARD) && worker.decodeCommand,
    )!;
    await issue.handle(
      { id: randomUUID(), target, name: LoyaltyPrivateCommand.ISSUE_REWARD, correlation, input: {} },
      { grantId, accountId: randomUUID(), benefit: 'one free drink', validDays: 30 },
    );
    const source = (
      await db.pool.query<{ body: Buffer }>('SELECT body FROM cafe.outbox_events WHERE aggregate_id=$1', [
        target,
      ])
    ).rows;
    assert.equal(source.length, 1);
    const event = decode(source[0]!.body);
    assert.equal(event.name, LoyaltyDomainEvent.REWARD_ISSUED);
    const payload = event.payload as RewardIssuedDomainEventPayload;
    const m: Metadata = {
      id: derivedId(consumer, event.id),
      target,
      name: consumer,
      correlation,
      consumer,
      sourceId: event.id,
      sourceHash: createHash('sha256').update(source[0]!.body).digest('hex'),
      sourceKind: event.aggregateKind,
      sourceVersion: event.aggregateVersion,
      sourceTime: event.occurredAt,
      input: payload,
      causation: event.id,
    };
    // A delayed subscriber publishes the immutable source revision even after the root advances.
    await loyaltyRequests(c).handle({
      kind: 'command',
      id: randomUUID(),
      owner: 'loyalty',
      name: 'redeemReward',
      metadata: { id: randomUUID(), target, name: 'redeemReward', correlation, input: {}, expected: 1 },
      body: { orderId: randomUUID() },
    });
    assert.equal((await c.getReward.execute({ id: target }))?.version, 2);
    await admin.query(
      `ALTER TABLE cafe.consumer_receipts ADD CONSTRAINT reject_publication_probe CHECK(event_id <> '${event.id}') NOT VALID`,
    );
    try {
      await assert.rejects(c.publishIssued.handle(m, payload), /reject_publication_probe/);
      assert.equal(await count('outbox_events', 'aggregate_id', target), 1);
      assert.equal(await count('consumer_receipts', 'event_id', event.id), 0);
    } finally {
      await admin.query('ALTER TABLE cafe.consumer_receipts DROP CONSTRAINT reject_publication_probe');
    }
    const outcomes = await Promise.all([
      c.publishIssued.handle(m, payload),
      c.publishIssued.handle(m, payload),
    ]);
    assert.deepEqual(outcomes[0], outcomes[1]);
    const outgoing = (
      await db.pool.query<{ body: Buffer; id: string }>(
        "SELECT body,id FROM cafe.outbox_events WHERE aggregate_id=$1 AND visibility='integration'",
        [target],
      )
    ).rows;
    assert.equal(outgoing.length, 1);
    const translated = decode(outgoing[0]!.body);
    assert.equal(translated.aggregateVersion, 1);
    assert.equal(translated.causationId, event.id);
    assert.equal(translated.occurredAt, event.occurredAt);
    assert.equal(translated.correlationId, correlation);
    assert.equal(await count('dispatches', 'event_id', outgoing[0]!.id), 1);
    await assert.rejects(
      c.publishIssued.handle({ ...m, sourceHash: 'b'.repeat(64) }, payload),
      /conflicting/,
    );
    // Lost acknowledgement: the recorded outcome is returned and original bytes remain unchanged.
    assert.deepEqual(await c.publishIssued.handle(m, payload), outcomes[0]);
    const again = (
      await db.pool.query<{ body: Buffer }>('SELECT body FROM cafe.outbox_events WHERE id=$1', [
        outgoing[0]!.id,
      ])
    ).rows[0]!;
    assert.deepEqual(again.body, outgoing[0]!.body);
    const audit = 'loyalty.audit-issued';
    const independent = new PostgresPublicationOutbox<RewardIssuedIntegrationEvent>(
      db,
      audit,
      'reward',
      (p) => published(LoyaltyIntegrationEvent.REWARD_ISSUED, p),
    );
    await independent.publish(
      { ...m, id: derivedId(audit, event.id), consumer: audit },
      translated.payload as RewardIssuedIntegrationEvent,
    );
    assert.equal(await count('consumer_receipts', 'event_id', event.id), 2);
    assert.equal(await count('outbox_events', 'aggregate_id', target), 3);
    assert.equal((await c.getReward.execute({ id: target }))?.version, 2);
  } finally {
    await container.dispose();
    await admin.end();
  }
});
