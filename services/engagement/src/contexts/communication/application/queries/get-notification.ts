import type {NotificationReadRepository} from '../ports/notification-read-repository.js';

/** Select one notification. */
export type GetNotificationQuery = Readonly<{id: string}>;

/** Execute the GetNotification use case through its context-owned read port. */
export class GetNotificationQueryHandler {
  constructor(private readRepository: NotificationReadRepository) {}
  async execute(query: GetNotificationQuery) { return this.readRepository.get(query.id); }
}
