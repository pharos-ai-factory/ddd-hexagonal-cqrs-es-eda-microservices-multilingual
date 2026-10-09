import type {PostgresContextDatabase} from '../../../../adaptors/postgres.js';
import {PostgresSnapshotReadRepository} from '../../../../adaptors/snapshot-read-repository.js';
import {MappedReadRepository} from '../../../../adaptors/mapped-read-repository.js';
import {restoreNotificationSnapshot} from './notification-snapshot.js';
import type {NotificationState} from '../../domain/notification.js';
import type {NotificationView} from '../../application/read-models/notification.js';

/** Select the stable application read fields independently from storage. */
export const notificationView = (s: NotificationState): NotificationView => ({id: s.id, recipient: s.recipient, subject: s.subject, body: s.body, status: s.status,
    ...(s.providerReceipt !== undefined ? {providerReceipt: s.providerReceipt} : {})});

/** PostgreSQL implementation of the communication NotificationReadRepository capability. */
export class PostgresNotificationReadRepository extends MappedReadRepository<NotificationState, NotificationView> {
  constructor(database: PostgresContextDatabase) {
    super(new PostgresSnapshotReadRepository(database, 'notification', restoreNotificationSnapshot), notificationView);
  }
}
