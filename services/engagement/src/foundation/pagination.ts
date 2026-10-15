import type { Loaded, QueryPort } from './application.js';
import type { PageRequest } from '../contracts/requests/generated/shared.js';
export type { PageRequest } from '../contracts/requests/generated/shared.js';
export type Page<S> = { items: Loaded<S>[]; nextId?: string | undefined };
export interface PagedQueryPort<S> extends QueryPort<S> {
  page(request: PageRequest): Promise<Page<S>>;
}
