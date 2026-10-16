import protobuf from 'protobufjs';

import schema from '../generated/requests.json' with { type: 'json' };

import { identifier as domainID } from '../../foundation/domain.js';

import { record, integer, text } from './request-values.js';

import type { Metadata } from '../../foundation/application.js';

export const identifier = (value: unknown) => domainID(text(value));

const root = protobuf.Root.fromJSON(schema);

const requestType = (owner: string) => root.lookupType('cafe.' + owner + '.requests.v1.Request');

export const replyType = (owner: string) => root.lookupType('cafe.' + owner + '.requests.v1.Reply');

export type Request = {
  kind: string;
  id: string;
  owner: string;
  name: string;
  body: Record<string, unknown>;
  metadata?: Metadata;
};

function selected(value: Record<string, unknown>): [string, Record<string, unknown>] {
  const name = text(value.payload);
  return [name, record(value[name])];
}

export function decodeRequest(body: Buffer, owner: string): Request {
  if (body.length > 65536) throw new Error('Request is too large');
  const wire = record(
    requestType(owner).toObject(requestType(owner).decode(body), { longs: Number, oneofs: true }),
  );
  if (wire.contractVersion !== 1 || wire.context !== owner) throw new Error('Invalid request envelope');
  const id = identifier(wire.requestId);
  const [kind, outer] = selected(wire),
    [name, payload] = selected(outer);
  const envelope = root.lookupType(
    'cafe.' + owner + '.requests.v1.' + (kind === 'command' ? 'Command' : 'Query'),
  );
  // The generated owner envelope defines the complete set of permitted operations.
  const field = envelope.fields[name];
  if (!field) throw new Error('Foreign context request');
  const request: Request = { kind, id, owner, name, body: payload };
  if (kind === 'command') {
    const metadata = record(outer.metadata);
    const expected = integer(metadata.expectedVersion);
    if (expected < 0 || !Number.isSafeInteger(expected)) throw new Error('Invalid expected version');
    request.metadata = {
      id: identifier(metadata.commandId),
      target: identifier(metadata.aggregateId),
      name: owner + '.' + name[0]!.toUpperCase() + name.slice(1),
      correlation: identifier(metadata.correlationId),
      ...(metadata.actorId ? { actorId: identifier(metadata.actorId) } : {}),
      expected,
      input: payload,
    };
    if (!(field?.resolvedType instanceof protobuf.Type)) throw new Error('Unknown command type');
    requiredInputs(field.resolvedType, payload);
  } else if (kind === 'query') {
    if ('id' in payload) identifier(payload.id);
    if ('page' in payload) {
      const page = record(payload.page),
        limit = integer(page.limit);
      if (limit < 1 || limit > 100) throw new Error('Invalid page limit');
      if ('after' in page) identifier(page.after);
    }
  } else throw new Error('Unknown request kind');
  return request;
}

export function requiredInputs(type: protobuf.Type, payload: Record<string, unknown>) {
  for (const [property, rule] of Object.entries(type.fields)) {
    if (rule.options?.['(cafe.requests.v1.required_input)'] && !(property in payload))
      throw new Error('Missing command field');
  }
}
