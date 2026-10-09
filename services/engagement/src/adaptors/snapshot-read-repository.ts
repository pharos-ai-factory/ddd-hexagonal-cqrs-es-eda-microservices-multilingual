import type {Loaded, QueryPort} from '../foundation/application.js';
import type {Page, PageRequest} from '../foundation/pagination.js';
import {identifier} from '../foundation/domain.js';
import {checkRootIdentity} from './receipts.js';
import {PostgresContextDatabase} from './postgres.js';

/** Reads validated snapshots from the context-owned database. */
export class PostgresSnapshotReadRepository<S> implements QueryPort<S> {
  constructor(private db: PostgresContextDatabase, private kind: string, private restore: (value: unknown) => S) {}
  private loaded(row: {id: string; version: number; state: unknown}): Loaded<S> {
    const state = this.restore(row.state);
    checkRootIdentity(state, row.id);
    return {exists: true, version: Number(row.version), state};
  }
  async get(id: string): Promise<Loaded<S> | undefined> {
    identifier(id);
    const {rows: [row]} = await this.db.pool.query('SELECT id,version,state FROM cafe.aggregates WHERE kind=$1 AND id=$2', [this.kind, id]);
    return row ? this.loaded(row) : undefined;
  }
  async page(request: PageRequest): Promise<Page<S>> {
    if (!Number.isInteger(request.limit) || request.limit < 1 || request.limit > 100) throw new Error('Invalid page size');
    if (request.after) identifier(request.after);
    const {rows} = await this.db.pool.query(
      'SELECT id,version,state FROM cafe.aggregates WHERE kind=$1 AND ($2::uuid IS NULL OR id>$2::uuid) ORDER BY id LIMIT $3',
      [this.kind, request.after ?? null, request.limit+1]);
    const selected = rows.slice(0, request.limit);
    return {items: selected.map(row => this.loaded(row)),
      nextId: rows.length > request.limit ? selected.at(-1)!.id as string : undefined};
  }
  async list(): Promise<Loaded<S>[]> {
    const {rows} = await this.db.pool.query('SELECT id,version,state FROM cafe.aggregates WHERE kind=$1 ORDER BY id', [this.kind]);
    return rows.map(row => this.loaded(row));
  }
}
