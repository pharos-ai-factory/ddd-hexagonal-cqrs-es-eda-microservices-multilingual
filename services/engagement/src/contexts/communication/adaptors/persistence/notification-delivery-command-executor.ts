import type {DeliveryPort, Metadata, QueryPort} from '../../../../foundation/application.js';
import type {AggregateTransaction, CommandExecutor} from '../../../../adaptors/command-execution.js';
import {bindCommand} from '../../../../adaptors/command-execution.js';
import {Notification, type NotificationState} from '../../domain/notification.js';
import {DeliverNotificationCommandHandler, type DeliverNotificationCommand} from '../../application/commands/deliver-notification.js';

/** Recover idempotent provider acceptance before entering the local aggregate transaction. */
export class NotificationDeliveryCommandExecutor implements CommandExecutor<DeliverNotificationCommand> {
  constructor(private transaction: AggregateTransaction<Notification>, private queries: QueryPort<NotificationState>, private provider: DeliveryPort) {}
  async execute(metadata: Metadata, command: DeliverNotificationCommand) {
    const loaded = await this.queries.get(command.notificationId);
    if (!loaded) throw new Error('Requested notification is missing');
    const snapshot = new Notification(loaded.state).snapshot();
    const receipt = snapshot.status === 'sent' ? snapshot.providerReceipt! : await this.provider.deliver(snapshot);
    return bindCommand(this.transaction, repository => new DeliverNotificationCommandHandler(repository, receipt)).execute(metadata, command);
  }
}
