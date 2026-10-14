import protobuf from 'protobufjs';

import eventSchema from '../generated/events.json' with { type: 'json' };

import catalogue from '../generated/catalogue.json' with { type: 'json' };

import { privateCodecs } from '../generated/private-event-registry.js';

import type { Metadata, Publication } from '../../foundation/application.js';

import { newId } from '../../foundation/identity.js';

import { contract, decodePayload, encodePayload } from './event-payload-codec.js';

import { validateMetadata, type WireEvent } from './event-envelope.js';

const eventType = protobuf.Root.fromJSON(eventSchema).lookupType('cafe.v1.Event');

function validate(event: WireEvent): void {
  const definition = contract(event.name);
  if (
    event.context !== definition.owner ||
    event.aggregateKind !== definition.kind ||
    event.visibility !== 'integration' ||
    event.contractVersion !== definition.version
  )
    throw new Error('Unknown contract or owner');
  validateMetadata(event);
}

/** Decode only declared public contracts or the owning context's private format. */
export function decode(body: Uint8Array): WireEvent {
  if (body.length > 256 * 1024) throw new Error('Event exceeds wire limit');
  const value = eventType.toObject(eventType.decode(body), { longs: Number }) as Record<string, unknown>;
  if (value.visibility === 'domain') {
    const codec = privateCodecs[String(value.name)];
    if (!codec) throw new Error('Foreign private message owner');
    return codec.decode(body);
  }
  const event = { ...value, payload: {} } as WireEvent;
  validate(event);
  if (!(value.payload instanceof Uint8Array) || !value.payload.length) throw new Error('Missing payload');
  event.payload = decodePayload(event.name, value.payload, event.aggregateId);
  return event;
}

/** Persist the returned bytes with the producing aggregate transaction. */
export function encode(
  owner: string,
  kind: string,
  id: string,
  version: number,
  m: Metadata,
  p: Publication,
  occurredAt?: string,
) {
  const definition = catalogue.find((item) => item.name === p.name);
  if (!definition) throw new Error('Unregistered publication');
  const event: WireEvent = {
    id: newId(),
    name: p.name,
    context: owner,
    visibility: definition.visibility,
    contractVersion: 'version' in definition ? definition.version : 1,
    aggregateKind: kind,
    aggregateId: id,
    aggregateVersion: version,
    correlationId: m.correlation,
    causationId: m.id,
    occurredAt: occurredAt ?? new Date().toISOString(),
    payload: p.payload,
  };
  if (event.visibility === 'domain') {
    const codec = privateCodecs[p.name];
    if (!codec) throw new Error('Foreign private message owner');
    return { event, body: Buffer.from(codec.encode(event)) };
  }
  validate(event);
  const value = { ...event, payload: encodePayload(event.name, event.payload, id) };
  const body = Buffer.from(eventType.encode(eventType.create(value)).finish());
  if (body.length > 256 * 1024) throw new Error('Event exceeds wire limit');
  return { event, body };
}

export { catalogue };

export type { WireEvent };
