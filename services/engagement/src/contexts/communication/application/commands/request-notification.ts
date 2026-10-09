import type {AggregateCommandPort, Metadata} from '../../../../foundation/application.js';
import {Notification, type NotificationState} from '../../domain/notification.js';

/** Owner-local notification content, prepared before the durable hand-off. */
export type RequestNotificationCommand = {recipient: string; subject: string; body: string};

/** Creates one notification and records its private delivery-requested fact atomically. */
export class RequestNotificationCommandHandler {
  constructor(private notifications: AggregateCommandPort<NotificationState>) {}
  execute(m: Metadata, command: RequestNotificationCommand) {
    return this.notifications.execute({...m, input: command}, loaded => {
      if (loaded) {
        const state = new Notification(loaded.state).snapshot();
        return {state, status: state.status, changed: false};
      }
      const notification = Notification.request(m.target, command.recipient, command.subject, command.body);
      return {state: notification.snapshot(), status: 'requested', changed: true, publications: [{
        name: 'communication.notification-requested', payload: {notificationId: m.target}}]};
    });
  }
}
