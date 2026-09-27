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
export interface CommandPort<S> {
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
