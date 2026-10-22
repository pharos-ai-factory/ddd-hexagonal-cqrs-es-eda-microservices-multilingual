import { Rejection } from './domain.js';

// Expected use-case failures share the recorded rejection contract with domain errors.
/** Base type for expected application-policy rejections recorded with command outcomes. */
export abstract class ApplicationError extends Rejection {}

/** Records rejection when the caller expected a different aggregate revision. */
export class VersionConflictApplicationError extends ApplicationError {
  constructor() {
    super('version_conflict', 'The expected version is stale');
  }
}

export type Metadata = {
  id: string;
  target: string;
  name: string;
  correlation: string;
  actorId?: string;
  input: unknown;
  expected?: number;
  causation?: string;
  consumer?: string;
  sourceId?: string;
  sourceHash?: string;
  sourceKind?: string;
  sourceVersion?: number;
  sourceTime?: string;
};
import type { Outcome } from '../contracts/requests/generated/shared-responses.js';
export type { Outcome } from '../contracts/requests/generated/shared-responses.js';
export type Loaded<S> = { exists: boolean; version: number; state: S };
export type Publication = { name: string; payload: object };
export type Change<S> = { state: S; status: string; changed: boolean; publications?: Publication[] };
export interface QueryPort<S> {
  get(id: string): Promise<Loaded<S> | undefined>;
  list(): Promise<Loaded<S>[]>;
}
export type IdentityFactory = (purpose: string, key: string) => string;
export interface DeliveryPort {
  deliver(message: { id: string; recipient: string; subject: string; body: string }): Promise<string>;
}

/** Durably accepts an owner-local command; completion belongs to its command consumer. */
export interface CommandBusPort<C> {
  enqueue(metadata: Metadata, command: C): Promise<Outcome>;
}

/** Atomically records one subscription receipt and the exact outgoing integration event. */
export interface IntegrationEventBusPort<P> {
  publish(metadata: Metadata, payload: P): Promise<Outcome>;
}

/** Dispatch a typed query to one registered handler and return its result. */
export interface QueryBusPort<Q, R> {
  ask(query: Q): Promise<R>;
}
/** Stage private domain facts inside their owning transaction. */
export interface DomainEventBusPort<E> {
  publish(event: E): Promise<void>;
}
/** Explicitly stage a browser message; the surrounding transaction commits it. */
export interface RealtimeBusPort<E> {
  publish(event: E): Promise<void>;
}
/** Application recipient resolved to a transport channel by an outer adaptor. */
export type RealtimeTarget = Readonly<{
  kind: 'context' | 'principal';
  id: string;
}>;
export const realtimeTarget = Object.freeze({
  context: (owner: string): RealtimeTarget => ({ kind: 'context', id: owner }),
  principal: (principalId: string): RealtimeTarget => ({ kind: 'principal', id: principalId }),
});
/** A typed delivery target and one authoritative payload. */
export type RealtimeMessage = { target: RealtimeTarget; source?: Metadata; revision?: number } & (
  | { snapshot: object; rejection?: never }
  | { rejection: { code: string; message: string }; snapshot?: never }
);
export const realtimeMessage = Object.freeze({
  snapshot: (target: RealtimeTarget, snapshot: object): RealtimeMessage => ({ target, snapshot }),
  rejection: (
    target: RealtimeTarget,
    rejection: Readonly<{ code: string; message: string }>,
  ): RealtimeMessage => ({ target, rejection }),
});
