import type {Notification} from '../../domain/notification.js';
import type {Publication} from '../../../../foundation/application.js';

/** Map the saved aggregate's private facts to this owner's delivered messages. */
export function notificationPublications(aggregate: Notification): Publication[] {
  return aggregate.events().map(fact => ({name: 'communication.notification-requested', payload: {notificationId: fact.notificationId}}));
}
