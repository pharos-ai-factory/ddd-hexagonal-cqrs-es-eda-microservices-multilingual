import type { ListNotificationsQuery } from '../../../../contracts/requests/generated/communication.js';
import type { NotificationReadRepository } from '../ports/notification-read-repository.js';

/** Execute the ListNotifications use case through its context-owned read port. */
export class ListNotificationsQueryHandler {
  constructor(private readRepository: NotificationReadRepository) {}
  async execute(query: ListNotificationsQuery) {
    return query.page ? this.readRepository.page(query.page) : { items: await this.readRepository.list() };
  }
}
