import type {AggregateCommandPort, DeliveryPort, Metadata, QueryPort} from '../../../../foundation/application.js';
import {Rejection} from '../../../../foundation/domain.js';
import {Notification, type NotificationState} from '../../domain/notification.js';

export type DeliverNotificationCommand = {notificationId: string};

/** Performs idempotent provider delivery and records the result through one aggregate command. */
export class DeliverNotificationCommandHandler {
  constructor(private commands: AggregateCommandPort<NotificationState>, private queries: QueryPort<NotificationState>,
    private provider: DeliveryPort) {}
  async execute(m: Metadata, command: DeliverNotificationCommand) {
    const loaded = await this.queries.get(command.notificationId);
    if (!loaded) throw new Error('Requested notification is missing');
    const snapshot = new Notification(loaded.state).snapshot();
    // Provider I/O happens before the aggregate transaction and honours a stable key.
    const receipt = snapshot.status === 'sent' ? snapshot.providerReceipt! :
      await this.provider.deliver(snapshot);
    return this.commands.execute(m, current => {
      if (!current) throw new Rejection('not_found', 'The notification does not exist');
      const notice = new Notification(current.state);
      const changed = notice.recordDelivery(receipt);
      return {state: notice.snapshot(), status: 'sent', changed};
    });
  }
}
