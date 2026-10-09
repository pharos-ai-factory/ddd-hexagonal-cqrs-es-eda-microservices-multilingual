import type {PostgresContextDatabase} from '../../../../adaptors/postgres.js';
import {PostgresAggregateQueries} from '../../../../adaptors/postgres.js';
import {MappedQueries} from '../../../../adaptors/mapped-queries.js';
import {storedObject} from '../../../../adaptors/stored-values.js';
import {Notification, type NotificationState} from '../../domain/notification.js';
import type {NotificationView} from '../../application/read-models/notification.js';

/** Restore the owning aggregate before exposing stored authority. */
export const restoreNotification = (value: unknown): NotificationState => new Notification(storedObject(value) as NotificationState).snapshot();
/** Select the stable application read fields independently from storage. */
export const notificationView = (s: NotificationState): NotificationView => ({id: s.id, recipient: s.recipient, subject: s.subject, body: s.body, status: s.status,
    ...(s.providerReceipt !== undefined ? {providerReceipt: s.providerReceipt} : {})});

/** PostgreSQL implementation of the communication NotificationReader capability. */
export class PostgresNotificationReader extends MappedQueries<NotificationState, NotificationView> {
  constructor(database: PostgresContextDatabase) {
    super(new PostgresAggregateQueries(database, 'notification', restoreNotification), notificationView);
  }
}
