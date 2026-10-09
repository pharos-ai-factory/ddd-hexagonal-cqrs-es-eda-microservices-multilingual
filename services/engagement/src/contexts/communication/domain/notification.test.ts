import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Notification} from './notification.js';
import {CorruptState, Rejection} from '../../../foundation/domain.js';
const id = '00000000-0000-4000-8000-000000000001';
test('delivery requires acceptance evidence and subsequent recordings are no-ops', () => {
  const notification = Notification.request(id, id, 'Ready', 'Collect your coffee');
  const before = notification.snapshot();
  assert.deepEqual(notification.events(), [{type: 'NotificationRequested', notificationId: id}]);
  assert.deepEqual(new Notification(before).events(), []);
  assert.throws(() => notification.recordDelivery(''), {code: 'missing_delivery_receipt'});
  assert.deepEqual(notification.snapshot(), before);
  assert.equal(notification.recordDelivery('receipt-1'), true);
  assert.equal(notification.recordDelivery('receipt-2'), false);
  assert.equal(notification.snapshot().providerReceipt, 'receipt-1');
});
test('invalid requested content is a business rejection, but corrupt restored content is retryable', () => {
  assert.throws(() => Notification.request(id, id, 'Ready', ''), Rejection);
  assert.throws(() => new Notification({id, recipient: id, subject: 'Ready', body: '', status: 'requested'}), CorruptState);
});
