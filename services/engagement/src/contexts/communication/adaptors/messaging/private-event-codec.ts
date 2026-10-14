import { CommunicationDomainEvent } from '../../../../adaptors/generated/message-names.js';
import type { NotificationRequestedDomainEventPayload } from '../../application/events.js';
import protobuf from 'protobufjs';
import schema from './generated/private_messages.json' with { type: 'json' };
import { PrivateEventCodec } from '../../../../adaptors/messaging/private-event-codec.js';
import { identifier } from '../../../../foundation/domain.js';

/** Communication accepts delivery work only for its source notification. */
export const privateEventCodec = new PrivateEventCodec(
  protobuf.Root.fromJSON(schema).lookupType('cafe.communication.internal.PrivateEvent'),
  {
    owner: 'communication',
    name: CommunicationDomainEvent.NOTIFICATION_REQUESTED,
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
export type PrivateEventPayloads = {
  [CommunicationDomainEvent.NOTIFICATION_REQUESTED]: NotificationRequestedDomainEventPayload;
};

export const privateEventCodecs = { [CommunicationDomainEvent.NOTIFICATION_REQUESTED]: privateEventCodec };
