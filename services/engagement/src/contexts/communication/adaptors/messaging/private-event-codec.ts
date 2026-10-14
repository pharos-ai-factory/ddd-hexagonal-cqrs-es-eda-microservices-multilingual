import type { NotificationRequested } from '../../application/events.js';
import protobuf from 'protobufjs';
import schema from './generated/private_messages.json' with { type: 'json' };
import { PrivateEventCodec } from '../../../../adaptors/private-event.js';
import { identifier } from '../../../../foundation/domain.js';

/** Communication accepts delivery work only for its source notification. */
export const privateEventCodec = new PrivateEventCodec(
  protobuf.Root.fromJSON(schema).lookupType('cafe.communication.internal.PrivateEvent'),
  {
    owner: 'communication',
    name: 'communication.notification-requested',
    kind: 'notification',
    payload: 'notificationRequested',
  },
  (payload, source) => {
    if (typeof payload.notificationId !== 'string') throw new Error('Missing notification identity');
    identifier(payload.notificationId);
    if (payload.notificationId !== source) throw new Error('Invalid source notification');
  },
);

/** Private payload association owned by communication. */
export type PrivateEventPayloads = { 'communication.notification-requested': NotificationRequested };
