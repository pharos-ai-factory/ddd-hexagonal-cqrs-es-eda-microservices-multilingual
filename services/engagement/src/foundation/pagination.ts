import type {Loaded, QueryPort} from './application.js';
export type PageRequest = {limit: number; after?: string | undefined};
export type Page<S> = {items: Loaded<S>[]; nextId?: string | undefined};
export interface PagedQueryPort<S> extends QueryPort<S> {
  page(request: PageRequest): Promise<Page<S>>;
}
