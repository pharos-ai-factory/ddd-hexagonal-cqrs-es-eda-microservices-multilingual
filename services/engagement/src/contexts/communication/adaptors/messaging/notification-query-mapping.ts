import { record, integer, text } from '../../../../adaptors/messaging/request-values.js';
import { identifier } from '../../../../foundation/domain.js';
import type { Loaded } from '../../../../foundation/application.js';
import type { Page } from '../../../../foundation/pagination.js';
import type { GetNotificationQuery } from '../../application/queries/get-notification.js';
import type { ListNotificationsQuery } from '../../application/queries/list-notifications.js';
import type { NotificationView } from '../../application/read-models/notification.js';
import { notification } from './notification-queries.js';

/** Decode the published item query into its application input. */
export function getNotification(value: unknown): GetNotificationQuery {
  return { id: identifier(text(record(value).id)) };
}

/** Preserve complete reads and explicitly requested pagination. */
export function listNotifications(value: unknown): ListNotificationsQuery {
  const fields = record(value);
  if (!('page' in fields)) return {};
  const page = record(fields.page);
  return {
    page: { limit: integer(page.limit), ...('after' in page ? { after: identifier(text(page.after)) } : {}) },
  };
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
