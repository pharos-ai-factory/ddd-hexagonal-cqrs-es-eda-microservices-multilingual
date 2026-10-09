import type {PoolClient} from 'pg';
import type {Loaded} from '../foundation/application.js';
import type {WriteRepository} from '../foundation/write-repository.js';

import {checkRootIdentity} from './receipts.js';

/** Own aggregate SQL and stage one authoritative root inside its command transaction. */
export class PostgresSnapshotWriteRepository<S> implements WriteRepository<S> {
  private active = true;
  private loaded: Loaded<S> | undefined;
  pending: S | undefined;
  get version() { return this.loaded?.version ?? 0; }
  constructor(private client: PoolClient, private kind: string, private target: string,
    private restore: (value: unknown) => S) {}
  async load() {
    const {rows: [row]} = await this.client.query('SELECT version,state FROM cafe.aggregates WHERE kind=$1 AND id=$2 FOR UPDATE', [this.kind, this.target]);
    if (row) {
      const state = this.restore(row.state);
      checkRootIdentity(state, this.target);
      this.loaded = {exists: true, version: Number(row.version), state};
    }
  }
  private check() { if (!this.active) throw new Error('Write repository used outside its command transaction'); }
  async get(id: string): Promise<Loaded<S> | undefined> {
    this.check();
    if (id !== this.target) throw new Error('Command transaction cannot access another aggregate');
    return this.pending !== undefined ? {exists: true, version: this.version, state: structuredClone(this.pending)} : structuredClone(this.loaded);
  }
  async save(state: S) {
    this.check();
    checkRootIdentity(state, this.target);
    this.pending = structuredClone(state);
  }
  close() { this.active = false; }
  /** The command transaction flushes staged state only after successful application execution. */
  async flush() {
    if (this.pending === undefined) return this.version;
    const version = this.version + 1;
    if (this.loaded) {
      const result = await this.client.query('UPDATE cafe.aggregates SET version=$3,state=$4 WHERE kind=$1 AND id=$2 AND version=$5',
        [this.kind, this.target, version, JSON.stringify(this.pending), this.version]);
      if (result.rowCount !== 1) throw new Error('Optimistic update lost its owner version');
    } else await this.client.query('INSERT INTO cafe.aggregates(kind,id,version,state) VALUES($1,$2,$3,$4)',
      [this.kind, this.target, version, JSON.stringify(this.pending)]);
    return version;
  }
}
