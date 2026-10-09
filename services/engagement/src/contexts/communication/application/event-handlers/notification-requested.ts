import type {DurableCommandPort, Metadata} from '../../../../foundation/application.js';
import type {NotificationRequested} from '../events.js';
import type {DeliverNotificationCommand} from '../commands/deliver-notification.js';

/** Requests delivery through its own command queue after notification creation commits. */
export class NotificationRequestedDomainEventHandler {
  constructor(private commands: DurableCommandPort<DeliverNotificationCommand>) {}
  handle(m: Metadata, event: NotificationRequested) {
    return this.commands.enqueue(m, {notificationId: event.notificationId});
  }
}
