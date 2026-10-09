import type {AggregateTransaction} from '../../../../adaptors/command-execution.js';
import {test} from 'node:test';
import assert from 'node:assert/strict';
import type {Loaded, Metadata, QueryPort} from '../../../../foundation/application.js';
import type {NotificationState, Notification} from '../../domain/notification.js';
import {NotificationDeliveryCommandExecutor as DeliverNotificationCommandHandler} from '../../adaptors/persistence/notification-delivery-command-executor.js';

test('invalid stored notification content is rejected before any provider request', async () => {
  const id = '00000000-0000-4000-8000-000000000001';
  const loaded: Loaded<NotificationState> = {exists: true, version: 1,
    state: {id, recipient: id, subject: 'Ready', body: '', status: 'requested'}};
  const queries: QueryPort<NotificationState> = {
    get: async () => loaded, list: async () => [loaded],
  };
  const commands: AggregateTransaction<Notification> = {
    async execute() { throw new Error('Invalid state must fail before entering a command transaction'); },
  };
  let providerRequests = 0;
  const handler = new DeliverNotificationCommandHandler(commands, queries, {
    async deliver() { providerRequests++; return 'accepted'; },
  });
  const metadata: Metadata = {id, target: id, name: 'communication.deliver-notice', correlation: id, input: {}};
  await assert.rejects(handler.execute(metadata, {notificationId: id}));
  assert.equal(providerRequests, 0, 'Invalid authoritative state must not escape through a provider effect');
});
