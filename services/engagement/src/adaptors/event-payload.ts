import protobuf from 'protobufjs';
import schema from './generated/events.json' with { type: 'json' };
import contracts from './generated/event-contracts.json' with { type: 'json' };
import { identifier } from '../foundation/domain.js';

const root = protobuf.Root.fromJSON(schema);
type Rules = {
  uuid?: boolean;
  non_empty?: boolean;
  source_identity?: boolean;
  timestamp?: boolean;
  pattern?: string;
  minimum?: number;
  maximum?: number;
  min_items?: number;
  max_items?: number;
};
type Field = { name: string; json: string; type: number; message: string; repeated: boolean; rules: Rules };
type Message = { type: string; fields: Field[] };
const messages: readonly Message[] = contracts.messages;

/** Require an explicit timezone and a parseable instant at the message boundary. */
export function timestamp(value: unknown): boolean {
  if (typeof value !== 'string') return false;
  const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d+)?(?:Z|[+-](\d{2}):(\d{2}))$/.exec(
    value,
  );
  if (!match || match[0] !== value) return false;
  const year = Number(match[1]),
    month = Number(match[2]),
    day = Number(match[3]);
  const leap = year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
  const days = [31, leap ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31][month - 1];
  return (
    year >= 1 &&
    days !== undefined &&
    day >= 1 &&
    day <= days &&
    Number(match[4]) <= 23 &&
    Number(match[5]) <= 59 &&
    Number(match[6]) <= 59 &&
    Number(match[7] ?? 0) <= 23 &&
    Number(match[8] ?? 0) <= 59 &&
    Number.isFinite(Date.parse(value))
  );
}

function validate(type: string, value: Record<string, unknown>, source: string): void {
  const message = messages.find((item) => item.type === type);
  if (!message) throw new Error('Undeclared payload type');
  for (const field of message.fields) {
    const item = value[field.json],
      rule = field.rules;
    if (field.repeated) {
      if (
        !Array.isArray(item) ||
        (rule.min_items !== undefined && item.length < rule.min_items) ||
        (rule.max_items !== undefined && item.length > rule.max_items)
      )
        throw new Error('Invalid collection: ' + field.json);
      for (const child of item) validate(field.message, child as Record<string, unknown>, source);
    } else if (field.message) {
      if (!item || typeof item !== 'object') throw new Error('Missing payload message');
      validate(field.message, item as Record<string, unknown>, source);
    } else if (field.type === 9) {
      if (typeof item !== 'string') throw new Error('Missing string: ' + field.json);
      if (rule.uuid) identifier(item);
      if (rule.source_identity && item !== source)
        throw new Error('Payload does not identify its source aggregate');
      if (
        (rule.non_empty && !item.trim()) ||
        (rule.pattern && new RegExp(rule.pattern, 'u').exec(item)?.[0] !== item) ||
        (rule.timestamp && !timestamp(item))
      )
        throw new Error('Invalid string: ' + field.json);
    } else if (field.type === 8) {
      if (typeof item !== 'boolean') throw new Error('Invalid boolean');
    } else if (
      typeof item !== 'number' ||
      !Number.isSafeInteger(item) ||
      (field.type === 5 && (item < -2147483648 || item > 2147483647)) ||
      ((field.type === 4 || field.type === 13) && item < 0) ||
      (field.type === 13 && item > 4294967295) ||
      (rule.minimum !== undefined && item < rule.minimum) ||
      (rule.maximum !== undefined && item > rule.maximum)
    ) {
      throw new Error('Invalid portable integer: ' + field.json);
    }
  }
}

/** Look up only contracts included by this service's generation declaration. */
export function contract(name: string) {
  const definition = contracts.events.find((item) => item.name === name);
  if (!definition) throw new Error('Undeclared event contract');
  return definition;
}

export function decodePayload(name: string, body: Uint8Array, source: string): object {
  const type = root.lookupType(contract(name).type);
  const value = type.toObject(type.decode(body), { longs: Number, defaults: true }) as Record<
    string,
    unknown
  >;
  validate(type.fullName.slice(1), value, source);
  return value;
}

export function encodePayload(name: string, payload: object, source: string): Uint8Array {
  const type = root.lookupType(contract(name).type);
  const error = type.verify(payload);
  if (error) throw new Error(error);
  const value = type.toObject(type.create(payload), { longs: Number, defaults: true }) as Record<
    string,
    unknown
  >;
  validate(type.fullName.slice(1), value, source);
  return type.encode(type.create(payload)).finish();
}
