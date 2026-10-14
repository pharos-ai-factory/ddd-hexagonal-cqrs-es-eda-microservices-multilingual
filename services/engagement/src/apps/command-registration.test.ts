import { test } from 'node:test';
import assert from 'node:assert/strict';
import { asValue } from 'awilix';
import { createLoyaltyContainer, loyaltyRequests } from './composition/loyalty.js';
import { createCommunicationContainer, communicationSubscriptions } from './composition/communication.js';
import type { AggregateTransaction } from '../adaptors/command-execution.js';
import { replyType } from '../adaptors/messaging/request-envelope.js';
import { Reward } from '../contexts/loyalty/domain/reward.js';
import { Notification, type NotificationState } from '../contexts/communication/domain/notification.js';
import type { Metadata, DeliveryPort, QueryPort } from '../foundation/application.js';

const id = '00000000-0000-4000-8000-000000000001';
const second = '00000000-0000-4000-8000-000000000002';
const url = 'postgresql://unused@127.0.0.1:1/unused';
const metadata = (target: string): Metadata => ({
  id: target,
  target,
  correlation: id,
  name: 'fixture',
  input: {},
});

void test('published registration parses the command and supplies the selected aggregate transaction', async () => {
  const container = createLoyaltyContainer(url, () => new Date('2026-01-01T00:00:00Z'));
  const captured: Metadata[] = [];
  const reward = new Reward({
    id,
    grantId: id,
    customerId: id,
    benefit: 'coffee',
    status: 'issued',
    expiresAt: '2026-01-02T00:00:00Z',
  });
  let saved = 0;
  const transaction: AggregateTransaction<Reward> = {
    execute: async (m, work) => {
      captured.push(m);
      const result = await work({
        get: async (target) => {
          assert.equal(target, id);
          return { exists: true, version: 1, state: reward };
        },
        save: async (value) => {
          assert.equal(value.snapshot().redeemedFor, second);
          saved++;
        },
      });
      return { aggregateId: m.target, version: 2, status: result.status };
    },
  };
  container.register({ rewardTransaction: asValue(transaction) });
  try {
    const registry = loyaltyRequests(container.cradle);
    const request = {
      owner: 'loyalty',
      kind: 'command',
      id,
      name: 'redeemReward',
      body: { orderId: second },
      metadata: { ...metadata(id), expected: 1 },
    };
    const type = replyType('loyalty');
    const reply = type.toObject(type.decode(await registry.handle(request)), { longs: Number }) as {
      outcome: { status: string };
    };
    assert.equal(reply.outcome.status, 'redeemed');
    assert.equal(saved, 1);
    assert.deepEqual(captured[0], { ...request.metadata, input: { orderId: second } });
    assert.throws(
      () =>
        container.cradle.rewardCommands.request(
          registry,
          'redeemReward',
          () => ({}),
          () => ({ execute: async () => ({ status: 'wrong' }) }),
        ),
      /Duplicate request handler/,
    );
    await registry.handle({ ...request, body: {} });
    assert.equal(captured.length, 1, 'invalid input reached the aggregate transaction');
  } finally {
    await container.dispose();
  }
});

void test('delivery registration prepares independent receipts before each transaction and propagates provider failures', async () => {
  const container = createCommunicationContainer(url, 'http://127.0.0.1:1', 'unused');
  const state = (target: string): NotificationState => ({
    id: target,
    recipient: id,
    subject: 'Ready',
    body: 'Collect your order',
    status: 'requested',
  });
  const saved: NotificationState[] = [];
  const prepared = new Set<string>();
  let active = 0,
    transactions = 0,
    unavailable = false;
  const queries: QueryPort<NotificationState> = {
    get: async (target) => {
      assert.equal(active, 0);
      return { exists: true, version: 1, state: state(target) };
    },
    list: async () => [],
  };
  const provider: DeliveryPort = {
    deliver: async (notice) => {
      assert.equal(active, 0);
      if (unavailable) throw new Error('provider unavailable');
      prepared.add(notice.id);
      return 'receipt-' + notice.id;
    },
  };
  const transaction: AggregateTransaction<Notification> = {
    execute: async (m, work) => {
      assert.ok(prepared.has(m.target));
      transactions++;
      active++;
      try {
        const result = await work({
          get: async (target) => ({ exists: true, version: 1, state: new Notification(state(target)) }),
          save: async (notice) => {
            saved.push(notice.snapshot());
          },
        });
        return { aggregateId: m.target, version: 2, status: result.status };
      } finally {
        active--;
      }
    },
  };
  container.register({
    notificationTransaction: asValue(transaction),
    notificationSnapshotReadRepository: asValue(queries),
    provider: asValue(provider),
  });
  try {
    const worker = communicationSubscriptions(container.cradle).find(
      (worker) => worker.consumer === 'communication.deliver-notice' && worker.decodeCommand,
    )!;
    const results = await Promise.all(
      [id, second].map((target) => worker.handle(metadata(target), { notificationId: target })),
    );
    assert.deepEqual(
      results.map((outcome) => outcome.status),
      ['sent', 'sent'],
    );
    assert.equal(transactions, 2);
    for (const notice of saved) assert.equal(notice.providerReceipt, 'receipt-' + notice.id);
    unavailable = true;
    await assert.rejects(worker.handle(metadata(id), { notificationId: id }), /provider unavailable/);
    assert.equal(transactions, 2);
  } finally {
    await container.dispose();
  }
});
