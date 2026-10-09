import {Rejection} from './domain.js';

// Expected use-case failures share the recorded rejection contract with domain errors.
/** Base type for expected application-policy rejections recorded with command outcomes. */
export abstract class ApplicationError extends Rejection {}

/** Records rejection when the caller expected a different aggregate revision. */
export class VersionConflictApplicationError extends ApplicationError {
  constructor() { super('version_conflict', 'The expected version is stale'); }
}

export type Metadata = {
  id: string; target: string; name: string; correlation: string; input: unknown;
  expected?: number; causation?: string; consumer?: string; sourceId?: string; sourceHash?: string;
};
export type Outcome = {
  aggregateId: string; version: number; status: string;
  rejection?: {code: string; message: string};
};
export type Loaded<S> = {exists: boolean; version: number; state: S};
export type Publication = {name: string; payload: object};
export type Change<S> = {state: S; status: string; changed: boolean; publications?: Publication[]};
export interface AggregateCommandPort<S> {
  execute(metadata: Metadata, decide: (loaded: Loaded<S> | undefined) => Change<S>): Promise<Outcome>;
}
export interface QueryPort<S> {
  get(id: string): Promise<Loaded<S> | undefined>;
  list(): Promise<Loaded<S>[]>;
}
export type IdentityFactory = (purpose: string, key: string) => string;
export interface DeliveryPort {
  deliver(message: {id: string; recipient: string; subject: string; body: string}): Promise<string>;
}

/** Durably accepts an owner-local command; completion belongs to its command consumer. */
export interface DurableCommandPort<C> {
  enqueue(metadata: Metadata, command: C): Promise<Outcome>;
}
