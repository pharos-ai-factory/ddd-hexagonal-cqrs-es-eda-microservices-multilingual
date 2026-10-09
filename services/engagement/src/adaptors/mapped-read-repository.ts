import type {Loaded} from '../foundation/application.js';
import type {PageRequest, PagedQueryPort} from '../foundation/pagination.js';

/** Translate validated storage values while preserving query revisions and continuations. */
export class MappedReadRepository<S, V> implements PagedQueryPort<V> {
  constructor(private source: PagedQueryPort<S>, private view: (state: S) => V) {}
  private map(value: Loaded<S>): Loaded<V> {
    return {exists: value.exists, version: value.version, state: this.view(value.state)};
  }
  async get(id: string) { const value = await this.source.get(id); return value ? this.map(value) : undefined; }
  async list() { return (await this.source.list()).map(value => this.map(value)); }
  async page(request: PageRequest) {
    const page = await this.source.page(request);
    return {items: page.items.map(value => this.map(value)), nextId: page.nextId};
  }
}
