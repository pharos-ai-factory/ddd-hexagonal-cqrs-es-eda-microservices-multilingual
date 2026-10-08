import type {Loaded} from '../../../../foundation/application.js';
import type {communication as wire} from '../../../../adaptors/generated/request-types.js';
import type {NotificationState} from '../../domain/notification.js';

export function notification(loaded: Loaded<NotificationState>): wire.LoadedNotification {
  const s = loaded.state;
  return {exists: loaded.exists, version: loaded.version, state: {
    id: s.id, recipient: s.recipient, subject: s.subject, body: s.body, status: s.status,
    ...(s.providerReceipt ? {providerReceipt: s.providerReceipt} : {}),
  }};
}
