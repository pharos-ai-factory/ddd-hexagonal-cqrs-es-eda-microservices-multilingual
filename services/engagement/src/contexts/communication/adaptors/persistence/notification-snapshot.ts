import {storedObject} from '../../../../adaptors/stored-values.js';
import {Notification, type NotificationState} from '../../domain/notification.js';

/** Restore the owning aggregate before exposing stored authority. */
export const restoreNotificationSnapshot = (value: unknown): NotificationState => new Notification(storedObject(value) as NotificationState).snapshot();
