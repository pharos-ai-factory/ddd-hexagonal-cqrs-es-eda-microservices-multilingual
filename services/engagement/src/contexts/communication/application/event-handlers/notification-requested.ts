import type { CommandBusPort, Metadata } from '../../../../foundation/application.js';
import type { NotificationRequestedDomainEventPayload } from '../events.js';
import type { DeliverNotificationCommand } from '../commands/deliver-notification.js';

/** Requests delivery through its own command queue after notification creation commits. */
export class NotificationRequestedDomainEventHandler {
  constructor(private commands: CommandBusPort<DeliverNotificationCommand>) {}
  handle(m: Metadata, event: NotificationRequestedDomainEventPayload) {
    return this.commands.enqueue(m, { notificationId: event.notificationId });
  }
}
