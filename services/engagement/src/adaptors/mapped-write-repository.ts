import type {WriteRepository} from '../foundation/write-repository.js';


/** Restore aggregates and extract snapshots through the owner persistence boundary. */
export class MappedWriteRepository<S, A> implements WriteRepository<A> {
  constructor(private source: WriteRepository<S>, private restore: (state: S) => A, private snapshot: (aggregate: A) => S) {}
  async get(id: string) {
    const loaded = await this.source.get(id);
    return loaded ? {...loaded, state: this.restore(loaded.state)} : undefined;
  }
  async save(aggregate: A) { await this.source.save(this.snapshot(aggregate)); }
}
