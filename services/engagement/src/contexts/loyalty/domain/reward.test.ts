import {Reward, RewardUnavailableDomainError} from './reward.js';
import {CorruptState, Rejection} from '../../../foundation/domain.js';
import {test} from 'node:test';
import assert from 'node:assert/strict';
const ids = Array.from({length: 5}, (_, index) => `00000000-0000-4000-8000-${String(index+1).padStart(12, '0')}`);
test('redemption is single use and validity has an exact supplied deadline', () => {
  const reward = Reward.issue(ids[0]!, {id: ids[1]!, accountId: ids[2]!, benefit: 'coffee', validDays: 7},
    new Date('2026-01-01T00:00:00Z'));
  const before = reward.snapshot();
  assert.deepEqual(reward.events(), [{type: 'RewardIssued', reward: before}]);
  assert.deepEqual(new Reward(before).events(), []);
  assert.throws(() => reward.redeem(ids[3]!, new Date('2026-01-08T00:00:00Z')), {code: 'reward_expired'});
  assert.deepEqual(reward.snapshot(), before);
  reward.redeem(ids[3]!, new Date('2026-01-07T23:59:59Z'));
  assert.throws(() => reward.redeem(ids[4]!, new Date('2026-01-07T23:59:59Z')),
    error => error instanceof RewardUnavailableDomainError && error.outcome().code === 'reward_unavailable');
  assert.equal(reward.snapshot().redeemedFor, ids[3]);
});
test('invalid issuance is a business rejection, but corrupt reward restoration is retryable', () => {
  const grant = {id: ids[1]!, accountId: ids[2]!, benefit: 'coffee', validDays: 7};
  const now = new Date('2026-01-01T00:00:00Z');
  assert.throws(() => Reward.issue('broken', grant, now), Rejection);
  const state = Reward.issue(ids[0]!, grant, now).snapshot();
  assert.throws(() => new Reward({...state, customerId: 'broken'}), CorruptState);
  assert.throws(() => new Reward({...state, benefit: ''}), CorruptState);
  assert.throws(() => new Reward({...state, status: 'redeemed'}), CorruptState);
});
