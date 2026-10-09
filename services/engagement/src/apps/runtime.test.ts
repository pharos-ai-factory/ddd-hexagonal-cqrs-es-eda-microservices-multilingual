import {test} from 'node:test';
import assert from 'node:assert/strict';
import {completeSubscriptions, commandSubscription} from './runtime.js';
import {createLoyaltyContainer} from './composition/loyalty.js';
const definitions = [{consumer: 'loyalty.credit-collection', event: 'collection.order-collected', command: 'creditCollection'}];
test('manifest completeness rejects missing, duplicate and mislabelled consumer pairs', async () => {
  const container = createLoyaltyContainer('postgresql://unused@127.0.0.1:1/unused');
  const c = container.cradle;
  try {
    const pair = commandSubscription('collection.order-collected', c.creditCodec, definitions, e => e.customerId, c.collected, c.credit);
    assert.equal(completeSubscriptions('loyalty', definitions, pair).length, 2);
    for (const workers of [[], pair.slice(1), [...pair, ...pair], pair.map(w => ({...w, event: 'incorrect'}))])
      assert.throws(() => completeSubscriptions('loyalty', definitions, workers));
    assert.throws(() => completeSubscriptions('loyalty', [...definitions, ...definitions], pair));
    assert.throws(() => commandSubscription('collection.order-collected', c.creditCodec, [{...definitions[0]!, event: 'incorrect'}], e => e.customerId, c.collected, c.credit));
  } finally { await container.dispose(); }
});

// Compile-time regression: the event name controls the accepted handler payload.
function incompatibleHandler(container: ReturnType<typeof createLoyaltyContainer>) {
  const c = container.cradle;
  const wrong = {handle: async (_m: import('../foundation/application.js').Metadata, event: import('../contracts/events.js').RewardIssued) => ({aggregateId: event.rewardId, version: 0, status: 'queued'})};
  // @ts-expect-error RewardIssued cannot handle an OrderCollected subscription.
  commandSubscription('collection.order-collected', c.creditCodec, definitions, e => e.customerId, wrong, c.credit);
}
void incompatibleHandler;
