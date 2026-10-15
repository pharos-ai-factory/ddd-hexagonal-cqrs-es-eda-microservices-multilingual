import type { ListNotificationsQuery } from '../../../../contracts/requests/generated/communication.js';
import type { ListNotificationsResponse } from '../../../../contracts/requests/generated/communication-responses.js';
import type { NotificationReadRepository } from '../ports/notification-read-repository.js';
/** Preserve revision, missing-resource and pagination policy in a generated response. */
export class ListNotificationsQueryHandler {
  constructor(private readRepository: NotificationReadRepository) {}
  async execute(query: ListNotificationsQuery): Promise<ListNotificationsResponse> {
    const page = query.page
      ? await this.readRepository.page(query.page)
      : { items: await this.readRepository.list() };
    return {
      notifications: {
        items: page.items,
        paged: query.page !== undefined,
        ...(page.nextId !== undefined ? { nextId: page.nextId } : {}),
      },
    };
  }
}
