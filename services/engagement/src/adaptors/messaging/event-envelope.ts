import { identifier } from '../../foundation/domain.js';
import { timestamp } from './event-payload-codec.js';

/** Technical publication metadata surrounds a decoded, plain payload. */
export type WireEvent = {
  id: string;
  name: string;
  context: string;
  visibility: string;
  contractVersion: number;
  aggregateKind: string;
  aggregateId: string;
  aggregateVersion: number;
  correlationId: string;
  causationId: string;
  occurredAt: string;
  payload: object;
};

export function validateMetadata(event: WireEvent): void {
  [event.id, event.aggregateId, event.correlationId, event.causationId].forEach(identifier);
  if (
    !Number.isSafeInteger(event.aggregateVersion) ||
    event.aggregateVersion < 1 ||
    !timestamp(event.occurredAt)
  )
    throw new Error('Invalid source metadata');
}
