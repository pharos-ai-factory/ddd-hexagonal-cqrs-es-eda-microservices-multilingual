import type { CommandBusPort, Metadata } from '../../../../foundation/application.js';
import type { PickupOpenedIntegrationEvent } from '../../../../contracts/events/generated/collection.js';
import type { RequestNotificationCommand } from '../commands/request-notification.js';

/** Owns notification wording and durably requests a notification aggregate. */
export class PickupOpenedIntegrationEventHandler {
  constructor(private commands: CommandBusPort<RequestNotificationCommand>) {}
  handle(m: Metadata, event: PickupOpenedIntegrationEvent) {
    return this.commands.enqueue(m, {
      recipient: event.customerId,
      subject: 'Your drinks are ready',
      body: 'Collect your order using code ' + event.collectionCode,
    });
  }
}
