import type {CommandPort, DeliveryPort, Metadata, QueryPort} from '../../../foundation/application.js';
import {Rejection} from '../../../foundation/domain.js';
import type {PickupOpened, RewardIssued} from '../../../contracts/events.js';
import type {NotificationRequested} from './events.js';
import {Notification, type NotificationState} from '../domain/notification.js';

export class RequestNotification {
  constructor(private notifications: CommandPort<NotificationState>) {}
  execute(m: Metadata, command: {recipient: string; subject: string; body: string}) {
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
export class PickupNotice {
  constructor(private notifications: CommandPort<NotificationState>) {}
  handle(m: Metadata, event: PickupOpened) {
    return new RequestNotification(this.notifications).execute(m, {recipient: event.customerId,
      subject: 'Your drinks are ready', body: 'Collect your order using code ' + event.collectionCode});
  }
}
export class RewardNotice {
  constructor(private notifications: CommandPort<NotificationState>) {}
  handle(m: Metadata, event: RewardIssued) {
    return new RequestNotification(this.notifications).execute(m, {recipient: event.customerId,
      subject: 'You earned a reward', body: event.benefit + '; valid until ' + event.expiresAt});
  }
}
export class DeliverNotification {
  constructor(private commands: CommandPort<NotificationState>, private queries: QueryPort<NotificationState>,
    private provider: DeliveryPort) {}
  async handle(m: Metadata, event: NotificationRequested) {
    const loaded = await this.queries.get(event.notificationId);
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
