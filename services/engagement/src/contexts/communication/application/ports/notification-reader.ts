import type {PagedQueryPort} from '../../../../foundation/pagination.js';
import type {NotificationView} from '../read-models/notification.js';

/** Read notifications and their stable revisions through the communication boundary. */
export interface NotificationReader extends PagedQueryPort<NotificationView> {}
