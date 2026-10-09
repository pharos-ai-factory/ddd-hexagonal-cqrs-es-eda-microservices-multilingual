import type {NotificationReader} from '../ports/notification-reader.js';

/** Select one notification. */
export type GetNotificationQuery = Readonly<{id: string}>;

/** Execute the GetNotification use case through its context-owned read port. */
export class GetNotificationQueryHandler {
  constructor(private reader: NotificationReader) {}
  async execute(query: GetNotificationQuery) { return this.reader.get(query.id); }
}
