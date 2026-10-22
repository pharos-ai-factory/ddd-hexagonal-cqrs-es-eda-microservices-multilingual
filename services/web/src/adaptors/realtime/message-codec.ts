import protobuf from 'protobufjs/light';
import schema from '../generated/realtime.json';
import { owners, type Kind, type Models } from '../../shared/realtime/model';
import type { RealtimeMessage } from '../../shared/realtime/messages';

const publication = protobuf.Root.fromJSON(schema).lookupType('cafe.realtime.v1.Publication');
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

type WirePublication = {
  contractVersion: number;
  eventId: string;
  aggregateId: string;
  aggregateKind: string;
  context: string;
  snapshot: string;
  revision: number;
  target?: { target?: string; context?: string; principal?: string };
  operationRejected?: { commandId: string; correlationId: string; code: string; message: string };
} & Partial<Models>;

/** Decode one authorised Centrifugo publication into the generated-style message catalogue. */
export function decodeRealtimeMessage(channel: string, bytes: unknown): RealtimeMessage {
  if (!(bytes instanceof Uint8Array) || bytes.length > 256 * 1024)
    throw new Error('Invalid binary publication');
  const value = publication.toObject(publication.decode(bytes), {
    longs: Number,
    defaults: true,
    oneofs: true,
  }) as WirePublication;
  const kind = value.aggregateKind as Kind;
  if (
    value.contractVersion !== 1 ||
    !uuid.test(value.eventId) ||
    !uuid.test(value.aggregateId) ||
    !(kind in owners) ||
    value.context !== owners[kind] ||
    !Number.isSafeInteger(value.revision) ||
    value.revision < 0
  ) {
    throw new Error('Invalid realtime contract or source');
  }
  if (value.snapshot === 'operationRejected') {
    const rejection = value.operationRejected;
    const principal = value.target?.principal;
    if (
      value.target?.target !== 'principal' ||
      !principal ||
      !uuid.test(principal) ||
      channel !== 'principal:' + principal ||
      !rejection ||
      !uuid.test(rejection.commandId) ||
      !uuid.test(rejection.correlationId) ||
      !rejection.code ||
      !rejection.message
    ) {
      throw new Error('Invalid operation rejection');
    }
    return {
      type: 'operationRejected',
      eventId: value.eventId,
      target: { kind: 'principal', id: principal },
      ...rejection,
    };
  }
  const state = value[kind];
  const context = value.target?.context;
  if (
    value.snapshot !== kind ||
    !state ||
    state.id !== value.aggregateId ||
    value.revision < 1 ||
    value.target?.target !== 'context' ||
    context !== owners[kind] ||
    channel !== 'cafe:' + context
  ) {
    throw new Error('Invalid realtime state message');
  }
  return {
    type: `${kind}Changed`,
    eventId: value.eventId,
    target: { kind: 'context', id: context },
    revision: value.revision,
    state,
  } as RealtimeMessage;
}
