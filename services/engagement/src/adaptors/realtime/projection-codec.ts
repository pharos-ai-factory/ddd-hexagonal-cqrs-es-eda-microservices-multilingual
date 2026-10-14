import protobuf from 'protobufjs';

import realtimeSchema from '../generated/realtime.json' with { type: 'json' };

import { newId } from '../../foundation/identity.js';

const realtimeType = protobuf.Root.fromJSON(realtimeSchema).lookupType('cafe.realtime.v1.Publication');

export function realtime(owner: string, kind: string, id: string, revision: number, state: unknown) {
  if (!realtimeType.fields[kind]) throw new Error('Undeclared realtime contract');
  const eventId = newId();
  const value = {
    eventId,
    contractVersion: 1,
    context: owner,
    aggregateKind: kind,
    aggregateId: id,
    revision,
    [kind]: state,
  };
  const error = realtimeType.verify(value);
  if (error) throw new Error(error);
  return { id: eventId, body: Buffer.from(realtimeType.encode(realtimeType.create(value)).finish()) };
}
