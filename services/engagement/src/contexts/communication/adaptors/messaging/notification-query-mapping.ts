import { parseGetNotificationQuery, parseListNotificationsQuery } from './generated/request-payloads.js';
import { identifier } from '../../../../foundation/domain.js';
import type { Loaded } from '../../../../foundation/application.js';
import type { Page } from '../../../../foundation/pagination.js';
import type { GetNotificationQuery } from '../../../../contracts/requests/generated/communication.js';
import type { ListNotificationsQuery } from '../../../../contracts/requests/generated/communication.js';
import type { NotificationView } from '../../application/read-models/notification.js';
import { notification } from './notification-queries.js';

/** Decode the published item query into its application input. */
export function getNotification(value: unknown): GetNotificationQuery {
  const query = parseGetNotificationQuery(value);
  identifier(query.id);
  return query;
}

/** Preserve complete reads and explicitly requested pagination. */
export function listNotifications(value: unknown): ListNotificationsQuery {
  const query = parseListNotificationsQuery(value);
  if (query.page?.after !== undefined) identifier(query.page.after);
  return query;
}

/** Map the resource revision or the existing missing-resource response. */
export function getReply(_query: GetNotificationQuery, loaded: Loaded<NotificationView> | undefined) {
  return loaded ? { notification: notification(loaded) } : { error: { code: 'not_found' } };
}

/** Retain pagination and its continuation from the handler result. */
export function listReply(query: ListNotificationsQuery, page: Page<NotificationView>) {
  return {
    notifications: {
      items: page.items.map(notification),
      paged: query.page !== undefined,
      ...(page.nextId ? { nextId: page.nextId } : {}),
    },
  };
}
