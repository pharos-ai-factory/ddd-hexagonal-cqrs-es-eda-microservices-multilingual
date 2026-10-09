import protobuf from 'protobufjs/light';
import schema from '../../adaptors/generated/realtime.json';
import {owners, type Kind, type Projection} from './model';
const publication = protobuf.Root.fromJSON(schema).lookupType('cafe.realtime.v1.Publication');
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;
export function decode(channel: string, bytes: Uint8Array): Projection {
  if (!(bytes instanceof Uint8Array) || bytes.length > 256*1024) throw new Error('Invalid binary publication');
  const value = publication.toObject(publication.decode(bytes), {longs: Number, defaults: true, oneofs: true});
  const kind = value.aggregateKind as Kind;
  const state = value[kind];
  if (value.contractVersion !== 1 || !uuid.test(value.eventId) || !uuid.test(value.aggregateId) ||
      !(kind in owners) || value.context !== owners[kind] || channel !== 'cafe:'+owners[kind] ||
      value.snapshot !== kind || !state || state.id !== value.aggregateId ||
      !Number.isSafeInteger(value.revision) || value.revision < 1) {
    throw new Error('Invalid realtime contract or source');
  }
  return {kind, version: value.revision, state};
}
