import type {PageRequest} from '../../../../foundation/pagination.js';
import type {NotificationReader} from '../ports/notification-reader.js';

/** Select notifications or an explicitly requested page. */
export type ListNotificationsQuery = Readonly<{page?: PageRequest | undefined}>;

/** Execute the ListNotifications use case through its context-owned read port. */
export class ListNotificationsQueryHandler {
  constructor(private reader: NotificationReader) {}
  async execute(query: ListNotificationsQuery) { return query.page ? this.reader.page(query.page) : {items: await this.reader.list()}; }
}
