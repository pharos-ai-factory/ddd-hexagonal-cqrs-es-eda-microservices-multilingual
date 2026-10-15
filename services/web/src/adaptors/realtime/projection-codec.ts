import protobuf from 'protobufjs/light';
import schema from '../generated/realtime.json';
import { owners, type Kind, type Projection, type Models } from '../../shared/realtime/model';
const publication = protobuf.Root.fromJSON(schema).lookupType('cafe.realtime.v1.Publication');
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;
export type OperationRejection = {
  type: 'operationRejected';
  eventId: string;
  commandId: string;
  correlationId: string;
  code: string;
  message: string;
};

export function decode(channel: string, bytes: unknown): Projection | OperationRejection {
  if (!(bytes instanceof Uint8Array) || bytes.length > 256 * 1024)
    throw new Error('Invalid binary publication');
  const value = publication.toObject(publication.decode(bytes), {
    longs: Number,
    defaults: true,
    oneofs: true,
  }) as {
    contractVersion: number;
    eventId: string;
    aggregateId: string;
    aggregateKind: string;
    context: string;
    snapshot: string;
    revision: number;
    operationRejected?: { commandId: string; correlationId: string; code: string; message: string };
  } & Partial<Models>;
  const kind = value.aggregateKind as Kind;
  const state = value[kind];
  if (
    value.contractVersion !== 1 ||
    !uuid.test(value.eventId) ||
    !uuid.test(value.aggregateId) ||
    !(kind in owners) ||
    value.context !== owners[kind] ||
    channel !== 'cafe:' + owners[kind] ||
    !Number.isSafeInteger(value.revision) ||
    value.revision < 0
  ) {
    throw new Error('Invalid realtime contract or source');
  }
  if (value.snapshot === 'operationRejected') {
    const rejection = value.operationRejected;
    if (
      !rejection ||
      !uuid.test(rejection.commandId) ||
      !uuid.test(rejection.correlationId) ||
      !rejection.code ||
      !rejection.message
    )
      throw new Error('Invalid operation rejection');
    return { type: 'operationRejected', eventId: value.eventId, ...rejection };
  }
  if (value.snapshot !== kind || !state || state.id !== value.aggregateId || value.revision < 1)
    throw new Error('Invalid realtime snapshot');
  return { kind, version: value.revision, state };
}
