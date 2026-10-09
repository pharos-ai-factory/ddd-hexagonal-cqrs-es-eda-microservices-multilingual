import type {WriteRepository} from '../../../../foundation/write-repository.js';

import {MappedWriteRepository} from '../../../../adaptors/mapped-write-repository.js';
import {Notification, type NotificationState} from '../../domain/notification.js';

/** Restore authoritative aggregates and stage snapshots in the active PostgreSQL transaction. */
export class PostgresNotificationWriteRepository extends MappedWriteRepository<NotificationState, Notification> {
  constructor(source: WriteRepository<NotificationState>) {
    super(source, state => new Notification(state), aggregate => aggregate.snapshot());
  }
}
