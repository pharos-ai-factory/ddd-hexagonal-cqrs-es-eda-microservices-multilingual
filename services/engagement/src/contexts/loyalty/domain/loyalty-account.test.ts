import {LoyaltyAccount} from './loyalty-account.js';
import {test} from 'node:test';
import assert from 'node:assert/strict';
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
