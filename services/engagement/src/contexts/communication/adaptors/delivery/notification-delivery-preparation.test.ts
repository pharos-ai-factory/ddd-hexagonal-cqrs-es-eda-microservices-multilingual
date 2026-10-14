import { test } from 'node:test';
import assert from 'node:assert/strict';
import type { Loaded, QueryPort } from '../../../../foundation/application.js';
import type { NotificationState } from '../../domain/notification.js';
import { NotificationDeliveryPreparation } from './notification-delivery-preparation.js';

void test('invalid stored notification content is rejected before any provider request', async () => {
  const id = '00000000-0000-4000-8000-000000000001';
  const loaded: Loaded<NotificationState> = {
    exists: true,
    version: 1,
    state: { id, recipient: id, subject: 'Ready', body: '', status: 'requested' },
  };
  const queries: QueryPort<NotificationState> = {
    get: async () => loaded,
    list: async () => [loaded],
  };
  let providerRequests = 0;
  const handler = new NotificationDeliveryPreparation(queries, {
    async deliver() {
      providerRequests++;
      return 'accepted';
    },
  });
  await assert.rejects(handler.prepare({ notificationId: id }));
  assert.equal(providerRequests, 0, 'Invalid authoritative state must not escape through a provider effect');
});
