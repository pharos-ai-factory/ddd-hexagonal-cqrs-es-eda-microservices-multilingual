import type {WriteRepository} from '../../../../foundation/write-repository.js';

import type {Notification} from '../../domain/notification.js';

/** Authoritative notification aggregate persistence inside its command transaction. */
export type NotificationWriteRepository = WriteRepository<Notification>;
