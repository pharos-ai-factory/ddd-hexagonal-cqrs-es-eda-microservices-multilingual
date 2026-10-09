import type {DurableCommandPort, Metadata} from '../../../foundation/application.js';
import type {PickupOpened, RewardIssued} from '../../../contracts/events.js';
import type {NotificationRequested} from './events.js';
import type {RequestNotificationCommand, DeliverNotificationCommand} from './commands.js';

/** Owns notification wording and durably requests a notification aggregate. */
export class PickupOpenedIntegrationEventHandler {
  constructor(private commands: DurableCommandPort<RequestNotificationCommand>) {}
  handle(m: Metadata, event: PickupOpened) {
    return this.commands.enqueue(m, {recipient: event.customerId, subject: 'Your drinks are ready',
      body: 'Collect your order using code ' + event.collectionCode});
  }
}
/** Translates a published reward into Communication's durable notification intent. */
export class RewardIssuedIntegrationEventHandler {
  constructor(private commands: DurableCommandPort<RequestNotificationCommand>) {}
  handle(m: Metadata, event: RewardIssued) {
    return this.commands.enqueue(m, {recipient: event.customerId, subject: 'You earned a reward',
      body: event.benefit + '; valid until ' + event.expiresAt});
  }
}
/** Requests delivery through its own command queue after notification creation commits. */
export class NotificationRequestedDomainEventHandler {
  constructor(private commands: DurableCommandPort<DeliverNotificationCommand>) {}
  handle(m: Metadata, event: NotificationRequested) {
    return this.commands.enqueue(m, {notificationId: event.notificationId});
  }
}
