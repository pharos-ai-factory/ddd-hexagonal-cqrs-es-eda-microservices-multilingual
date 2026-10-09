import type {DurableCommandPort, Metadata} from '../../../../foundation/application.js';
import type {PickupOpened} from '../../../../contracts/events.js';
import type {RequestNotificationCommand} from '../commands/request-notification.js';

/** Owns notification wording and durably requests a notification aggregate. */
export class PickupOpenedIntegrationEventHandler {
  constructor(private commands: DurableCommandPort<RequestNotificationCommand>) {}
  handle(m: Metadata, event: PickupOpened) {
    return this.commands.enqueue(m, {recipient: event.customerId, subject: 'Your drinks are ready',
      body: 'Collect your order using code ' + event.collectionCode});
  }
}
