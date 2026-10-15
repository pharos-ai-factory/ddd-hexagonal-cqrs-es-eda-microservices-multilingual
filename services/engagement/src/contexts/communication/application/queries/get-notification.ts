import type { GetNotificationQuery } from '../../../../contracts/requests/generated/communication.js';
import type { GetNotificationResponse } from '../../../../contracts/requests/generated/communication-responses.js';
import type { NotificationReadRepository } from '../ports/notification-read-repository.js';
/** Preserve revision, missing-resource and pagination policy in a generated response. */
export class GetNotificationQueryHandler {
  constructor(private readRepository: NotificationReadRepository) {}
  async execute(query: GetNotificationQuery): Promise<GetNotificationResponse> {
    const value = await this.readRepository.get(query.id);
    return value ? { notification: value } : { error: { code: 'not_found', message: '' } };
  }
}
