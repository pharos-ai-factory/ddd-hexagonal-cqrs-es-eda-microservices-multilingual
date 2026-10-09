import {test} from 'node:test';
import assert from 'node:assert/strict';
import {CorruptState} from '../../../../foundation/domain.js';
const id = '00000000-0000-4000-8000-000000000001';
const other = '00000000-0000-4000-8000-000000000002';
import {restoreNotificationSnapshot} from './notification-snapshot.js';
import {notificationView} from './notification-read-repository.js';
test('notification reads retain delivery receipts and reject inconsistent delivery state', () => {
  const state = {id, recipient: other, subject: 'Ready', body: 'Collect', status: 'sent', providerReceipt: 'receipt'};
  assert.deepEqual(notificationView(restoreNotificationSnapshot(state)), state);
  assert.throws(() => restoreNotificationSnapshot({...state, status: 'requested'}), CorruptState);
});
