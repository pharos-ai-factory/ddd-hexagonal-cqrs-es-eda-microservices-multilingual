import {test} from 'node:test';
import assert from 'node:assert/strict';
import {CorruptState} from '../../../../foundation/domain.js';
const id = '00000000-0000-4000-8000-000000000001';
const other = '00000000-0000-4000-8000-000000000002';
import {restoreAccountSnapshot} from './account-snapshot.js';
import {accountView} from './account-read-repository.js';
test('account restoration rejects corrupt grant ownership and mapping copies the grant', () => {
  const state = {id, collections: 3, grantsEarned: 1, stampBalance: 0,
    lastGrant: {id: other, accountId: id, benefit: 'one free drink', validDays: 7}};
  const restored = restoreAccountSnapshot(state), view = accountView(restored);
  assert.deepEqual(view, state);
  assert.notEqual(view.lastGrant, restored.lastGrant);
  assert.throws(() => restoreAccountSnapshot({...state, lastGrant: {...state.lastGrant, accountId: other}}), CorruptState);
});
