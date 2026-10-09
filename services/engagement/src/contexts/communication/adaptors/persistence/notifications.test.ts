import {test} from 'node:test';
import assert from 'node:assert/strict';
import {CorruptState} from '../../../../foundation/domain.js';
const id = '00000000-0000-4000-8000-000000000001';
const other = '00000000-0000-4000-8000-000000000002';
import {restoreNotification, notificationView} from './notifications.js';
test('notification reads retain delivery receipts and reject inconsistent delivery state', () => {
  const state = {id, recipient: other, subject: 'Ready', body: 'Collect', status: 'sent', providerReceipt: 'receipt'};
  assert.deepEqual(notificationView(restoreNotification(state)), state);
  assert.throws(() => restoreNotification({...state, status: 'requested'}), CorruptState);
});
