import type {Page, PageRequest} from '../../../foundation/pagination.js';
import type {NotificationView} from '../application/read-models/notification.js';
import type {GetNotificationQueryHandler} from '../application/queries/get-notification.js';
import type {ListNotificationsQueryHandler} from '../application/queries/list-notifications.js';

/** Translate HTTP and RabbitMQ query arguments into named communication use cases. */
export function notificationQueryEndpoints(get: GetNotificationQueryHandler, list: ListNotificationsQueryHandler) {
  return {
    get: (id: string) => get.execute({id}),
    list: async () => (await list.execute({})).items,
    page: async (page: PageRequest): Promise<Page<NotificationView>> => list.execute({page}),
  };
}
