import {test} from 'node:test';
import assert from 'node:assert/strict';
import {LoyaltyAccount} from './account.js';
import {Reward} from './reward.js';
import {CorruptState, Rejection} from '../../../foundation/domain.js';

const ids = Array.from({length: 5}, (_, index) => `00000000-0000-4000-8000-${String(index+1).padStart(12, '0')}`);
test('the third collection atomically earns a grant, while issuing a reward is independent', () => {
  const account = LoyaltyAccount.open(ids[0]!);
  account.credit(ids[1]!, ids[4]!); account.credit(ids[2]!, ids[4]!);
  assert.equal(account.events().filter(f => f.type === 'RewardEarned').length, 0);
  account.credit(ids[3]!, ids[4]!);
  const state = account.snapshot();
  assert.equal(state.stampBalance, 0);
  assert.equal(state.grantsEarned, 1);
  assert.equal(state.collections, 3);
  assert.equal(state.lastGrant?.id, ids[4]);
  assert.equal(account.events().filter(f => f.type === 'RewardEarned').length, 1);
  const restored = new LoyaltyAccount(state);
  assert.deepEqual(restored.events(), []);
  (state as {collections: number}).collections = 99;
  assert.equal(account.snapshot().collections, 3);
});
test('invalid credit cannot change accounting or emit facts', () => {
  const account = LoyaltyAccount.open(ids[0]!);
  assert.throws(() => account.credit('invalid', ids[1]!));
  assert.equal(account.snapshot().collections, 0);
  assert.deepEqual(account.events(), []);
  assert.throws(() => new LoyaltyAccount({id: ids[0]!, collections: 3, grantsEarned: 0, stampBalance: 0}));
});
test('redemption is single use and validity has an exact supplied deadline', () => {
  const reward = Reward.issue(ids[0]!, {id: ids[1]!, accountId: ids[2]!, benefit: 'coffee', validDays: 7},
    new Date('2026-01-01T00:00:00Z'));
  const before = reward.snapshot();
  assert.throws(() => reward.redeem(ids[3]!, new Date('2026-01-08T00:00:00Z')), {code: 'reward_expired'});
  assert.deepEqual(reward.snapshot(), before);
  reward.redeem(ids[3]!, new Date('2026-01-07T23:59:59Z'));
  assert.throws(() => reward.redeem(ids[4]!, new Date('2026-01-07T23:59:59Z')), {code: 'reward_unavailable'});
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
