import protobuf from 'protobufjs';
import loyaltyPrivate from '../contexts/loyalty/adaptors/messaging/generated/private_messages.json' with {type: 'json'};
import communicationPrivate from '../contexts/communication/adaptors/messaging/generated/private_messages.json' with {type: 'json'};
import eventSchema from './generated/events.json' with {type: 'json'};
import realtimeSchema from './generated/realtime.json' with {type: 'json'};
import catalogue from './generated/catalogue.json' with {type: 'json'};
import type {Metadata, Publication} from '../foundation/application.js';
import {identifier} from '../foundation/domain.js';
import {newId} from '../foundation/identity.js';

const eventType = protobuf.Root.fromJSON(eventSchema).lookupType('cafe.v1.Event');
const privateTypes: Record<string, protobuf.Type> = {
  loyalty: protobuf.Root.fromJSON(loyaltyPrivate).lookupType('cafe.loyalty.internal.PrivateEvent'),
  communication: protobuf.Root.fromJSON(communicationPrivate).lookupType('cafe.communication.internal.PrivateEvent'),
};
const realtimeType = protobuf.Root.fromJSON(realtimeSchema).lookupType('cafe.realtime.v1.Publication');
const payloadNames: Record<string, string> = {
  'menu.edition-published': 'menuPublished',
  'ordering.order-placed': 'orderPlaced', 'preparation.drinks-ready': 'drinksReady',
  'collection.pickup-opened': 'pickupOpened', 'collection.order-collected': 'orderCollected',
  'loyalty.reward-earned': 'rewardEarned', 'loyalty.reward-issued': 'rewardIssued',
  'communication.notification-requested': 'notificationRequested',
};
const payloadIdentities: Record<string, readonly string[]> = {
  'menu.edition-published': ['editionId'],
  'ordering.order-placed': ['orderId', 'customerId', 'editionId'],
  'preparation.drinks-ready': ['orderId', 'customerId'],
  'collection.pickup-opened': ['pickupId', 'orderId', 'customerId'],
  'collection.order-collected': ['orderId', 'customerId'],
  'loyalty.reward-earned': ['accountId', 'grantId'],
  'loyalty.reward-issued': ['rewardId', 'customerId'],
  'communication.notification-requested': ['notificationId'],
};
const sourceIdentities: Record<string, string> = {
  'menu.edition-published': 'editionId',
  'ordering.order-placed': 'orderId', 'collection.pickup-opened': 'pickupId',
  'loyalty.reward-earned': 'accountId', 'loyalty.reward-issued': 'rewardId',
  'communication.notification-requested': 'notificationId',
};
export type WireEvent = {
  id: string; name: string; context: string; visibility: string; contractVersion: number;
  aggregateKind: string; aggregateId: string; aggregateVersion: number;
  correlationId: string; causationId: string; occurredAt: string; payload: object;
};
function timestamp(value: unknown): boolean {
  return typeof value === 'string' &&
    /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/.test(value) &&
    Number.isFinite(Date.parse(value));
}
function validate(event: WireEvent) {
  const definition = catalogue.find(item => item.name === event.name);
  if (!definition || definition.owner !== event.context || definition.kind !== event.aggregateKind ||
      definition.visibility !== event.visibility || event.contractVersion !== 1) throw new Error('Unknown contract or owner');
  [event.id, event.aggregateId, event.correlationId, event.causationId].forEach(identifier);
  if (!Number.isSafeInteger(event.aggregateVersion) || event.aggregateVersion < 1 ||
      !timestamp(event.occurredAt)) throw new Error('Invalid source metadata');
  const p = event.payload as Record<string, unknown>;
  for (const key of payloadIdentities[event.name]!) {
    const value = p[key];
    if (typeof value !== 'string') throw new Error('Missing payload identity: '+key);
    identifier(value);
  }
  const source = sourceIdentities[event.name];
  if (source && p[source] !== event.aggregateId) {
    throw new Error('Payload does not identify its source aggregate');
  }
  if (event.name === 'loyalty.reward-earned' &&
      (typeof p.benefit !== 'string' || !p.benefit || !Number.isInteger(p.validDays) ||
        Number(p.validDays) < 1 || Number(p.validDays) > 30)) throw new Error('Invalid grant');
  if (event.name === 'loyalty.reward-issued' &&
      (typeof p.benefit !== 'string' || !p.benefit || !timestamp(p.expiresAt))) throw new Error('Invalid issued reward');
  if (event.name === 'collection.pickup-opened' &&
      !/^[A-Z0-9]{6}$/.test(String(p.collectionCode))) throw new Error('Invalid pickup');
}
export function decode(body: Uint8Array): WireEvent {
  if (body.length > 256*1024) throw new Error('Event exceeds wire limit');
  // Header field numbers are shared by the historical internal delivery format.
  let object = eventType.toObject(eventType.decode(body), {longs: Number, oneofs: true}) as Record<string, unknown>;
  if (object.visibility === 'domain') {
    const type = privateTypes[String(object.context)];
    if (!type) throw new Error('Foreign private message owner');
    object = type.toObject(type.decode(body), {longs: Number, oneofs: true}) as Record<string, unknown>;
    object.contractVersion = object.formatRevision;
    delete object.formatRevision;
  }
  const name = String(object.name), key = payloadNames[name];
  if (!key || object.payload !== key || !object[key]) throw new Error('Mismatched payload');
  const event = {...object, payload: object[key]} as WireEvent;
  validate(event);
  return event;
}
export function encode(owner: string, kind: string, id: string, version: number, m: Metadata, p: Publication) {
  const definition = catalogue.find(item => item.name === p.name);
  if (!definition) throw new Error('Unregistered publication');
  const event: WireEvent = {id: newId(), name: p.name, context: owner, visibility: definition.visibility,
    contractVersion: 1, aggregateKind: kind, aggregateId: id, aggregateVersion: version,
    correlationId: m.correlation, causationId: m.id, occurredAt: new Date().toISOString(), payload: p.payload};
  validate(event);
  const {payload, ...envelope} = event;
  const type = event.visibility === 'domain' ? privateTypes[owner] : eventType;
  if (!type) throw new Error('Foreign private message owner');
  const value: Record<string, unknown> = {...envelope, [payloadNames[event.name]!]: payload};
  if (event.visibility === 'domain') { value.formatRevision = value.contractVersion; delete value.contractVersion; }
  const error = type.verify(value);
  if (error) throw new Error(error);
  return {event, body: Buffer.from(type.encode(type.create(value)).finish())};
}
export function realtime(owner: string, kind: string, id: string, revision: number, state: unknown) {
  const eventId = newId();
  const value = {eventId, contractVersion: 1, context: owner, aggregateKind: kind, aggregateId: id, revision, [kind]: state};
  const error = realtimeType.verify(value);
  if (error) throw new Error(error);
  return {id: eventId, body: Buffer.from(realtimeType.encode(realtimeType.create(value)).finish())};
}
export {catalogue};
