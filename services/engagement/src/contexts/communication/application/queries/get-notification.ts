import type { GetNotificationQuery } from '../../../../contracts/requests/generated/communication.js';
import type { NotificationReadRepository } from '../ports/notification-read-repository.js';

/** Execute the GetNotification use case through its context-owned read port. */
export class GetNotificationQueryHandler {
  constructor(private readRepository: NotificationReadRepository) {}
  async execute(query: GetNotificationQuery) {
    return this.readRepository.get(query.id);
  }
}
