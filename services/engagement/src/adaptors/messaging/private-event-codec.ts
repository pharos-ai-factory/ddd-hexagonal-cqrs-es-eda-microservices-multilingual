import type { DomainEventName } from '../generated/message-names.js';
import protobuf from 'protobufjs';
import { validateMetadata, type WireEvent } from './event-envelope.js';

/** Private format bindings are supplied by the owning context. */
export class PrivateEventCodec {
  constructor(
    private readonly type: protobuf.Type,
    private readonly definition: { owner: string; name: DomainEventName; kind: string; payload: string },
    private readonly validatePayload: (payload: Record<string, unknown>, source: string) => void,
  ) {}

  private validate(event: WireEvent): void {
    const d = this.definition;
    if (
      event.context !== d.owner ||
      event.name !== String(d.name) ||
      event.aggregateKind !== d.kind ||
      event.visibility !== 'domain' ||
      event.contractVersion !== 1
    )
      throw new Error('Foreign private contract or owner');
    validateMetadata(event);
    this.validatePayload(event.payload as Record<string, unknown>, event.aggregateId);
  }

  decode(body: Uint8Array): WireEvent {
    const value = this.type.toObject(this.type.decode(body), { longs: Number, oneofs: true }) as Record<
      string,
      unknown
    >;
    if (value.payload !== this.definition.payload || !value[this.definition.payload])
      throw new Error('Missing private payload');
    const { formatRevision, ...metadata } = value;
    const event = {
      ...metadata,
      contractVersion: formatRevision,
      payload: value[this.definition.payload],
    } as WireEvent;
    this.validate(event);
    return event;
  }

  encode(event: WireEvent): Uint8Array {
    this.validate(event);
    const { payload, contractVersion, ...metadata } = event;
    const value = { ...metadata, formatRevision: contractVersion, [this.definition.payload]: payload };
    const error = this.type.verify(value);
    if (error) throw new Error(error);
    return this.type.encode(this.type.create(value)).finish();
  }
}
