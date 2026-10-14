import { CreditCollectionCommandHandler } from '../contexts/loyalty/application/commands/credit-collection.js';
import { derivedId } from '../foundation/identity.js';
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { completeSubscriptions } from './runtime.js';
import { createLoyaltyContainer } from './composition/loyalty.js';
const definitions = [
  { consumer: 'loyalty.credit-collection', event: 'collection.order-collected', command: 'creditCollection' },
];
void test('manifest completeness rejects missing, duplicate and mislabelled consumer pairs', async () => {
  const container = createLoyaltyContainer('postgresql://unused@127.0.0.1:1/unused');
  const c = container.cradle;
  try {
    const pair = c.accountCommands.subscription(
      'collection.order-collected',
      c.creditCodec,
      definitions,
      (e) => e.customerId,
      c.collected,
      (repository) => new CreditCollectionCommandHandler(repository, derivedId),
    );
    assert.equal(completeSubscriptions('loyalty', definitions, pair).length, 2);
    for (const workers of [
      [],
      pair.slice(1),
      [...pair, ...pair],
      pair.map((w) => ({ ...w, event: 'incorrect' })),
    ])
      assert.throws(() => completeSubscriptions('loyalty', definitions, workers));
    assert.throws(() => completeSubscriptions('loyalty', [...definitions, ...definitions], pair));
    assert.throws(() =>
      c.accountCommands.subscription(
        'collection.order-collected',
        c.creditCodec,
        [{ ...definitions[0]!, event: 'incorrect' }],
        (e) => e.customerId,
        c.collected,
        (repository) => new CreditCollectionCommandHandler(repository, derivedId),
      ),
    );
  } finally {
    await container.dispose();
  }
});

// Compile-time regression: the event name controls the accepted handler payload.
function incompatibleHandler(container: ReturnType<typeof createLoyaltyContainer>) {
  const c = container.cradle;
  const wrong = {
    handle: async (
      _m: import('../foundation/application.js').Metadata,
      event: import('../contracts/events/generated/loyalty.js').RewardIssuedIntegrationEvent,
    ) => ({ aggregateId: event.rewardId, version: 0, status: 'queued' }),
  };
  c.accountCommands.subscription(
    'collection.order-collected',
    c.creditCodec,
    definitions,
    (e) => e.customerId,
    // @ts-expect-error RewardIssuedIntegrationEvent cannot handle an OrderCollectedIntegrationEvent subscription.
    wrong,
    (repository) => new CreditCollectionCommandHandler(repository, derivedId),
  );
}
void incompatibleHandler;

void test('publication subscriptions require one private worker and exactly one reaction mode', async () => {
  const { publicationSubscription } = await import('./runtime.js');
  const container = createLoyaltyContainer('postgresql://unused@127.0.0.1:1/unused');
  const declared = {
    consumer: 'loyalty.publish-reward-issued',
    event: 'loyalty.reward-issued.domain',
    publication: 'loyalty.reward-issued',
  };
  try {
    const worker = publicationSubscription(
      'loyalty',
      declared.consumer,
      'loyalty.reward-issued.domain',
      declared.publication,
      [declared],
      container.cradle.publishIssued,
    );
    assert.equal(worker.sourceTarget, true);
    assert.equal(completeSubscriptions('loyalty', [declared], [worker]).length, 1);
    for (const workers of [[], [worker, worker], [{ ...worker, owner: 'communication' }]])
      assert.throws(() => completeSubscriptions('loyalty', [declared], workers));
    assert.throws(() =>
      completeSubscriptions('loyalty', [{ ...declared, command: 'issueReward' }], [worker]),
    );
    assert.throws(() =>
      publicationSubscription(
        'loyalty',
        declared.consumer,
        'loyalty.reward-issued.domain',
        'wrong',
        [declared],
        container.cradle.publishIssued,
      ),
    );
  } finally {
    await container.dispose();
  }
});
