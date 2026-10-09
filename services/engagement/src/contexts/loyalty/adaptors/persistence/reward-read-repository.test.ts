import {test} from 'node:test';
import assert from 'node:assert/strict';
import {CorruptState} from '../../../../foundation/domain.js';
const id = '00000000-0000-4000-8000-000000000001';
const other = '00000000-0000-4000-8000-000000000002';
import {restoreRewardSnapshot} from './reward-snapshot.js';
import {rewardView} from './reward-read-repository.js';
test('reward read mapping retains redemption and rejects corrupt customer authority', () => {
  const state = {id, grantId: other, customerId: other, benefit: 'coffee', status: 'redeemed',
    redeemedFor: other, expiresAt: '2026-10-01T00:00:00Z'};
  assert.deepEqual(rewardView(restoreRewardSnapshot(state)), state);
  assert.throws(() => restoreRewardSnapshot({...state, customerId: 'invalid'}), CorruptState);
});
