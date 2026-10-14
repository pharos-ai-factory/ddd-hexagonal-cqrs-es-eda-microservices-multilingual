import type { SubscriptionName, PrivateCommandName } from '../generated/message-names.js';
import protobuf from 'protobufjs';

import type { Metadata } from '../../foundation/application.js';

import { identifier } from '../../foundation/domain.js';

import { derivedId } from '../../foundation/identity.js';

/** Serialises an owner-private command and preserves its originating receipt material. */
export class InternalCommandCodec<C extends object> {
  private envelope: protobuf.Type;
  constructor(
    readonly owner: string,
    readonly consumer: SubscriptionName,
    readonly command: PrivateCommandName,
    schema: protobuf.INamespace,
    private parse: (value: unknown) => C,
  ) {
    if (!consumer.startsWith(owner + '.')) throw new Error('Foreign command subscription');
    this.envelope = protobuf.Root.fromJSON(schema).lookupType(`cafe.internal.${owner}.CommandEnvelope`);
    if (!this.envelope.fields[command]) throw new Error('Unknown owner command');
  }
  encode(m: Metadata, command: C): Buffer {
    this.validate(m);
    const value = {
      id: m.id,
      consumer: m.consumer,
      sourceEventId: m.sourceId,
      sourceHash: m.sourceHash,
      target: m.target,
      correlationId: m.correlation,
      receiptMaterial: Buffer.from(JSON.stringify(m.input)),
      [this.command]: this.parse(command),
    };
    const error = this.envelope.verify(value);
    if (error) throw new Error(error);
    return Buffer.from(this.envelope.encode(this.envelope.create(value)).finish());
  }
  decode(body: Buffer): { metadata: Metadata; payload: C } {
    const envelope = this.envelope.toObject(this.envelope.decode(body), { oneofs: true }) as Record<
      string,
      unknown
    >;
    if (envelope.command !== this.command || !Buffer.isBuffer(envelope.receiptMaterial))
      throw new Error('Wrong command payload');
    const metadata: Metadata = {
      id: identifier(String(envelope.id)),
      consumer: String(envelope.consumer),
      sourceId: identifier(String(envelope.sourceEventId)),
      sourceHash: String(envelope.sourceHash),
      target: identifier(String(envelope.target)),
      correlation: identifier(String(envelope.correlationId)),
      name: this.consumer,
      causation: identifier(String(envelope.sourceEventId)),
      input: JSON.parse(envelope.receiptMaterial.toString('utf8')) as unknown,
    };
    this.validate(metadata);
    return { metadata, payload: this.parse(envelope[this.command]) };
  }
  private validate(m: Metadata) {
    [m.id, m.target, m.correlation, m.sourceId].forEach((value) => identifier(String(value)));
    if (
      m.consumer !== this.consumer ||
      m.id !== derivedId(this.consumer, m.sourceId!) ||
      !/^[a-f0-9]{64}$/.test(m.sourceHash ?? '')
    )
      throw new Error('Invalid command identity');
  }
}

/** Header reader shared by the command relay; typed payload checks belong to its receiver. */
export function commandPublication(schema: protobuf.INamespace, owner: string) {
  const type = protobuf.Root.fromJSON(schema).lookupType(`cafe.internal.${owner}.CommandEnvelope`);
  return (body: Buffer) => {
    const value = type.toObject(type.decode(body)) as { id: string; consumer: string; correlationId: string };
    identifier(value.id);
    identifier(value.correlationId);
    if (!value.consumer.startsWith(owner + '.')) throw new Error('Foreign command destination');
    return {
      id: value.id,
      name: value.consumer + '.command',
      context: owner,
      correlationId: value.correlationId,
      visibility: 'command',
      payload: {},
    };
  };
}
