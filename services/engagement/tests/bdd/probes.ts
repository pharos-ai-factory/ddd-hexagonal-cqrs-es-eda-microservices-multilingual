import assert from 'node:assert/strict';
import type {Change, AggregateCommandPort, Loaded, Metadata, Outcome, Publication, QueryPort} from '../../src/foundation/application.js';
import {Rejection} from '../../src/foundation/domain.js';

// A decision probe calls the real handler. It deliberately does not implement
// durable receipts, transactions or delivery; the integration lane proves those.
export class CommandProbe<S> implements AggregateCommandPort<S>, QueryPort<S> {
  loaded: Loaded<S> | undefined;
  before: Loaded<S> | undefined;
  publications: Publication[] = [];
  outcome: Outcome | undefined;
  calls = 0;

  async execute(m: Metadata, decide: (loaded: Loaded<S> | undefined) => Change<S>): Promise<Outcome> {
    this.calls++;
    this.before = structuredClone(this.loaded);
    this.publications = [];
    this.outcome = {aggregateId: m.target, version: this.loaded?.version ?? 0, status: ''};
    try {
      const change = decide(structuredClone(this.loaded));
      this.publications = structuredClone(change.publications ?? []);
      if (change.changed) this.loaded = {exists: true, version: (this.loaded?.version ?? 0) + 1, state: structuredClone(change.state)};
      this.outcome = {...this.outcome, version: this.loaded?.version ?? 0, status: change.status};
    } catch (error) {
      if (!(error instanceof Rejection)) throw error;
      this.outcome.rejection = {code: error.code, message: error.message};
    }
    return this.outcome;
  }
  async get(_id: string) { return structuredClone(this.loaded); }
  async list() { return this.loaded ? [structuredClone(this.loaded)] : []; }
  succeeded() { assert.ok(this.outcome); assert.equal(this.outcome.rejection, undefined); }
  unchanged() { assert.deepEqual(this.loaded, this.before); assert.deepEqual(this.publications, []); }
}

export const customer = '00000000-0000-4000-8000-000000000001';
export const selectedOrder = '00000000-0000-4000-8000-000000000002';
export const notification = '00000000-0000-4000-8000-000000000003';
export function metadata(target: string): Metadata {
  return {id: customer, target, name: 'scenario-command', correlation: customer, input: {}};
}
