import { parseGetNotificationQuery, parseListNotificationsQuery } from './generated/request-payloads.js';
import { identifier } from '../../../../foundation/domain.js';
import type { GetNotificationQuery } from '../../../../contracts/requests/generated/communication.js';
import type { ListNotificationsQuery } from '../../../../contracts/requests/generated/communication.js';

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
