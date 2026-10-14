import type { DeliveryPort, QueryPort } from '../../../../foundation/application.js';
import type { NotificationState } from '../../domain/notification.js';
import { restoreNotificationSnapshot } from '../persistence/notification-snapshot.js';
import type { DeliverNotificationCommand } from '../../application/commands/deliver-notification.js';

/** Recover idempotent provider acceptance before the command opens its aggregate transaction. */
export class NotificationDeliveryPreparation {
  constructor(
    private queries: QueryPort<NotificationState>,
    private provider: DeliveryPort,
  ) {}

  async prepare(command: DeliverNotificationCommand): Promise<string> {
    const loaded = await this.queries.get(command.notificationId);
    if (!loaded) throw new Error('Requested notification is missing');
    const snapshot = restoreNotificationSnapshot(loaded.state);
    return snapshot.status === 'sent' ? snapshot.providerReceipt! : this.provider.deliver(snapshot);
  }
}
